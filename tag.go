package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

// TagOptions はタグ編集操作のオプションを表す
type TagOptions struct {
	Interactive bool      // インタラクティブモード（survey を使用）
	Writer      io.Writer // 出力先
}

// EditTags はファイルのタグをインタラクティブに編集する
// インタラクティブモードでは、既存のタグを選択・解除し、新しいタグを追加できる
func EditTags(filePath string, opts TagOptions) error {
	// ファイルの存在チェック
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", filePath)
		}
		return fmt.Errorf("failed to access file: %w", err)
	}

	// ディレクトリは処理できない
	if fileInfo.IsDir() {
		return fmt.Errorf("cannot edit tags for directory: %s", filePath)
	}

	fileName := filepath.Base(filePath)
	dirPath := filepath.Dir(filePath)

	// ファイル名が正しいフォーマットかチェック
	components, err := ParseFileName(fileName)
	if err != nil {
		return fmt.Errorf("file name is not in correct format: %w", err)
	}

	// インタラクティブモードでタグを編集
	if opts.Interactive {
		newTags, err := promptForTags(components.Tags)
		if err != nil {
			return fmt.Errorf("failed to get tags: %w", err)
		}

		// タグが変更されたかチェック
		if !tagsEqual(components.Tags, newTags) {
			// 新しいファイル名を生成
			components.Tags = newTags
			newFileName := components.FormatFileName()
			newFilePath := filepath.Join(dirPath, newFileName)

			// ファイルをリネーム
			if err := os.Rename(filePath, newFilePath); err != nil {
				return fmt.Errorf("failed to rename file: %w", err)
			}

			fmt.Fprintf(opts.Writer, "✓ Renamed: %s → %s\n", fileName, newFileName)
		} else {
			fmt.Fprintln(opts.Writer, "✓ No changes made")
		}
	}

	return nil
}

// promptForTags はインタラクティブにタグを選択・編集する
func promptForTags(currentTags []string) ([]string, error) {
	// 既存のタグをすべて選択状態にする
	var selectedTags []string
	if len(currentTags) > 0 {
		selectedTags = make([]string, len(currentTags))
		copy(selectedTags, currentTags)
	}

	// よく使われるタグの候補リスト
	commonTags := []string{
		"work",
		"personal",
		"important",
		"urgent",
		"draft",
		"final",
		"review",
		"archive",
		"temp",
		"backup",
		"network",
		"infra",
		"dev",
		"prod",
		"test",
	}

	// 既存のタグと候補を統合（重複を除く）
	tagOptions := make([]string, 0)
	seenTags := make(map[string]bool)

	// 既存のタグを優先的に追加
	for _, tag := range currentTags {
		if !seenTags[tag] {
			tagOptions = append(tagOptions, tag)
			seenTags[tag] = true
		}
	}

	// 候補タグを追加
	for _, tag := range commonTags {
		if !seenTags[tag] {
			tagOptions = append(tagOptions, tag)
			seenTags[tag] = true
		}
	}

	// カスタムタグ追加オプション
	tagOptions = append(tagOptions, "[+ Add custom tag]")

	for {
		// タグ選択プロンプト
		prompt := &survey.MultiSelect{
			Message: "Select tags (space to toggle, enter to confirm):",
			Options: tagOptions,
			Default: selectedTags,
		}

		var selected []string
		err := survey.AskOne(prompt, &selected)
		if err != nil {
			return nil, err
		}

		// カスタムタグ追加が選択されたかチェック
		addCustom := false
		finalTags := make([]string, 0)
		for _, tag := range selected {
			if tag == "[+ Add custom tag]" {
				addCustom = true
			} else {
				finalTags = append(finalTags, tag)
			}
		}

		// カスタムタグを追加
		if addCustom {
			customTag, err := promptForCustomTag()
			if err != nil {
				return nil, err
			}
			if customTag != "" {
				// カスタムタグを追加
				finalTags = append(finalTags, customTag)
				// オプションリストに追加
				tagOptions = append([]string{customTag}, tagOptions...)
				// 再度選択
				selectedTags = finalTags
				continue
			}
		}

		// タグをソート
		sort.Strings(finalTags)

		return finalTags, nil
	}
}

// promptForCustomTag はカスタムタグの入力を求める
func promptForCustomTag() (string, error) {
	prompt := &survey.Input{
		Message: "Enter custom tag:",
	}

	var tag string
	err := survey.AskOne(prompt, &tag, survey.WithValidator(survey.Required))
	if err != nil {
		return "", err
	}

	// タグをトリム
	tag = strings.TrimSpace(tag)

	// タグに不正な文字が含まれていないかチェック
	if strings.ContainsAny(tag, "/_--.") {
		return "", fmt.Errorf("tag cannot contain special characters: /, _, -, .")
	}

	return tag, nil
}

// tagsEqual は2つのタグスライスが等しいかチェックする
func tagsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// ソートしてから比較
	aCopy := make([]string, len(a))
	bCopy := make([]string, len(b))
	copy(aCopy, a)
	copy(bCopy, b)

	sort.Strings(aCopy)
	sort.Strings(bCopy)

	for i := range aCopy {
		if aCopy[i] != bCopy[i] {
			return false
		}
	}

	return true
}

// ShowTags は指定されたファイルの現在のタグを表示する
func ShowTags(filePath string, w io.Writer) error {
	// ファイルの存在チェック
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", filePath)
		}
		return fmt.Errorf("failed to access file: %w", err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("cannot show tags for directory: %s", filePath)
	}

	fileName := filepath.Base(filePath)

	// ファイル名をパース
	components, err := ParseFileName(fileName)
	if err != nil {
		return fmt.Errorf("file name is not in correct format: %w", err)
	}

	// タグを表示
	fmt.Fprintf(w, "File: %s\n", fileName)
	fmt.Fprintf(w, "Timestamp: %s\n", components.Timestamp)
	fmt.Fprintf(w, "Comment: %s\n", components.Comment)

	if len(components.Tags) > 0 {
		fmt.Fprintf(w, "Tags: %s\n", strings.Join(components.Tags, ", "))
	} else {
		fmt.Fprintln(w, "Tags: (none)")
	}

	return nil
}

// SetTags はファイルのタグを直接設定する（非インタラクティブ）
func SetTags(filePath string, tags []string, w io.Writer) error {
	// ファイルの存在チェック
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", filePath)
		}
		return fmt.Errorf("failed to access file: %w", err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("cannot set tags for directory: %s", filePath)
	}

	fileName := filepath.Base(filePath)
	dirPath := filepath.Dir(filePath)

	// ファイル名をパース
	components, err := ParseFileName(fileName)
	if err != nil {
		return fmt.Errorf("file name is not in correct format: %w", err)
	}

	// タグをソート
	sort.Strings(tags)

	// タグが変更されたかチェック
	if !tagsEqual(components.Tags, tags) {
		// 新しいファイル名を生成
		components.Tags = tags
		newFileName := components.FormatFileName()
		newFilePath := filepath.Join(dirPath, newFileName)

		// ファイルをリネーム
		if err := os.Rename(filePath, newFilePath); err != nil {
			return fmt.Errorf("failed to rename file: %w", err)
		}

		fmt.Fprintf(w, "✓ Renamed: %s → %s\n", fileName, newFileName)
	} else {
		fmt.Fprintln(w, "✓ No changes made")
	}

	return nil
}
