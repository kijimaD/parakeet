package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// RenameOptions はリネーム操作のオプションを表す
type RenameOptions struct {
	DryRun  bool      // ドライランモード（実際にはリネームしない）
	Verbose bool      // 詳細出力モード
	Writer  io.Writer // 出力先
}

// GenerateFileNames はディレクトリ内のすべてのファイルにフォーマット済みファイル名を生成する
func GenerateFileNames(targetDir string, opts RenameOptions) error {
	// ディレクトリの存在チェック
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", targetDir)
	}

	// ディレクトリを読み込む
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	timestamp := GenerateTimestamp()
	processedCount := 0
	skippedCount := 0

	for _, entry := range entries {
		// ディレクトリはスキップ
		if entry.IsDir() {
			continue
		}

		oldName := entry.Name()
		oldPath := filepath.Join(targetDir, oldName)

		// すでにフォーマット済みの場合はスキップ
		if IsFormatted(oldName) {
			if opts.Verbose {
				fmt.Fprintf(opts.Writer, "Skipped (already formatted): %s\n", oldName)
			}
			skippedCount++
			continue
		}

		// 現在のファイル名からコメントとタグを抽出
		ext := filepath.Ext(oldName)
		baseName := strings.TrimSuffix(oldName, ext)
		if ext != "" {
			ext = ext[1:] // 先頭のドットを削除
		}

		// タイムスタンプ付きの新しいファイル名を作成
		components := FileNameComponents{
			Timestamp: timestamp,
			Comment:   baseName,
			Tags:      []string{}, // デフォルトではタグなし
			Extension: ext,
		}

		newName := components.FormatFileName()
		newPath := filepath.Join(targetDir, newName)

		// 新しいファイル名がすでに存在するかチェック
		if _, err := os.Stat(newPath); err == nil {
			fmt.Fprintf(opts.Writer, "Warning: target file already exists, skipping: %s\n", newName)
			skippedCount++
			continue
		}

		if opts.DryRun {
			fmt.Fprintf(opts.Writer, "[DRY RUN] Would rename: %s -> %s\n", oldName, newName)
		} else {
			if err := os.Rename(oldPath, newPath); err != nil {
				fmt.Fprintf(opts.Writer, "Error renaming %s: %v\n", oldName, err)
				continue
			}
			if opts.Verbose {
				fmt.Fprintf(opts.Writer, "Renamed: %s -> %s\n", oldName, newName)
			}
		}

		processedCount++
	}

	// サマリーを出力
	fmt.Fprintf(opts.Writer, "\nSummary:\n")
	fmt.Fprintf(opts.Writer, "  Processed: %d\n", processedCount)
	fmt.Fprintf(opts.Writer, "  Skipped: %d\n", skippedCount)
	if opts.DryRun {
		fmt.Fprintf(opts.Writer, "  (Dry run - no files were actually renamed)\n")
	}

	return nil
}
