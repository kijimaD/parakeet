package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_GenerateAndValidate は実際のディレクトリとファイルを使った統合テスト
func TestIntegration_GenerateAndValidate(t *testing.T) {
	t.Parallel()
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-integration-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test files with various extensions
	testFiles := []string{
		"document.pdf",
		"image.jpg",
		"notes.txt",
		"presentation.pptx",
		"spreadsheet.xlsx",
	}

	for _, name := range testFiles {
		filePath := filepath.Join(tmpDir, name)
		err := os.WriteFile(filePath, []byte("content of "+name), 0644)
		require.NoError(t, err)
	}

	// Step 1: Validate before formatting (should all be invalid)
	validateBuf := &bytes.Buffer{}
	validateOpts := ValidateOptions{
		Writer:     validateBuf,
		Extensions: nil,
	}

	result, err := ValidateFileNames(tmpDir, validateOpts)
	require.NoError(t, err)
	assert.Equal(t, 5, result.TotalFiles)
	assert.Equal(t, 0, result.ValidFiles, "All files should be invalid before formatting")
	assert.Equal(t, 5, len(result.InvalidFiles), "All files should be invalid before formatting")

	output := validateBuf.String()
	assert.Contains(t, output, "Some files have invalid format")

	// Step 2: Generate formatted names
	generateBuf := &bytes.Buffer{}
	generateOpts := RenameOptions{
		Writer:     generateBuf,
		Extensions: []string{"pdf", "jpg", "txt", "pptx", "xlsx"},
	}

	err = GenerateFileNames(tmpDir, generateOpts)
	require.NoError(t, err)

	output = generateBuf.String()
	assert.Contains(t, output, "Processed: 5")
	assert.Contains(t, output, "Skipped: 0")

	// Step 3: Verify files were renamed
	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	require.Len(t, entries, 5, "Should still have 5 files")

	for _, entry := range entries {
		assert.True(t, IsFormatted(entry.Name()), "File %s should be formatted", entry.Name())
	}

	// Step 4: Validate after formatting (should all be valid)
	validateBuf2 := &bytes.Buffer{}
	validateOpts2 := ValidateOptions{
		Writer:     validateBuf2,
		Extensions: nil,
	}

	result2, err := ValidateFileNames(tmpDir, validateOpts2)
	require.NoError(t, err)
	assert.Equal(t, 5, result2.TotalFiles)
	assert.Equal(t, 5, result2.ValidFiles, "All files should be valid after formatting")
	assert.Equal(t, 0, len(result2.InvalidFiles), "No files should be invalid after formatting")

	output2 := validateBuf2.String()
	assert.Contains(t, output2, "All files are properly formatted")

	// Step 5: Verify file contents were preserved
	for _, entry := range entries {
		content, err := os.ReadFile(filepath.Join(tmpDir, entry.Name()))
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(string(content), "content of "), "Content should be preserved")
	}
}

// TestIntegration_ExtensionFiltering は拡張子フィルタリングの統合テスト
func TestIntegration_ExtensionFiltering(t *testing.T) {
	t.Parallel()
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-integration-ext-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create files with various extensions
	testFiles := map[string]string{
		"report.pdf":     "PDF content",
		"readme.md":      "Markdown content",
		"data.csv":       "CSV content",
		"script.py":      "Python script",
		"notes.txt":      "Text notes",
		"config.yaml":    "YAML config",
	}

	for name, content := range testFiles {
		filePath := filepath.Join(tmpDir, name)
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Step 1: Generate formatted names for PDF and TXT files only
	generateBuf := &bytes.Buffer{}
	generateOpts := RenameOptions{
		Writer:     generateBuf,
		Extensions: []string{"pdf", "txt"},
	}

	err = GenerateFileNames(tmpDir, generateOpts)
	require.NoError(t, err)

	output := generateBuf.String()
	assert.Contains(t, output, "Processed: 2", "Should process only 2 files")

	// Step 2: Verify only PDF and TXT were renamed
	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	require.Len(t, entries, 6, "Should still have 6 files")

	formattedCount := 0
	unformattedCount := 0

	for _, entry := range entries {
		if IsFormatted(entry.Name()) {
			formattedCount++
			ext := filepath.Ext(entry.Name())
			if ext != "" {
				ext = ext[1:]
			}
			assert.True(t, ext == "pdf" || ext == "txt", "Only PDF and TXT should be formatted")
		} else {
			unformattedCount++
		}
	}

	assert.Equal(t, 2, formattedCount, "Should have 2 formatted files")
	assert.Equal(t, 4, unformattedCount, "Should have 4 unformatted files")

	// Step 3: Validate only MD and CSV files
	validateBuf := &bytes.Buffer{}
	validateOpts := ValidateOptions{
		Writer:     validateBuf,
		Extensions: []string{"md", "csv"},
	}

	result, err := ValidateFileNames(tmpDir, validateOpts)
	require.NoError(t, err)
	assert.Equal(t, 2, result.TotalFiles, "Should check only 2 files")
	assert.Equal(t, 0, result.ValidFiles, "MD and CSV should be invalid")
	assert.Equal(t, 2, len(result.InvalidFiles), "Should have 2 invalid files")
}

// TestIntegration_MixedScenario は混在シナリオの統合テスト
func TestIntegration_MixedScenario(t *testing.T) {
	t.Parallel()
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-integration-mixed-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a mix of formatted and unformatted files
	files := map[string]bool{
		"20250903T083109--already-formatted.pdf": true,
		"unformatted-file.pdf":                   false,
		"20250903T083110--another-valid.txt":     true,
		"invalid-name.txt":                       false,
		"20250903T083111--image.jpg":             true,
	}

	for name := range files {
		filePath := filepath.Join(tmpDir, name)
		err := os.WriteFile(filePath, []byte("content"), 0644)
		require.NoError(t, err)
	}

	// Step 1: Validate current state
	validateBuf1 := &bytes.Buffer{}
	validateOpts1 := ValidateOptions{
		Writer:     validateBuf1,
		Extensions: nil,
	}

	result1, err := ValidateFileNames(tmpDir, validateOpts1)
	require.NoError(t, err)
	assert.Equal(t, 5, result1.TotalFiles)
	assert.Equal(t, 3, result1.ValidFiles)
	assert.Equal(t, 2, len(result1.InvalidFiles))

	output1 := validateBuf1.String()
	assert.Contains(t, output1, "✗", "Should have crosses for invalid files")

	// Step 2: Generate formatted names for unformatted files
	generateBuf := &bytes.Buffer{}
	generateOpts := RenameOptions{
		Writer:     generateBuf,
		Extensions: []string{"pdf", "txt", "jpg"},
	}

	err = GenerateFileNames(tmpDir, generateOpts)
	require.NoError(t, err)

	output2 := generateBuf.String()
	assert.Contains(t, output2, "Processed: 2", "Should process 2 unformatted files")
	assert.Contains(t, output2, "Skipped: 3", "Should skip 3 already formatted files")

	// Step 3: Validate all files are now formatted
	validateBuf2 := &bytes.Buffer{}
	validateOpts2 := ValidateOptions{
		Writer:     validateBuf2,
		Extensions: nil,
	}

	result2, err := ValidateFileNames(tmpDir, validateOpts2)
	require.NoError(t, err)
	assert.Equal(t, 5, result2.TotalFiles)
	assert.Equal(t, 5, result2.ValidFiles, "All files should be valid now")
	assert.Equal(t, 0, len(result2.InvalidFiles))

	output3 := validateBuf2.String()
	assert.Contains(t, output3, "All files are properly formatted")
}
