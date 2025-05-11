package main

import (
	"reflect"
	"testing"
)

func TestSanitizeExt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Single extension with dot",
			input:    ".jpg",
			expected: []string{".jpg"},
		},
		{
			name:     "Single extension without dot",
			input:    "png",
			expected: []string{".png"},
		},
		{
			name:     "Multiple extensions with and without dots",
			input:    ".jpg png .gif",
			expected: []string{".jpg", ".png", ".gif"},
		},
		{
			name:     "Empty input",
			input:    "",
			expected: []string{},
		},
		{
			name:     "Input with spaces only",
			input:    "   ",
			expected: []string{},
		},
		{
			name:     "Mixed case extensions",
			input:    ".JPG PnG .Gif",
			expected: []string{".jpg", ".png", ".gif"},
		},
		{
			name:     "Trailing and leading spaces",
			input:    "  .jpg  png  ",
			expected: []string{".jpg", ".png"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeExt(tt.input)
			if len(result) == 0 && len(tt.expected) == 0 {
				return
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("sanitizeExt(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
func TestFindFiles(t *testing.T) {
	tests := []struct {
		name     string
		aFileStr string
		exts     []string
		expected []string
	}{
		{
			name:     "Single valid file with matching extension",
			aFileStr: "image.jpg\n",
			exts:     []string{".jpg"},
			expected: []string{"image.jpg"},
		},
		{
			name:     "Multiple files with mixed extensions",
			aFileStr: "image.jpg\nvideo.mp4\ndocument.pdf\n",
			exts:     []string{".jpg", ".mp4"},
			expected: []string{"image.jpg", "video.mp4"},
		},
		{
			name:     "No matching extensions",
			aFileStr: "image.jpg\nvideo.mp4\ndocument.pdf\n",
			exts:     []string{".png"},
			expected: []string{},
		},
		{
			name:     "Empty input string",
			aFileStr: "",
			exts:     []string{".jpg"},
			expected: []string{},
		},
		{
			name:     "File with no extension",
			aFileStr: "file\n",
			exts:     []string{".jpg"},
			expected: []string{},
		},
		{
			name:     "File with leading and trailing spaces",
			aFileStr: "  image.jpg  \n",
			exts:     []string{".jpg"},
			expected: []string{"image.jpg"},
		},
		{
			name:     "Directory path in input",
			aFileStr: "image.jpg\ndir/\n",
			exts:     []string{".jpg"},
			expected: []string{"image.jpg"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findFiles(tt.aFileStr, tt.exts)
			if len(result) == 0 && len(tt.expected) == 0 {
				return
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("findFiles(%q, %v) = %v, want %v", tt.aFileStr, tt.exts, result, tt.expected)
			}
		})
	}
}

