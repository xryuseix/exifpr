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

func TestFilterExt(t *testing.T) {
	tests := []struct {
		name      string
		files     []string
		targetExt []string
		expected  []string
	}{
		{
			name:      "Single matching extension",
			files:     []string{"image.jpg", "document.pdf", "photo.png"},
			targetExt: []string{".jpg"},
			expected:  []string{"image.jpg"},
		},
		{
			name:      "Multiple matching extensions",
			files:     []string{"image.jpg", "photo.png", "graphic.gif"},
			targetExt: []string{".jpg", ".png"},
			expected:  []string{"image.jpg", "photo.png"},
		},
		{
			name:      "No matching extensions",
			files:     []string{"image.jpg", "photo.png"},
			targetExt: []string{".gif"},
			expected:  []string{},
		},
		{
			name:      "Empty file list",
			files:     []string{},
			targetExt: []string{".jpg"},
			expected:  []string{},
		},
		{
			name:      "Empty target extensions",
			files:     []string{"image.jpg", "photo.png"},
			targetExt: []string{},
			expected:  []string{},
		},
		{
			name:      "Case insensitive matching",
			files:     []string{"image.JPG", "photo.PNG", "graphic.GIF"},
			targetExt: []string{".jpg", ".png"},
			expected:  []string{"image.JPG", "photo.PNG"},
		},
		{
			name:      "Files without extensions",
			files:     []string{"README", "LICENSE", "image.jpg"},
			targetExt: []string{".jpg"},
			expected:  []string{"image.jpg"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterExt(tt.files, tt.targetExt)
			if len(result) == 0 && len(tt.expected) == 0 {
				return
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("filterExt(%v, %v) = %v, want %v", tt.files, tt.targetExt, result, tt.expected)
			}
		})
	}
}
