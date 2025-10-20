package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateTimestamp(t *testing.T) {
	t.Parallel()
	timestamp := GenerateTimestamp()

	// Check format (YYYYMMDDTHHMMSS)
	assert.Len(t, timestamp, 15, "Timestamp should be 15 characters")
	assert.Contains(t, timestamp, "T", "Timestamp should contain 'T' separator")

	// Check if parseable as expected format
	_, err := time.Parse("20060102T150405", timestamp)
	assert.NoError(t, err, "Timestamp should be parseable")
}

func TestFormatFileName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		components FileNameComponents
		expected   string
	}{
		{
			name: "basic filename with extension",
			components: FileNameComponents{
				Timestamp: "20250903T083109",
				Comment:   "TCPIP入門",
				Tags:      []string{"network", "infra"},
				Extension: "pdf",
			},
			expected: "20250903T083109--TCPIP入門__network_infra.pdf",
		},
		{
			name: "filename without tags",
			components: FileNameComponents{
				Timestamp: "20250903T083109",
				Comment:   "sample",
				Tags:      []string{},
				Extension: "txt",
			},
			expected: "20250903T083109--sample.txt",
		},
		{
			name: "filename with single tag",
			components: FileNameComponents{
				Timestamp: "20250903T083109",
				Comment:   "document",
				Tags:      []string{"important"},
				Extension: "doc",
			},
			expected: "20250903T083109--document__important.doc",
		},
		{
			name: "filename without extension",
			components: FileNameComponents{
				Timestamp: "20250903T083109",
				Comment:   "noext",
				Tags:      []string{"tag1", "tag2"},
				Extension: "",
			},
			expected: "20250903T083109--noext__tag1_tag2",
		},
		{
			name: "filename with multiple tags",
			components: FileNameComponents{
				Timestamp: "20250101T120000",
				Comment:   "multi-tag-file",
				Tags:      []string{"tag1", "tag2", "tag3", "tag4"},
				Extension: "md",
			},
			expected: "20250101T120000--multi-tag-file__tag1_tag2_tag3_tag4.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.components.FormatFileName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseFileName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		expected *FileNameComponents
		wantErr  bool
	}{
		{
			name:     "valid filename with tags",
			filename: "20250903T083109--TCPIP入門__network_infra.pdf",
			expected: &FileNameComponents{
				Timestamp: "20250903T083109",
				Comment:   "TCPIP入門",
				Tags:      []string{"network", "infra"},
				Extension: "pdf",
			},
			wantErr: false,
		},
		{
			name:     "valid filename without tags",
			filename: "20250903T083109--sample.txt",
			expected: &FileNameComponents{
				Timestamp: "20250903T083109",
				Comment:   "sample",
				Tags:      nil,
				Extension: "txt",
			},
			wantErr: false,
		},
		{
			name:     "valid filename with single tag",
			filename: "20250903T083109--document__important.doc",
			expected: &FileNameComponents{
				Timestamp: "20250903T083109",
				Comment:   "document",
				Tags:      []string{"important"},
				Extension: "doc",
			},
			wantErr: false,
		},
		{
			name:     "valid filename without extension",
			filename: "20250903T083109--noext__tag1_tag2",
			expected: &FileNameComponents{
				Timestamp: "20250903T083109",
				Comment:   "noext",
				Tags:      []string{"tag1", "tag2"},
				Extension: "",
			},
			wantErr: false,
		},
		{
			name:     "invalid filename - no double dash",
			filename: "20250903T083109_TCPIP.pdf",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid filename - empty",
			filename: "",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "valid filename with multiple tag groups",
			filename: "20250101T120000--multi__tag1_tag2__tag3_tag4.md",
			expected: &FileNameComponents{
				Timestamp: "20250101T120000",
				Comment:   "multi",
				Tags:      []string{"tag1", "tag2", "tag3", "tag4"},
				Extension: "md",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := ParseFileName(tt.filename)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.Timestamp, result.Timestamp)
				assert.Equal(t, tt.expected.Comment, result.Comment)
				assert.Equal(t, tt.expected.Extension, result.Extension)
				assert.Equal(t, tt.expected.Tags, result.Tags)
			}
		})
	}
}

func TestIsFormatted(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "valid formatted filename",
			filename: "20250903T083109--TCPIP入門__network_infra.pdf",
			expected: true,
		},
		{
			name:     "valid formatted filename without tags",
			filename: "20250903T083109--sample.txt",
			expected: true,
		},
		{
			name:     "invalid filename",
			filename: "regular_file.pdf",
			expected: false,
		},
		{
			name:     "invalid filename - no timestamp",
			filename: "TCPIP入門__network.pdf",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := IsFormatted(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatParseRoundTrip(t *testing.T) {
	t.Parallel()
	// Test that formatting and parsing are inverse operations
	original := FileNameComponents{
		Timestamp: "20250903T083109",
		Comment:   "test-file",
		Tags:      []string{"tag1", "tag2", "tag3"},
		Extension: "pdf",
	}

	formatted := original.FormatFileName()
	parsed, err := ParseFileName(formatted)

	require.NoError(t, err)
	assert.Equal(t, original.Timestamp, parsed.Timestamp)
	assert.Equal(t, original.Comment, parsed.Comment)
	assert.Equal(t, original.Extension, parsed.Extension)
	assert.Equal(t, original.Tags, parsed.Tags)
}

func TestMatchesExtensions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		filename   string
		extensions []string
		expected   bool
	}{
		{
			name:       "no extensions filter - match all",
			filename:   "test.txt",
			extensions: []string{},
			expected:   true,
		},
		{
			name:       "nil extensions filter - match all",
			filename:   "test.pdf",
			extensions: nil,
			expected:   true,
		},
		{
			name:       "exact match",
			filename:   "test.pdf",
			extensions: []string{"pdf"},
			expected:   true,
		},
		{
			name:       "case insensitive match",
			filename:   "test.PDF",
			extensions: []string{"pdf"},
			expected:   true,
		},
		{
			name:       "multiple extensions - first match",
			filename:   "test.txt",
			extensions: []string{"txt", "pdf", "md"},
			expected:   true,
		},
		{
			name:       "multiple extensions - last match",
			filename:   "test.md",
			extensions: []string{"txt", "pdf", "md"},
			expected:   true,
		},
		{
			name:       "no match",
			filename:   "test.jpg",
			extensions: []string{"txt", "pdf", "md"},
			expected:   false,
		},
		{
			name:       "no extension file - match when no ext specified",
			filename:   "test",
			extensions: []string{},
			expected:   true,
		},
		{
			name:       "no extension file - no match when ext specified",
			filename:   "test",
			extensions: []string{"txt"},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := MatchesExtensions(tt.filename, tt.extensions)
			assert.Equal(t, tt.expected, result)
		})
	}
}
