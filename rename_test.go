package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateFileNames(t *testing.T) {
	tests := []struct {
		name          string
		setupFiles    []string
		opts          RenameOptions
		expectError   bool
		expectRenamed int
		expectSkipped int
	}{
		{
			name: "rename multiple files",
			setupFiles: []string{
				"file1.txt",
				"file2.pdf",
				"document.doc",
			},
			opts: RenameOptions{
				Writer:     &bytes.Buffer{},
				Extensions: nil,
			},
			expectError:   false,
			expectRenamed: 3,
			expectSkipped: 0,
		},
		{
			name: "skip already formatted files",
			setupFiles: []string{
				"file1.txt",
				"20250903T083109--already-formatted__tag1.pdf",
			},
			opts: RenameOptions{
				Writer:     &bytes.Buffer{},
				Extensions: nil,
			},
			expectError:   false,
			expectRenamed: 1,
			expectSkipped: 1,
		},
		{
			name: "handle files with various extensions",
			setupFiles: []string{
				"test.jpg",
				"test.png",
				"test.md",
				"noextension",
			},
			opts: RenameOptions{
				Writer:     &bytes.Buffer{},
				Extensions: nil,
			},
			expectError:   false,
			expectRenamed: 4,
			expectSkipped: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir, err := os.MkdirTemp("", "parakeet-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			// Setup test files
			for _, filename := range tt.setupFiles {
				filePath := filepath.Join(tmpDir, filename)
				err := os.WriteFile(filePath, []byte("test content"), 0644)
				require.NoError(t, err)
			}

			// Run the function
			err = GenerateFileNames(tmpDir, tt.opts)

			// Check error expectation
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Verify renamed files exist and follow the format
			entries, err := os.ReadDir(tmpDir)
			require.NoError(t, err)

			renamedCount := 0
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}

				// Check if file is formatted (should be for renamed files)
				if IsFormatted(entry.Name()) {
					// Verify the file wasn't already formatted in setup
					alreadyFormatted := false
					for _, setupFile := range tt.setupFiles {
						if setupFile == entry.Name() && IsFormatted(setupFile) {
							alreadyFormatted = true
							break
						}
					}
					if !alreadyFormatted {
						renamedCount++
					}
				}
			}

			assert.Equal(t, tt.expectRenamed, renamedCount, "Number of renamed files should match")
		})
	}
}

func TestGenerateFileNames_NonExistentDirectory(t *testing.T) {
	opts := RenameOptions{
		Writer:     &bytes.Buffer{},
		Extensions: nil,
	}

	err := GenerateFileNames("/non/existent/directory", opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "directory does not exist")
}

func TestGenerateFileNames_EmptyDirectory(t *testing.T) {
	// Create temporary empty directory
	tmpDir, err := os.MkdirTemp("", "parakeet-test-empty-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	opts := RenameOptions{
		Writer:     &bytes.Buffer{},
		Extensions: nil,
	}

	err = GenerateFileNames(tmpDir, opts)
	assert.NoError(t, err)

	// Verify directory is still empty
	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestGenerateFileNames_SkipsDirectories(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-test-subdir-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	// Create a file
	filePath := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(filePath, []byte("test"), 0644)
	require.NoError(t, err)

	opts := RenameOptions{
		Writer:     &bytes.Buffer{},
		Extensions: nil,
	}

	err = GenerateFileNames(tmpDir, opts)
	assert.NoError(t, err)

	// Verify subdirectory still exists with original name
	_, err = os.Stat(subDir)
	assert.NoError(t, err, "Subdirectory should not be renamed")

	// Verify file was renamed
	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)

	fileCount := 0
	dirCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			dirCount++
			assert.Equal(t, "subdir", entry.Name(), "Directory name should not change")
		} else {
			fileCount++
			assert.True(t, IsFormatted(entry.Name()), "File should be formatted")
		}
	}

	assert.Equal(t, 1, fileCount, "Should have one renamed file")
	assert.Equal(t, 1, dirCount, "Should have one directory")
}

func TestGenerateFileNames_PreservesExtension(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-test-ext-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	testFiles := []struct {
		original  string
		expectExt string
	}{
		{"test.pdf", "pdf"},
		{"document.docx", "docx"},
		{"image.jpg", "jpg"},
		{"noext", ""},
	}

	for _, tf := range testFiles {
		filePath := filepath.Join(tmpDir, tf.original)
		err := os.WriteFile(filePath, []byte("test"), 0644)
		require.NoError(t, err)
	}

	opts := RenameOptions{
		Writer:     &bytes.Buffer{},
		Extensions: nil,
	}

	err = GenerateFileNames(tmpDir, opts)
	require.NoError(t, err)

	// Verify extensions are preserved
	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)

	for _, entry := range entries {
		ext := filepath.Ext(entry.Name())
		if ext != "" {
			ext = ext[1:] // Remove leading dot
		}

		// Find matching original file
		found := false
		for _, tf := range testFiles {
			if tf.expectExt == ext {
				found = true
				break
			}
		}
		assert.True(t, found, "Extension %s should match one of the original files", ext)
	}
}

func TestGenerateFileNames_WithExtensionFilter(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-test-ext-filter-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create files with various extensions
	testFiles := []struct {
		name          string
		shouldProcess bool
	}{
		{"file1.txt", true},
		{"file2.pdf", true},
		{"file3.md", false},
		{"file4.jpg", false},
		{"file5.txt", true},
	}

	for _, tf := range testFiles {
		filePath := filepath.Join(tmpDir, tf.name)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		require.NoError(t, err)
	}

	// Test with extension filter
	buf := &bytes.Buffer{}
	opts := RenameOptions{
		Writer:     buf,
		Extensions: []string{"txt", "pdf"},
	}

	err = GenerateFileNames(tmpDir, opts)
	require.NoError(t, err)

	// Verify only txt and pdf files were renamed
	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)

	renamedCount := 0
	for _, entry := range entries {
		if IsFormatted(entry.Name()) {
			renamedCount++
			ext := filepath.Ext(entry.Name())
			if ext != "" {
				ext = ext[1:]
			}
			assert.True(t, ext == "txt" || ext == "pdf", "Only txt and pdf should be renamed")
		}
	}

	assert.Equal(t, 3, renamedCount, "Should rename 3 files")

	// Check output summary
	output := buf.String()
	assert.Contains(t, output, "Processed: 3", "Should process 3 files")
	assert.Contains(t, output, "Skipped: 0", "Should skip 0 files")
}

func TestGenerateFileNames_ActualRename(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-test-actual-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test files
	originalFiles := []string{
		"document.pdf",
		"image.jpg",
		"notes.txt",
	}

	for _, name := range originalFiles {
		filePath := filepath.Join(tmpDir, name)
		err := os.WriteFile(filePath, []byte("test content for "+name), 0644)
		require.NoError(t, err)
	}

	// Perform actual rename
	buf := &bytes.Buffer{}
	opts := RenameOptions{
		Writer:     buf,
		Extensions: nil,
	}

	err = GenerateFileNames(tmpDir, opts)
	require.NoError(t, err)

	// Verify files were actually renamed
	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	require.Len(t, entries, 3, "Should have 3 files")

	// Check all files are now formatted
	for _, entry := range entries {
		assert.True(t, IsFormatted(entry.Name()), "File %s should be formatted", entry.Name())
	}

	// Verify file contents were preserved
	for _, entry := range entries {
		content, err := os.ReadFile(filepath.Join(tmpDir, entry.Name()))
		require.NoError(t, err)
		assert.Contains(t, string(content), "test content for", "File content should be preserved")
	}

	// Check output
	output := buf.String()
	assert.Contains(t, output, "Processed: 3", "Should process 3 files")
	assert.Contains(t, output, "Skipped: 0", "Should skip 0 files")
}
