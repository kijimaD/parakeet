package main

import (
	"fmt"
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
