// Package main はアプリケーションのエントリーポイント
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "parakeet",
		Usage: "タイムスタンプベースのフォーマットでファイル名を管理するツール",
		Commands: []*cli.Command{
			{
				Name:  "generate",
				Usage: "ディレクトリ内のファイルにタイムスタンプ付きのフォーマット済みファイル名を生成する",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:    "ext",
						Aliases: []string{"e"},
						Usage:   "対象拡張子（カンマ区切り、例: pdf,txt,md）",
					},
				},
				Action: func(_ context.Context, cmd *cli.Command) error {
					// 拡張子指定は必須
					extensions := cmd.StringSlice("ext")
					if len(extensions) == 0 {
						return fmt.Errorf("--ext flag is required: specify at least one file extension (e.g., --ext pdf --ext txt)")
					}

					// 対象ディレクトリを取得（デフォルトはカレントディレクトリ）
					targetDir := "."
					if cmd.Args().Len() > 0 {
						targetDir = cmd.Args().Get(0)
					}

					opts := RenameOptions{
						Writer:     os.Stdout,
						Extensions: extensions,
					}

					return GenerateFileNames(targetDir, opts)
				},
			},
			{
				Name:  "validate",
				Usage: "ディレクトリ内のファイル名が正しいフォーマットかをチェックする",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:    "ext",
						Aliases: []string{"e"},
						Usage:   "対象拡張子（カンマ区切り、例: pdf,txt,md）",
					},
				},
				Action: func(_ context.Context, cmd *cli.Command) error {
					// 対象ディレクトリを取得（デフォルトはカレントディレクトリ）
					targetDir := "."
					if cmd.Args().Len() > 0 {
						targetDir = cmd.Args().Get(0)
					}

					opts := ValidateOptions{
						Writer:     os.Stdout,
						Extensions: cmd.StringSlice("ext"),
					}

					result, err := ValidateFileNames(targetDir, opts)
					if err != nil {
						return err
					}

					// 無効なファイルがある場合は終了コード1を返す
					if len(result.InvalidFiles) > 0 {
						os.Exit(1)
					}

					return nil
				},
			},
			{
				Name:  "md",
				Usage: "ディレクトリ内のファイル一覧をMarkdown表形式で出力する",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:    "ext",
						Aliases: []string{"e"},
						Usage:   "対象拡張子（カンマ区切り、例: pdf,txt,md）",
					},
				},
				Action: func(_ context.Context, cmd *cli.Command) error {
					// 対象ディレクトリを取得（デフォルトはカレントディレクトリ）
					targetDir := "."
					if cmd.Args().Len() > 0 {
						targetDir = cmd.Args().Get(0)
					}

					opts := MarkdownOptions{
						Writer:     os.Stdout,
						Extensions: cmd.StringSlice("ext"),
					}

					return GenerateMarkdownTable(targetDir, opts)
				},
			},
			{
				Name:      "tag",
				Usage:     "ファイルのタグをインタラクティブに編集する",
				ArgsUsage: "<id>",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "show",
						Aliases: []string{"s"},
						Usage:   "現在のタグを表示する",
					},
					&cli.StringSliceFlag{
						Name:    "set",
						Aliases: []string{"t"},
						Usage:   "タグを直接指定する（カンマ区切り、例: --set tag1 --set tag2）",
					},
				},
				Action: func(_ context.Context, cmd *cli.Command) error {
					// IDを取得
					if cmd.Args().Len() == 0 {
						return fmt.Errorf("ID is required")
					}
					id := cmd.Args().Get(0)

					// IDでファイルを検索
					filePath, err := FindFileByID(".", id)
					if err != nil {
						return fmt.Errorf("file not found: %w", err)
					}

					// --show フラグの場合はタグを表示
					if cmd.Bool("show") {
						return ShowTags(filePath, os.Stdout)
					}

					// --set フラグが指定された場合は非インタラクティブモード
					if setTags := cmd.StringSlice("set"); len(setTags) > 0 {
						// tag.tomlに対してバリデーション
						if err := ValidateTags(setTags, "tag.toml"); err != nil {
							return err
						}

						// タグを設定
						return SetTags(filePath, setTags, os.Stdout)
					}

					// デフォルトはインタラクティブモード
					opts := TagOptions{
						Interactive: true,
						Writer:      os.Stdout,
					}

					return EditTags(filePath, opts)
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		log.Fatal(err)
	}
}
