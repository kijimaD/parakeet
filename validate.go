package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ValidateOptions はバリデーション操作のオプションを表す
type ValidateOptions struct {
	Writer     io.Writer // 出力先
	Extensions []string  // 対象拡張子（空の場合は全ファイル）
}

// ValidateResult はバリデーション結果を表す
type ValidateResult struct {
	TotalFiles        int                 // 総ファイル数
	ValidFiles        int                 // 有効なファイル数
	InvalidFiles      []string            // 無効なファイル名のリスト
	DuplicateFiles    []string            // 重複するタイムスタンプを持つファイルのリスト
	HasDuplicates     bool                // 重複があるかどうか
	UndefinedTagFiles map[string][]string // 未定義タグを持つファイル: ファイル名 -> 未定義タグリスト
	HasUndefinedTags  bool                // 未定義タグがあるかどうか
}

// ValidateFileNames はディレクトリ内のファイル名をバリデーションする
func ValidateFileNames(targetDir string, opts ValidateOptions) (*ValidateResult, error) {
	// ディレクトリの存在チェック
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", targetDir)
	}

	// ディレクトリを読み込む
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	result := &ValidateResult{
		InvalidFiles:      []string{},
		DuplicateFiles:    []string{},
		UndefinedTagFiles: make(map[string][]string),
	}

	// タイムスタンプの出現回数を記録
	timestampMap := make(map[string][]string)

	// tag.tomlを読み込む（targetDir内に存在する場合）
	tomlPath := filepath.Join(targetDir, "tag.toml")
	tagDefs, err := LoadTagsFromTOML(tomlPath)
	if err != nil {
		// エラーがあっても続行（tag.tomlがない場合はタグチェックをスキップ）
		tagDefs = []TagDefinition{}
	}

	// 定義済みタグのセットを作成
	validTags := make(map[string]bool)
	for _, tagDef := range tagDefs {
		validTags[tagDef.Key] = true
	}
	hasTagDefinitions := len(validTags) > 0

	for _, entry := range entries {
		// ディレクトリはスキップ
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()

		// 拡張子フィルタリング
		if !MatchesExtensions(fileName, opts.Extensions) {
			continue
		}

		result.TotalFiles++

		// ファイル名が正しいフォーマットかチェック
		if IsFormatted(fileName) {
			result.ValidFiles++

			// タイムスタンプを抽出して重複チェック
			if components, err := ParseFileName(fileName); err == nil {
				timestampMap[components.Timestamp] = append(timestampMap[components.Timestamp], fileName)

				// タグの定義チェック（tag.tomlが存在する場合のみ）
				if hasTagDefinitions && len(components.Tags) > 0 {
					var undefinedTags []string
					for _, tag := range components.Tags {
						if !validTags[tag] {
							undefinedTags = append(undefinedTags, tag)
						}
					}
					if len(undefinedTags) > 0 {
						result.HasUndefinedTags = true
						result.UndefinedTagFiles[fileName] = undefinedTags
					}
				}
			}
		} else {
			result.InvalidFiles = append(result.InvalidFiles, fileName)
			_, _ = fmt.Fprintf(opts.Writer, "✗ %s (invalid format)\n", fileName)
		}
	}

	// 重複チェック
	for timestamp, files := range timestampMap {
		if len(files) > 1 {
			result.HasDuplicates = true
			for _, file := range files {
				result.DuplicateFiles = append(result.DuplicateFiles, file)
				_, _ = fmt.Fprintf(opts.Writer, "⚠ %s (duplicate timestamp: %s)\n", file, timestamp)
			}
		}
	}

	// 未定義タグの出力
	if result.HasUndefinedTags {
		for fileName, tags := range result.UndefinedTagFiles {
			_, _ = fmt.Fprintf(opts.Writer, "⚠ %s (undefined tags: %v)\n", fileName, tags)
		}
	}

	// サマリーを出力
	_, _ = fmt.Fprintf(opts.Writer, "\nValidation Summary:\n")
	_, _ = fmt.Fprintf(opts.Writer, "  Total files: %d\n", result.TotalFiles)
	_, _ = fmt.Fprintf(opts.Writer, "  Valid: %d\n", result.ValidFiles)
	_, _ = fmt.Fprintf(opts.Writer, "  Invalid: %d\n", len(result.InvalidFiles))
	_, _ = fmt.Fprintf(opts.Writer, "  Duplicates: %d\n", len(result.DuplicateFiles))
	_, _ = fmt.Fprintf(opts.Writer, "  Undefined tags: %d\n", len(result.UndefinedTagFiles))

	if len(result.InvalidFiles) == 0 && !result.HasDuplicates && !result.HasUndefinedTags {
		_, _ = fmt.Fprintf(opts.Writer, "\n✓ All files are properly formatted!\n")
	} else {
		if len(result.InvalidFiles) > 0 {
			_, _ = fmt.Fprintf(opts.Writer, "\n✗ Some files have invalid format.\n")
		}
		if result.HasDuplicates {
			_, _ = fmt.Fprintf(opts.Writer, "\n⚠ Some files have duplicate timestamps.\n")
		}
		if result.HasUndefinedTags {
			_, _ = fmt.Fprintf(opts.Writer, "\n⚠ Some files have undefined tags.\n")
		}
	}

	return result, nil
}

// ValidateFileName は単一のファイル名をバリデーションする
func ValidateFileName(filename string) error {
	components, err := ParseFileName(filename)
	if err != nil {
		return err
	}

	// タイムスタンプの形式チェック（YYYYMMDDTHHMMSS）
	if len(components.Timestamp) != 15 {
		return fmt.Errorf("invalid timestamp length: expected 15, got %d", len(components.Timestamp))
	}

	// コメントが空でないかチェック
	if components.Comment == "" {
		return fmt.Errorf("comment cannot be empty")
	}

	return nil
}

// GetInvalidFiles はディレクトリ内の無効なファイル名のリストを返す
func GetInvalidFiles(targetDir string) ([]string, error) {
	// ディレクトリの存在チェック
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", targetDir)
	}

	// ディレクトリを読み込む
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var invalidFiles []string

	for _, entry := range entries {
		// ディレクトリはスキップ
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()

		// ファイル名が正しいフォーマットかチェック
		if !IsFormatted(fileName) {
			invalidFiles = append(invalidFiles, filepath.Join(targetDir, fileName))
		}
	}

	return invalidFiles, nil
}
