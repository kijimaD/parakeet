package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// MarkdownOptions はMarkdown出力操作のオプションを表す
type MarkdownOptions struct {
	Writer     io.Writer // 出力先
	Extensions []string  // 対象拡張子（空の場合は全ファイル）
}

// GenerateMarkdownTable はディレクトリ内のファイル一覧をMarkdown表形式で出力する
func GenerateMarkdownTable(targetDir string, opts MarkdownOptions) error {
	// ディレクトリの存在チェック
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", targetDir)
	}

	// ディレクトリを読み込む
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// ヘッダーを出力
	_, _ = fmt.Fprintln(opts.Writer, "| ID | Title | Tags |")
	_, _ = fmt.Fprintln(opts.Writer, "|---|---|---|")

	// ファイルを処理
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

		// フォーマット済みファイルのみ処理
		components, err := ParseFileName(fileName)
		if err != nil {
			// フォーマット外のファイルはスキップ
			continue
		}

		// タグを結合
		tagsStr := ""
		if len(components.Tags) > 0 {
			tagsStr = strings.Join(components.Tags, ", ")
		}

		// Markdown行を出力
		_, _ = fmt.Fprintf(opts.Writer, "| %s | %s | %s |\n",
			components.Timestamp,
			components.Comment,
			tagsStr,
		)
	}

	return nil
}
