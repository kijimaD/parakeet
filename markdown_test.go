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

func TestGenerateMarkdownTable(t *testing.T) {
	t.Parallel()
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-markdown-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test files
	testFiles := []string{
		"20250903T083109--TCPIP入門__network_infra.pdf",
		"20250903T083110--sample.txt",
		"20250903T083111--document__important.doc",
		"invalid-file.txt", // このファイルはスキップされる
	}

	for _, name := range testFiles {
		filePath := filepath.Join(tmpDir, name)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		require.NoError(t, err)
	}

	// Generate markdown table
	buf := &bytes.Buffer{}
	opts := MarkdownOptions{
		Writer:     buf,
		Extensions: nil, // すべてのファイルを対象
	}

	err = GenerateMarkdownTable(tmpDir, opts)
	require.NoError(t, err)

	output := buf.String()

	// Check header
	assert.Contains(t, output, "| ID | Title | Tags |")
	assert.Contains(t, output, "|---|---|---|")

	// Check data rows
	assert.Contains(t, output, "| 20250903T083109 | TCPIP入門 | network, infra |")
	assert.Contains(t, output, "| 20250903T083110 | sample |  |")
	assert.Contains(t, output, "| 20250903T083111 | document | important |")

	// Check invalid file is skipped
	assert.NotContains(t, output, "invalid-file")

	// Verify table structure
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Equal(t, 5, len(lines), "Should have header (2 lines) + 3 data rows")
}

func TestGenerateMarkdownTable_WithExtensionFilter(t *testing.T) {
	t.Parallel()
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-markdown-ext-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test files
	testFiles := []string{
		"20250903T083109--document1.pdf",
		"20250903T083110--document2.pdf",
		"20250903T083111--note.txt",
		"20250903T083112--image.jpg",
	}

	for _, name := range testFiles {
		filePath := filepath.Join(tmpDir, name)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		require.NoError(t, err)
	}

	// Generate markdown table with extension filter
	buf := &bytes.Buffer{}
	opts := MarkdownOptions{
		Writer:     buf,
		Extensions: []string{"pdf"},
	}

	err = GenerateMarkdownTable(tmpDir, opts)
	require.NoError(t, err)

	output := buf.String()

	// Check only PDF files are included
	assert.Contains(t, output, "document1")
	assert.Contains(t, output, "document2")
	assert.NotContains(t, output, "note")
	assert.NotContains(t, output, "image")

	// Verify row count
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Equal(t, 4, len(lines), "Should have header (2 lines) + 2 PDF rows")
}

func TestGenerateMarkdownTable_EmptyDirectory(t *testing.T) {
	t.Parallel()
	// Create temporary empty directory
	tmpDir, err := os.MkdirTemp("", "parakeet-markdown-empty-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Generate markdown table
	buf := &bytes.Buffer{}
	opts := MarkdownOptions{
		Writer:     buf,
		Extensions: nil,
	}

	err = GenerateMarkdownTable(tmpDir, opts)
	require.NoError(t, err)

	output := buf.String()

	// Should still have header
	assert.Contains(t, output, "| ID | Title | Tags |")
	assert.Contains(t, output, "|---|---|---|")

	// Should only have header
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Equal(t, 2, len(lines), "Should only have header rows")
}

func TestGenerateMarkdownTable_NoTags(t *testing.T) {
	t.Parallel()
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-markdown-notags-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test files without tags
	testFiles := []string{
		"20250903T083109--file1.pdf",
		"20250903T083110--file2.txt",
	}

	for _, name := range testFiles {
		filePath := filepath.Join(tmpDir, name)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		require.NoError(t, err)
	}

	// Generate markdown table
	buf := &bytes.Buffer{}
	opts := MarkdownOptions{
		Writer:     buf,
		Extensions: nil,
	}

	err = GenerateMarkdownTable(tmpDir, opts)
	require.NoError(t, err)

	output := buf.String()

	// Check that tags column is empty
	assert.Contains(t, output, "| 20250903T083109 | file1 |  |")
	assert.Contains(t, output, "| 20250903T083110 | file2 |  |")
}

func TestGenerateMarkdownTable_NonExistentDirectory(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	opts := MarkdownOptions{
		Writer:     buf,
		Extensions: nil,
	}

	err := GenerateMarkdownTable("/non/existent/directory", opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "directory does not exist")
}

func TestGenerateMarkdownTable_MultipleTags(t *testing.T) {
	t.Parallel()
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "parakeet-markdown-multitags-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test file with multiple tags
	testFiles := []string{
		"20250903T083109--document__tag1_tag2_tag3.pdf",
		"20250903T083110--note__urgent_important.txt",
	}

	for _, name := range testFiles {
		filePath := filepath.Join(tmpDir, name)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		require.NoError(t, err)
	}

	// Generate markdown table
	buf := &bytes.Buffer{}
	opts := MarkdownOptions{
		Writer:     buf,
		Extensions: nil,
	}

	err = GenerateMarkdownTable(tmpDir, opts)
	require.NoError(t, err)

	output := buf.String()

	// Check tags are comma-separated
	assert.Contains(t, output, "| 20250903T083109 | document | tag1, tag2, tag3 |")
	assert.Contains(t, output, "| 20250903T083110 | note | urgent, important |")
}
