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
					&cli.BoolFlag{
						Name:    "dry-run",
						Aliases: []string{"n"},
						Usage:   "実際にリネームせず、リネーム予定を表示する",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "詳細な出力を表示する",
					},
					&cli.StringSliceFlag{
						Name:    "ext",
						Aliases: []string{"e"},
						Usage:   "対象拡張子（カンマ区切り、例: pdf,txt,md）",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					// 対象ディレクトリを取得（デフォルトはカレントディレクトリ）
					targetDir := "."
					if cmd.Args().Len() > 0 {
						targetDir = cmd.Args().Get(0)
					}

					opts := RenameOptions{
						DryRun:     cmd.Bool("dry-run"),
						Verbose:    cmd.Bool("verbose"),
						Writer:     os.Stdout,
						Extensions: cmd.StringSlice("ext"),
					}

					return GenerateFileNames(targetDir, opts)
				},
			},
			{
				Name:  "validate",
				Usage: "ディレクトリ内のファイル名が正しいフォーマットかをチェックする",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "詳細な出力を表示する",
					},
					&cli.StringSliceFlag{
						Name:    "ext",
						Aliases: []string{"e"},
						Usage:   "対象拡張子（カンマ区切り、例: pdf,txt,md）",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					// 対象ディレクトリを取得（デフォルトはカレントディレクトリ）
					targetDir := "."
					if cmd.Args().Len() > 0 {
						targetDir = cmd.Args().Get(0)
					}

					opts := ValidateOptions{
						Verbose:    cmd.Bool("verbose"),
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
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		log.Fatal(err)
	}
}
