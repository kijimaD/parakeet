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
				Name:      "tag",
				Usage:     "ファイルのタグをインタラクティブに編集する",
				ArgsUsage: "<file_path>",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "show",
						Aliases: []string{"s"},
						Usage:   "現在のタグを表示する",
					},
				},
				Action: func(_ context.Context, cmd *cli.Command) error {
					// ファイルパスを取得
					if cmd.Args().Len() == 0 {
						return fmt.Errorf("file path is required")
					}
					filePath := cmd.Args().Get(0)

					// --show フラグの場合はタグを表示
					if cmd.Bool("show") {
						return ShowTags(filePath, os.Stdout)
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
