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

func TestValidateFileNames(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		setupFiles        []string
		expectedValid     int
		expectedInvalid   int
		expectedTotal     int
		shouldContainText string
	}{
		{
			name: "all files valid",
			setupFiles: []string{
				"20250903T083109--TCPIP入門.pdf",
				"20250903T083110--sample.txt",
				"20250903T083111--document.doc",
			},
			expectedValid:     3,
			expectedInvalid:   0,
			expectedTotal:     3,
			shouldContainText: "All files are properly formatted",
		},
		{
			name: "all files invalid",
			setupFiles: []string{
				"invalid1.txt",
				"invalid2.pdf",
				"no-format.doc",
			},
			expectedValid:     0,
			expectedInvalid:   3,
			expectedTotal:     3,
			shouldContainText: "Some files have invalid format",
		},
		{
			name: "mixed valid and invalid",
			setupFiles: []string{
				"20250903T083109--valid-file.txt",
				"invalid-file.txt",
				"20250903T083109--another-valid__tag1.pdf",
				"bad-format.doc",
			},
			expectedValid:     2,
			expectedInvalid:   2,
			expectedTotal:     4,
			shouldContainText: "Some files have invalid format",
		},
		{
			name:              "empty directory",
			setupFiles:        []string{},
			expectedValid:     0,
			expectedInvalid:   0,
			expectedTotal:     0,
			shouldContainText: "All files are properly formatted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create temporary directory
			tmpDir, err := os.MkdirTemp("", "parakeet-validate-test-*")
			require.NoError(t, err)
			defer func() { _ = os.RemoveAll(tmpDir) }()

			// Setup test files
			for _, filename := range tt.setupFiles {
				filePath := filepath.Join(tmpDir, filename)
				err := os.WriteFile(filePath, []byte("test content"), 0644)
				require.NoError(t, err)
			}

			// Run validation
			buf := &bytes.Buffer{}
			opts := ValidateOptions{
				Writer:     buf,
				Extensions: nil,
			}

			result, err := ValidateFileNames(tmpDir, opts)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Check results
			assert.Equal(t, tt.expectedTotal, result.TotalFiles, "Total files should match")
			assert.Equal(t, tt.expectedValid, result.ValidFiles, "Valid files should match")
			assert.Equal(t, tt.expectedInvalid, len(result.InvalidFiles), "Invalid files should match")

			// Check output
			output := buf.String()
			assert.Contains(t, output, tt.shouldContainText, "Output should contain expected text")
		})
	}
}


func TestValidateFileNames_NonExistentDirectory(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	opts := ValidateOptions{
		Writer:     buf,
		Extensions: nil,
	}

	result, err := ValidateFileNames("/non/existent/directory", opts)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "directory does not exist")
}

func TestValidateFileNames_SkipsDirectories(t *testing.T) {
	t.Parallel()
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-validate-subdir-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	// Create a file
	validFile := "20250903T083109--test.txt"
	err = os.WriteFile(filepath.Join(tmpDir, validFile), []byte("test"), 0644)
	require.NoError(t, err)

	// Run validation
	buf := &bytes.Buffer{}
	opts := ValidateOptions{
		Writer:     buf,
		Extensions: nil,
	}

	result, err := ValidateFileNames(tmpDir, opts)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should only count the file, not the directory
	assert.Equal(t, 1, result.TotalFiles, "Should only count files, not directories")
	assert.Equal(t, 1, result.ValidFiles, "File should be valid")
}

func TestValidateFileName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "valid filename",
			filename: "20250903T083109--test__tag1.txt",
			wantErr:  false,
		},
		{
			name:     "valid filename without tags",
			filename: "20250903T083109--test.txt",
			wantErr:  false,
		},
		{
			name:     "invalid filename format",
			filename: "invalid.txt",
			wantErr:  true,
		},
		{
			name:     "invalid timestamp",
			filename: "2025--test.txt",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateFileName(tt.filename)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetInvalidFiles(t *testing.T) {
	t.Parallel()
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-invalid-files-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Setup test files
	validFiles := []string{
		"20250903T083109--valid1.txt",
		"20250903T083109--valid2.pdf",
	}
	invalidFiles := []string{
		"invalid1.txt",
		"invalid2.pdf",
	}

	for _, f := range validFiles {
		err := os.WriteFile(filepath.Join(tmpDir, f), []byte("test"), 0644)
		require.NoError(t, err)
	}
	for _, f := range invalidFiles {
		err := os.WriteFile(filepath.Join(tmpDir, f), []byte("test"), 0644)
		require.NoError(t, err)
	}

	// Get invalid files
	result, err := GetInvalidFiles(tmpDir)
	require.NoError(t, err)

	// Should return only invalid files
	assert.Equal(t, len(invalidFiles), len(result), "Should return only invalid files")

	// Check that all invalid files are in the result
	for _, invalidFile := range invalidFiles {
		found := false
		for _, resultFile := range result {
			if strings.Contains(resultFile, invalidFile) {
				found = true
				break
			}
		}
		assert.True(t, found, "Invalid file %s should be in result", invalidFile)
	}
}

func TestGetInvalidFiles_NonExistentDirectory(t *testing.T) {
	t.Parallel()
	result, err := GetInvalidFiles("/non/existent/directory")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "directory does not exist")
}

func TestGetInvalidFiles_EmptyDirectory(t *testing.T) {
	t.Parallel()
	// Create temporary empty directory
	tmpDir, err := os.MkdirTemp("", "parakeet-empty-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	result, err := GetInvalidFiles(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, result, "Should return empty list for empty directory")
}

func TestValidateFileNames_WithExtensionFilter(t *testing.T) {
	t.Parallel()
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-validate-ext-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create files with various extensions
	testFiles := []struct {
		name        string
		shouldCheck bool
	}{
		{"20250903T083109--valid.txt", true},
		{"20250903T083109--valid.pdf", true},
		{"20250903T083109--valid.md", false},
		{"invalid.txt", true},
		{"invalid.pdf", true},
		{"invalid.jpg", false},
	}

	for _, tf := range testFiles {
		filePath := filepath.Join(tmpDir, tf.name)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		require.NoError(t, err)
	}

	// Test with extension filter (txt, pdf only)
	buf := &bytes.Buffer{}
	opts := ValidateOptions{
		Writer:     buf,
		Extensions: []string{"txt", "pdf"},
	}

	result, err := ValidateFileNames(tmpDir, opts)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should only check txt and pdf files
	assert.Equal(t, 4, result.TotalFiles, "Should check 4 files (txt and pdf only)")
	assert.Equal(t, 2, result.ValidFiles, "Should have 2 valid files")
	assert.Equal(t, 2, len(result.InvalidFiles), "Should have 2 invalid files")

	// Check output
	output := buf.String()
	assert.Contains(t, output, "invalid.txt", "Should mention invalid txt file")
	assert.Contains(t, output, "invalid.pdf", "Should mention invalid pdf file")
	assert.NotContains(t, output, "invalid.jpg", "Should not check jpg file")
	assert.NotContains(t, output, ".md", "Should not check md files")
	assert.Contains(t, output, "Total files: 4", "Should show correct total")
}

func TestValidateFileNames_AllValidOutput(t *testing.T) {
	t.Parallel()
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-validate-allvalid-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create only valid files
	validFiles := []string{
		"20250903T083109--document.pdf",
		"20250903T083110--image.jpg",
		"20250903T083111--notes.md",
	}

	for _, name := range validFiles {
		filePath := filepath.Join(tmpDir, name)
		err := os.WriteFile(filePath, []byte("content"), 0644)
		require.NoError(t, err)
	}

	// Run validation
	buf := &bytes.Buffer{}
	opts := ValidateOptions{
		Writer:     buf,
		Extensions: nil,
	}

	result, err := ValidateFileNames(tmpDir, opts)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check result
	assert.Equal(t, 3, result.TotalFiles)
	assert.Equal(t, 3, result.ValidFiles)
	assert.Equal(t, 0, len(result.InvalidFiles))

	// Check success output
	output := buf.String()
	assert.Contains(t, output, "All files are properly formatted", "Should show success message")
	assert.Contains(t, output, "Total files: 3", "Should show total count")
	assert.Contains(t, output, "Valid: 3", "Should show valid count")
	assert.Contains(t, output, "Invalid: 0", "Should show zero invalid")
}

func TestValidateFileNames_CaseInsensitiveExtension(t *testing.T) {
	t.Parallel()
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-validate-case-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create files with different case extensions
	files := []string{
		"20250903T083109--test.PDF",
		"20250903T083109--test.Pdf",
		"20250903T083109--test.txt",
		"invalid.TXT",
	}

	for _, name := range files {
		filePath := filepath.Join(tmpDir, name)
		err := os.WriteFile(filePath, []byte("content"), 0644)
		require.NoError(t, err)
	}

	// Test with lowercase extension filter
	buf := &bytes.Buffer{}
	opts := ValidateOptions{
		Writer:     buf,
		Extensions: []string{"pdf", "txt"},
	}

	result, err := ValidateFileNames(tmpDir, opts)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should match case-insensitively
	assert.Equal(t, 4, result.TotalFiles, "Should check all pdf and txt files regardless of case")
	assert.Equal(t, 3, result.ValidFiles, "Should have 3 valid files")
	assert.Equal(t, 1, len(result.InvalidFiles), "Should have 1 invalid file")

	// Check that invalid.TXT is caught
	output := buf.String()
	assert.Contains(t, output, "invalid.TXT", "Should find invalid TXT file")
}

func TestValidateFileNames_WithDuplicateTimestamps(t *testing.T) {
	t.Parallel()
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-validate-duplicate-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create files with duplicate timestamps
	testFiles := []string{
		"20250903T083109--file1.txt",
		"20250903T083109--file2.pdf",  // 同じタイムスタンプ
		"20250903T083110--file3.doc",
		"20250903T083110--file4.jpg",  // 同じタイムスタンプ
		"20250903T083111--file5.md",   // ユニーク
	}

	for _, name := range testFiles {
		filePath := filepath.Join(tmpDir, name)
		err := os.WriteFile(filePath, []byte("content"), 0644)
		require.NoError(t, err)
	}

	// Run validation
	buf := &bytes.Buffer{}
	opts := ValidateOptions{
		Writer:     buf,
		Extensions: nil,
	}

	result, err := ValidateFileNames(tmpDir, opts)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check result
	assert.Equal(t, 5, result.TotalFiles)
	assert.Equal(t, 5, result.ValidFiles)
	assert.Equal(t, 0, len(result.InvalidFiles))
	assert.True(t, result.HasDuplicates, "Should detect duplicates")
	assert.Equal(t, 4, len(result.DuplicateFiles), "Should have 4 duplicate files")

	// Check output
	output := buf.String()
	assert.Contains(t, output, "⚠", "Should show warning for duplicates")
	assert.Contains(t, output, "duplicate timestamp", "Should mention duplicate timestamps")
	assert.Contains(t, output, "Duplicates: 4", "Should show duplicate count")
}

func TestValidateFileNames_NoDuplicates(t *testing.T) {
	t.Parallel()
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-validate-nodup-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create files with unique timestamps
	testFiles := []string{
		"20250903T083109--file1.txt",
		"20250903T083110--file2.pdf",
		"20250903T083111--file3.doc",
	}

	for _, name := range testFiles {
		filePath := filepath.Join(tmpDir, name)
		err := os.WriteFile(filePath, []byte("content"), 0644)
		require.NoError(t, err)
	}

	// Run validation
	buf := &bytes.Buffer{}
	opts := ValidateOptions{
		Writer:     buf,
		Extensions: nil,
	}

	result, err := ValidateFileNames(tmpDir, opts)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check result
	assert.Equal(t, 3, result.TotalFiles)
	assert.Equal(t, 3, result.ValidFiles)
	assert.Equal(t, 0, len(result.InvalidFiles))
	assert.False(t, result.HasDuplicates, "Should not detect duplicates")
	assert.Equal(t, 0, len(result.DuplicateFiles), "Should have 0 duplicate files")

	// Check output
	output := buf.String()
	assert.Contains(t, output, "All files are properly formatted", "Should show success message")
	assert.NotContains(t, output, "⚠", "Should not show warnings")
}

func TestValidateFileNames_WithUndefinedTags(t *testing.T) {
	t.Parallel()
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-validate-undefined-tags-*")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	// Create tag.toml with defined tags
	tomlContent := `[[tag]]
key = "network"
desc = "Network related"

[[tag]]
key = "infra"
desc = "Infrastructure"

[[tag]]
key = "security"
desc = "Security related"
`
	tomlPath := filepath.Join(tmpDir, "tag.toml")
	err = os.WriteFile(tomlPath, []byte(tomlContent), 0644)
	require.NoError(t, err)

	// Create files with valid and invalid tags
	testFiles := []string{
		"20250903T083109--file1__network_infra.txt",      // 定義済みタグ
		"20250903T083110--file2__undefined_tag.pdf",      // 未定義タグ
		"20250903T083111--file3__security.doc",           // 定義済みタグ
		"20250903T083112--file4__network_invalid.jpg",   // 1つ定義済み、1つ未定義
	}

	for _, name := range testFiles {
		filePath := filepath.Join(tmpDir, name)
		err := os.WriteFile(filePath, []byte("content"), 0644)
		require.NoError(t, err)
	}

	// Run validation
	buf := &bytes.Buffer{}
	opts := ValidateOptions{
		Writer:     buf,
		Extensions: nil,
	}

	result, err := ValidateFileNames(tmpDir, opts)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check result (tag.toml is counted as invalid file)
	assert.Equal(t, 5, result.TotalFiles, "Should count 4 test files + tag.toml")
	assert.Equal(t, 4, result.ValidFiles)
	assert.Equal(t, 1, len(result.InvalidFiles), "tag.toml is invalid format")
	assert.True(t, result.HasUndefinedTags, "Should detect undefined tags")
	assert.Equal(t, 2, len(result.UndefinedTagFiles), "Should have 2 files with undefined tags")

	// Check specific undefined tags
	assert.Contains(t, result.UndefinedTagFiles, "20250903T083110--file2__undefined_tag.pdf")
	assert.Contains(t, result.UndefinedTagFiles["20250903T083110--file2__undefined_tag.pdf"], "undefined")
	assert.Contains(t, result.UndefinedTagFiles["20250903T083110--file2__undefined_tag.pdf"], "tag")

	assert.Contains(t, result.UndefinedTagFiles, "20250903T083112--file4__network_invalid.jpg")
	assert.Contains(t, result.UndefinedTagFiles["20250903T083112--file4__network_invalid.jpg"], "invalid")
	assert.NotContains(t, result.UndefinedTagFiles["20250903T083112--file4__network_invalid.jpg"], "network")

	// Check output
	output := buf.String()
	assert.Contains(t, output, "⚠", "Should show warning for undefined tags")
	assert.Contains(t, output, "undefined tags", "Should mention undefined tags")
	assert.Contains(t, output, "Undefined tags: 2", "Should show undefined tag count")
}

func TestValidateFileNames_NoTagToml(t *testing.T) {
	t.Parallel()
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-validate-no-toml-*")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	// Create files with any tags (no tag.toml in tmpDir)
	testFiles := []string{
		"20250903T083109--file1__anytag.txt",
		"20250903T083110--file2__random_tags.pdf",
	}

	for _, name := range testFiles {
		filePath := filepath.Join(tmpDir, name)
		err := os.WriteFile(filePath, []byte("content"), 0644)
		require.NoError(t, err)
	}

	// Run validation
	buf := &bytes.Buffer{}
	opts := ValidateOptions{
		Writer:     buf,
		Extensions: nil,
	}

	result, err := ValidateFileNames(tmpDir, opts)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check result - should not check tags if tag.toml doesn't exist
	assert.Equal(t, 2, result.TotalFiles)
	assert.Equal(t, 2, result.ValidFiles)
	assert.False(t, result.HasUndefinedTags, "Should not check tags when tag.toml doesn't exist")
	assert.Equal(t, 0, len(result.UndefinedTagFiles), "Should have 0 files with undefined tags")

	// Check output
	output := buf.String()
	assert.Contains(t, output, "All files are properly formatted", "Should show success message")
}
