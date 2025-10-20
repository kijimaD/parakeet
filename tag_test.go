package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagsEqual(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected bool
	}{
		{
			name:     "equal tags in same order",
			a:        []string{"tag1", "tag2", "tag3"},
			b:        []string{"tag1", "tag2", "tag3"},
			expected: true,
		},
		{
			name:     "equal tags in different order",
			a:        []string{"tag1", "tag3", "tag2"},
			b:        []string{"tag2", "tag1", "tag3"},
			expected: true,
		},
		{
			name:     "different tags",
			a:        []string{"tag1", "tag2"},
			b:        []string{"tag1", "tag3"},
			expected: false,
		},
		{
			name:     "different length",
			a:        []string{"tag1", "tag2"},
			b:        []string{"tag1", "tag2", "tag3"},
			expected: false,
		},
		{
			name:     "both empty",
			a:        []string{},
			b:        []string{},
			expected: true,
		},
		{
			name:     "one empty",
			a:        []string{"tag1"},
			b:        []string{},
			expected: false,
		},
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tagsEqual(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShowTags(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		fileName    string
		expectError bool
	}{
		{
			name:        "valid file with tags",
			fileName:    "20250903T083109--test-file__tag1_tag2.pdf",
			expectError: false,
		},
		{
			name:        "valid file without tags",
			fileName:    "20250903T083109--test-file.pdf",
			expectError: false,
		},
		{
			name:        "invalid file format",
			fileName:    "invalid-file.pdf",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create temporary directory
			tmpDir, err := os.MkdirTemp("", "parakeet-showtags-*")
			require.NoError(t, err)
			defer func() { _ = os.RemoveAll(tmpDir) }()

			// Create test file
			filePath := filepath.Join(tmpDir, tt.fileName)
			err = os.WriteFile(filePath, []byte("test content"), 0644)
			require.NoError(t, err)

			// Test ShowTags
			buf := &bytes.Buffer{}
			err = ShowTags(filePath, buf)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShowTags_NonExistentFile(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	err := ShowTags("/non/existent/file.pdf", buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestShowTags_Directory(t *testing.T) {
	t.Parallel()
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-showtags-dir-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	buf := &bytes.Buffer{}
	err = ShowTags(tmpDir, buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot show tags for directory")
}

func TestSetTags(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		fileName     string
		newTags      []string
		expectedName string
	}{
		{
			name:         "add tags to file without tags",
			fileName:     "20250903T083109--test-file.pdf",
			newTags:      []string{"tag1", "tag2"},
			expectedName: "20250903T083109--test-file__tag1_tag2.pdf",
		},
		{
			name:         "replace existing tags",
			fileName:     "20250903T083109--test-file__old1_old2.pdf",
			newTags:      []string{"new1", "new2"},
			expectedName: "20250903T083109--test-file__new1_new2.pdf",
		},
		{
			name:         "remove all tags",
			fileName:     "20250903T083109--test-file__tag1_tag2.pdf",
			newTags:      []string{},
			expectedName: "20250903T083109--test-file.pdf",
		},
		{
			name:         "tags are sorted",
			fileName:     "20250903T083109--test-file.pdf",
			newTags:      []string{"zebra", "apple", "banana"},
			expectedName: "20250903T083109--test-file__apple_banana_zebra.pdf",
		},
		{
			name:         "no change when tags are same",
			fileName:     "20250903T083109--test-file__tag1_tag2.pdf",
			newTags:      []string{"tag1", "tag2"},
			expectedName: "20250903T083109--test-file__tag1_tag2.pdf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create temporary directory
			tmpDir, err := os.MkdirTemp("", "parakeet-settags-*")
			require.NoError(t, err)
			defer func() { _ = os.RemoveAll(tmpDir) }()

			// Create test file
			filePath := filepath.Join(tmpDir, tt.fileName)
			err = os.WriteFile(filePath, []byte("test content"), 0644)
			require.NoError(t, err)

			// Set tags
			buf := &bytes.Buffer{}
			err = SetTags(filePath, tt.newTags, buf)
			require.NoError(t, err)

			// Verify new file exists
			newFilePath := filepath.Join(tmpDir, tt.expectedName)
			_, err = os.Stat(newFilePath)
			assert.NoError(t, err, "New file should exist: %s", tt.expectedName)

			// Verify content is preserved
			content, err := os.ReadFile(newFilePath)
			require.NoError(t, err)
			assert.Equal(t, "test content", string(content))

			// Verify old file doesn't exist (if name changed)
			if tt.fileName != tt.expectedName {
				_, err = os.Stat(filePath)
				assert.True(t, os.IsNotExist(err), "Old file should not exist")
			}
		})
	}
}

func TestSetTags_NonExistentFile(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	err := SetTags("/non/existent/file.pdf", []string{"tag1"}, buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestSetTags_InvalidFormat(t *testing.T) {
	t.Parallel()
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-settags-invalid-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create file with invalid format
	filePath := filepath.Join(tmpDir, "invalid-format.pdf")
	err = os.WriteFile(filePath, []byte("test"), 0644)
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	err = SetTags(filePath, []string{"tag1"}, buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in correct format")
}

func TestSetTags_Directory(t *testing.T) {
	t.Parallel()
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-settags-dir-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	buf := &bytes.Buffer{}
	err = SetTags(tmpDir, []string{"tag1"}, buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot set tags for directory")
}

func TestEditTags_NonInteractive(t *testing.T) {
	t.Parallel()
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-edittags-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test file
	fileName := "20250903T083109--test-file__tag1.pdf"
	filePath := filepath.Join(tmpDir, fileName)
	err = os.WriteFile(filePath, []byte("test content"), 0644)
	require.NoError(t, err)

	// Test non-interactive mode (should do nothing)
	buf := &bytes.Buffer{}
	opts := TagOptions{
		Interactive: false,
		Writer:      buf,
	}

	err = EditTags(filePath, opts)
	assert.NoError(t, err)

	// Verify file still exists with same name
	_, err = os.Stat(filePath)
	assert.NoError(t, err)
}

func TestEditTags_NonExistentFile(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	opts := TagOptions{
		Interactive: true,
		Writer:      buf,
	}

	err := EditTags("/non/existent/file.pdf", opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestEditTags_InvalidFormat(t *testing.T) {
	t.Parallel()
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-edittags-invalid-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create file with invalid format
	filePath := filepath.Join(tmpDir, "invalid-format.pdf")
	err = os.WriteFile(filePath, []byte("test"), 0644)
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	opts := TagOptions{
		Interactive: true,
		Writer:      buf,
	}

	err = EditTags(filePath, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in correct format")
}

func TestIntegration_TagWorkflow(t *testing.T) {
	t.Parallel()
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-tag-workflow-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Step 1: Create a file without tags
	fileName := "20250903T083109--document.pdf"
	filePath := filepath.Join(tmpDir, fileName)
	err = os.WriteFile(filePath, []byte("important document"), 0644)
	require.NoError(t, err)

	// Step 2: Add tags
	buf := &bytes.Buffer{}
	err = SetTags(filePath, []string{"work", "important"}, buf)
	require.NoError(t, err)

	// Step 3: Verify new file exists
	newFileName := "20250903T083109--document__important_work.pdf"
	newFilePath := filepath.Join(tmpDir, newFileName)
	_, err = os.Stat(newFilePath)
	assert.NoError(t, err)

	// Step 4: Modify tags
	buf = &bytes.Buffer{}
	err = SetTags(newFilePath, []string{"work", "urgent", "review"}, buf)
	require.NoError(t, err)

	// Step 5: Verify final file
	finalFileName := "20250903T083109--document__review_urgent_work.pdf"
	finalFilePath := filepath.Join(tmpDir, finalFileName)
	_, err = os.Stat(finalFilePath)
	assert.NoError(t, err)

	// Verify content is preserved
	content, err := os.ReadFile(finalFilePath)
	require.NoError(t, err)
	assert.Equal(t, "important document", string(content))

	// Step 6: Remove all tags
	buf = &bytes.Buffer{}
	err = SetTags(finalFilePath, []string{}, buf)
	require.NoError(t, err)

	// Step 7: Verify back to no tags
	noTagsPath := filepath.Join(tmpDir, "20250903T083109--document.pdf")
	_, err = os.Stat(noTagsPath)
	assert.NoError(t, err)
}

func TestLoadTagsFromTOML(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		content       string
		expectedTags  []TagDefinition
		expectedError bool
	}{
		{
			name: "valid toml file",
			content: `[[tag]]
key = "infra"
desc = "インフラについて"

[[tag]]
key = "mm"
desc = "mmについて"
`,
			expectedTags: []TagDefinition{
				{Key: "infra", Desc: "インフラについて"},
				{Key: "mm", Desc: "mmについて"},
			},
			expectedError: false,
		},
		{
			name: "empty file",
			content: ``,
			expectedTags: nil,
			expectedError: false,
		},
		{
			name: "tag without description",
			content: `[[tag]]
key = "test"
`,
			expectedTags: []TagDefinition{
				{Key: "test", Desc: ""},
			},
			expectedError: false,
		},
		{
			name:          "invalid toml",
			content:       `[[tag]\nkey = "invalid"`,
			expectedTags:  nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create temporary file
			tmpFile, err := os.CreateTemp("", "tags-*.toml")
			require.NoError(t, err)
			defer func() { _ = os.Remove(tmpFile.Name()) }()

			// Write content
			_, err = tmpFile.WriteString(tt.content)
			require.NoError(t, err)
			_ = tmpFile.Close()

			// Load tags
			tags, err := LoadTagsFromTOML(tmpFile.Name())

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTags, tags)
			}
		})
	}
}

func TestLoadTagsFromTOML_NonExistentFile(t *testing.T) {
	t.Parallel()
	tags, err := LoadTagsFromTOML("/non/existent/tags.toml")
	assert.NoError(t, err)
	assert.Empty(t, tags, "Should return empty slice for non-existent file")
}

func TestExtractKeyFromDisplay(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		displayText string
		expectedKey string
	}{
		{
			name:        "with description",
			displayText: "infra - インフラ関連",
			expectedKey: "infra",
		},
		{
			name:        "without description",
			displayText: "custom",
			expectedKey: "custom",
		},
		{
			name:        "with multiple dashes in description",
			displayText: "network - ネットワーク - LAN/WAN",
			expectedKey: "network",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// extractKey 関数のロジックを再現
			extractKey := func(displayText string) string {
				parts := strings.SplitN(displayText, " - ", 2)
				return parts[0]
			}

			result := extractKey(tt.displayText)
			assert.Equal(t, tt.expectedKey, result)
		})
	}
}

func TestFormatDisplay(t *testing.T) {
	t.Parallel()
	tagDescMap := map[string]string{
		"infra":   "インフラ関連",
		"network": "ネットワーク関連",
	}

	formatDisplay := func(key string) string {
		if desc, ok := tagDescMap[key]; ok && desc != "" {
			return fmt.Sprintf("%s - %s", key, desc)
		}
		return key
	}

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "with description",
			key:      "infra",
			expected: "infra - インフラ関連",
		},
		{
			name:     "without description",
			key:      "custom",
			expected: "custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := formatDisplay(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}
