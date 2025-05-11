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
