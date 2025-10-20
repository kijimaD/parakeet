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
	TotalFiles       int      // 総ファイル数
	ValidFiles       int      // 有効なファイル数
	InvalidFiles     []string // 無効なファイル名のリスト
	DuplicateFiles   []string // 重複するタイムスタンプを持つファイルのリスト
	HasDuplicates    bool     // 重複があるかどうか
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
		InvalidFiles:   []string{},
		DuplicateFiles: []string{},
	}

	// タイムスタンプの出現回数を記録
	timestampMap := make(map[string][]string)

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

	// サマリーを出力
	_, _ = fmt.Fprintf(opts.Writer, "\nValidation Summary:\n")
	_, _ = fmt.Fprintf(opts.Writer, "  Total files: %d\n", result.TotalFiles)
	_, _ = fmt.Fprintf(opts.Writer, "  Valid: %d\n", result.ValidFiles)
	_, _ = fmt.Fprintf(opts.Writer, "  Invalid: %d\n", len(result.InvalidFiles))
	_, _ = fmt.Fprintf(opts.Writer, "  Duplicates: %d\n", len(result.DuplicateFiles))

	if len(result.InvalidFiles) == 0 && !result.HasDuplicates {
		_, _ = fmt.Fprintf(opts.Writer, "\n✓ All files are properly formatted!\n")
	} else {
		if len(result.InvalidFiles) > 0 {
			_, _ = fmt.Fprintf(opts.Writer, "\n✗ Some files have invalid format.\n")
		}
		if result.HasDuplicates {
			_, _ = fmt.Fprintf(opts.Writer, "\n⚠ Some files have duplicate timestamps.\n")
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
