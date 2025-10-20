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
				"20250903T083109--TCPIP入門__network_infra.pdf",
				"20250903T083109--sample.txt",
				"20250903T083109--document__important.doc",
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
			// Create temporary directory
			tmpDir, err := os.MkdirTemp("", "parakeet-validate-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			// Setup test files
			for _, filename := range tt.setupFiles {
				filePath := filepath.Join(tmpDir, filename)
				err := os.WriteFile(filePath, []byte("test content"), 0644)
				require.NoError(t, err)
			}

			// Run validation
			buf := &bytes.Buffer{}
			opts := ValidateOptions{
				Verbose:    false,
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

func TestValidateFileNames_Verbose(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-validate-verbose-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Setup test files
	validFile := "20250903T083109--valid.txt"
	invalidFile := "invalid.txt"

	err = os.WriteFile(filepath.Join(tmpDir, validFile), []byte("test"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, invalidFile), []byte("test"), 0644)
	require.NoError(t, err)

	// Run validation with verbose mode
	buf := &bytes.Buffer{}
	opts := ValidateOptions{
		Verbose:    true,
		Writer:     buf,
		Extensions: nil,
	}

	result, err := ValidateFileNames(tmpDir, opts)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check output contains check marks
	output := buf.String()
	assert.Contains(t, output, "✓", "Output should contain valid file marker")
	assert.Contains(t, output, "✗", "Output should contain invalid file marker")
	assert.Contains(t, output, validFile, "Output should contain valid filename")
	assert.Contains(t, output, invalidFile, "Output should contain invalid filename")
}

func TestValidateFileNames_NonExistentDirectory(t *testing.T) {
	buf := &bytes.Buffer{}
	opts := ValidateOptions{
		Verbose:    false,
		Writer:     buf,
		Extensions: nil,
	}

	result, err := ValidateFileNames("/non/existent/directory", opts)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "directory does not exist")
}

func TestValidateFileNames_SkipsDirectories(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-validate-subdir-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

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
		Verbose:    false,
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
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-invalid-files-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

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
	result, err := GetInvalidFiles("/non/existent/directory")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "directory does not exist")
}

func TestGetInvalidFiles_EmptyDirectory(t *testing.T) {
	// Create temporary empty directory
	tmpDir, err := os.MkdirTemp("", "parakeet-empty-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	result, err := GetInvalidFiles(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, result, "Should return empty list for empty directory")
}

func TestValidateFileNames_WithExtensionFilter(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-validate-ext-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

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
		Verbose:    false,
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

func TestValidateFileNames_DetailedOutput(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-validate-detail-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create mixed valid/invalid files
	files := map[string]bool{
		"20250903T083109--document.pdf":   true,
		"20250903T083109--image.jpg":      true,
		"invalid-name.txt":                false,
		"20250903T083109--notes__tag.md":  true,
		"another-invalid.pdf":             false,
	}

	for name := range files {
		filePath := filepath.Join(tmpDir, name)
		err := os.WriteFile(filePath, []byte("content"), 0644)
		require.NoError(t, err)
	}

	// Run validation with verbose
	buf := &bytes.Buffer{}
	opts := ValidateOptions{
		Verbose:    true,
		Writer:     buf,
		Extensions: nil,
	}

	result, err := ValidateFileNames(tmpDir, opts)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check result
	assert.Equal(t, 5, result.TotalFiles)
	assert.Equal(t, 3, result.ValidFiles)
	assert.Equal(t, 2, len(result.InvalidFiles))

	// Check detailed output
	output := buf.String()
	for name, isValid := range files {
		assert.Contains(t, output, name, "Output should mention %s", name)
		if isValid {
			// Verbose mode should show checkmarks for valid files
			assert.True(t, strings.Contains(output, "✓"), "Should have checkmark in output")
		} else {
			assert.Contains(t, output, "✗", "Should have cross mark for invalid files")
			assert.Contains(t, output, "invalid format", "Should mention invalid format")
		}
	}

	assert.Contains(t, output, "Validation Summary", "Should show summary")
	assert.Contains(t, output, "Some files have invalid format", "Should show failure message")
}

func TestValidateFileNames_AllValidOutput(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-validate-allvalid-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create only valid files
	validFiles := []string{
		"20250903T083109--document.pdf",
		"20250903T083109--image.jpg",
		"20250903T083109--notes__tag1_tag2.md",
	}

	for _, name := range validFiles {
		filePath := filepath.Join(tmpDir, name)
		err := os.WriteFile(filePath, []byte("content"), 0644)
		require.NoError(t, err)
	}

	// Run validation
	buf := &bytes.Buffer{}
	opts := ValidateOptions{
		Verbose:    false,
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
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-validate-case-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

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
		Verbose:    false,
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
