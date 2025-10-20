package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileNameComponents はフォーマット済みファイル名の構成要素を表す
type FileNameComponents struct {
	Timestamp string   // タイムスタンプ（ISO8601形式: 20250903T083109）
	Comment   string   // 人間が読めるコメント
	Tags      []string // タグのリスト
	Extension string   // 拡張子
}

// GenerateTimestamp は現在時刻からタイムスタンプを生成する
// フォーマット: YYYYMMDDTHHMMSS
func GenerateTimestamp() string {
	return time.Now().Format("20060102T150405")
}

// GenerateUniqueTimestamp は既存のタイムスタンプと重複しないタイムスタンプを生成する
// existingTimestamps に既存のタイムスタンプのリストを渡す
func GenerateUniqueTimestamp(existingTimestamps map[string]bool) string {
	timestamp := GenerateTimestamp()

	// 重複がなければそのまま返す
	if !existingTimestamps[timestamp] {
		return timestamp
	}

	// 重複する場合は、タイムスタンプをパースして1秒ずつ進める
	t, err := time.Parse("20060102T150405", timestamp)
	if err != nil {
		// パースに失敗した場合は現在時刻を返す
		return timestamp
	}

	// 重複しないタイムスタンプが見つかるまで1秒ずつ進める
	for {
		t = t.Add(1 * time.Second)
		newTimestamp := t.Format("20060102T150405")
		if !existingTimestamps[newTimestamp] {
			return newTimestamp
		}
	}
}

// CollectExistingTimestamps はディレクトリ内のフォーマット済みファイルからタイムスタンプを収集する
func CollectExistingTimestamps(dirPath string) (map[string]bool, error) {
	timestamps := make(map[string]bool)

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// フォーマット済みファイルからタイムスタンプを抽出
		if components, err := ParseFileName(entry.Name()); err == nil {
			timestamps[components.Timestamp] = true
		}
	}

	return timestamps, nil
}

// FindFileByID はディレクトリ内からIDに一致するファイルを検索する
// 複数のファイルが見つかった場合はエラーを返す
func FindFileByID(dirPath, id string) (string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	var matchedFiles []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// フォーマット済みファイルからタイムスタンプを抽出
		if components, err := ParseFileName(entry.Name()); err == nil {
			if components.Timestamp == id {
				matchedFiles = append(matchedFiles, filepath.Join(dirPath, entry.Name()))
			}
		}
	}

	if len(matchedFiles) == 0 {
		return "", fmt.Errorf("no file found with ID: %s", id)
	}

	if len(matchedFiles) > 1 {
		return "", fmt.Errorf("multiple files found with ID %s:\n%s", id, strings.Join(matchedFiles, "\n"))
	}

	return matchedFiles[0], nil
}

// FormatFileName は構成要素からフォーマット済みファイル名を生成する
// フォーマット: {timestamp}--{comment}__{tag1}_{tag2}.{extension}
func (c FileNameComponents) FormatFileName() string {
	var parts []string

	// タイムスタンプとコメント部分
	parts = append(parts, fmt.Sprintf("%s--%s", c.Timestamp, c.Comment))

	// タグ部分（存在する場合）
	if len(c.Tags) > 0 {
		parts = append(parts, strings.Join(c.Tags, "_"))
	}

	// ダブルアンダースコアで結合
	baseName := strings.Join(parts, "__")

	// 拡張子を追加
	if c.Extension != "" {
		return fmt.Sprintf("%s.%s", baseName, c.Extension)
	}

	return baseName
}

// ParseFileName はフォーマット済みファイル名を構成要素にパースする
func ParseFileName(filename string) (*FileNameComponents, error) {
	// 拡張子を削除
	ext := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, ext)
	if ext != "" {
		ext = ext[1:] // 先頭のドットを削除
	}

	// ダブルアンダースコアで分割
	parts := strings.Split(baseName, "__")
	if len(parts) < 1 {
		return nil, fmt.Errorf("invalid filename format: %s", filename)
	}

	// タイムスタンプとコメントをパース（最初の部分）
	timestampCommentParts := strings.SplitN(parts[0], "--", 2)
	if len(timestampCommentParts) != 2 {
		return nil, fmt.Errorf("invalid timestamp-comment format: %s", parts[0])
	}

	components := &FileNameComponents{
		Timestamp: timestampCommentParts[0],
		Comment:   timestampCommentParts[1],
		Extension: ext,
	}

	// タグをパース（残りの部分）
	if len(parts) > 1 {
		for i := 1; i < len(parts); i++ {
			tags := strings.Split(parts[i], "_")
			components.Tags = append(components.Tags, tags...)
		}
	}

	return components, nil
}

// IsFormatted はファイル名が正しいフォーマットかどうかをチェックする
func IsFormatted(filename string) bool {
	_, err := ParseFileName(filename)
	return err == nil
}

// MatchesExtensions はファイル名が指定された拡張子のいずれかに一致するかチェックする
// extensions が空の場合は常に true を返す
func MatchesExtensions(filename string, extensions []string) bool {
	// 拡張子指定がない場合はすべて対象
	if len(extensions) == 0 {
		return true
	}

	ext := filepath.Ext(filename)
	if ext != "" {
		ext = ext[1:] // 先頭のドットを削除
	}

	// 拡張子が一致するかチェック
	for _, targetExt := range extensions {
		if strings.EqualFold(ext, targetExt) {
			return true
		}
	}

	return false
}
