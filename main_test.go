package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
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
	// filterExt calls IsFile, so test files must exist on disk
	dir := t.TempDir()
	createFile := func(name string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte{}, 0644); err != nil {
			t.Fatal(err)
		}
		return p
	}

	jpg := createFile("image.jpg")
	pdf := createFile("document.pdf")
	png := createFile("photo.png")
	gif := createFile("graphic.gif")
	jpgUpper := createFile("image.JPG")
	pngUpper := createFile("photo.PNG")
	gifUpper := createFile("graphic.GIF")
	readme := createFile("README")
	license := createFile("LICENSE")

	tests := []struct {
		name      string
		files     []string
		targetExt []string
		expected  []string
	}{
		{
			name:      "Single matching extension",
			files:     []string{jpg, pdf, png},
			targetExt: []string{".jpg"},
			expected:  []string{jpg},
		},
		{
			name:      "Multiple matching extensions",
			files:     []string{jpg, png, gif},
			targetExt: []string{".jpg", ".png"},
			expected:  []string{jpg, png},
		},
		{
			name:      "No matching extensions",
			files:     []string{jpg, png},
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
			files:     []string{jpg, png},
			targetExt: []string{},
			expected:  []string{},
		},
		{
			name:      "Case insensitive matching",
			files:     []string{jpgUpper, pngUpper, gifUpper},
			targetExt: []string{".jpg", ".png"},
			expected:  []string{jpgUpper, pngUpper},
		},
		{
			name:      "Files without extensions",
			files:     []string{readme, license, jpg},
			targetExt: []string{".jpg"},
			expected:  []string{jpg},
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

func TestIsFile(t *testing.T) {
	t.Run("Regular file returns true", func(t *testing.T) {
		f, err := os.CreateTemp(t.TempDir(), "testfile")
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
		if !IsFile(f.Name()) {
			t.Errorf("IsFile(%q) = false, want true", f.Name())
		}
	})

	t.Run("Directory returns false", func(t *testing.T) {
		dir := t.TempDir()
		if IsFile(dir) {
			t.Errorf("IsFile(%q) = true, want false", dir)
		}
	})

	t.Run("Non-existent path returns false", func(t *testing.T) {
		if IsFile("/nonexistent/path/file.txt") {
			t.Error("IsFile for non-existent path = true, want false")
		}
	})
}

func TestGenReport(t *testing.T) {
	t.Run("Single file", func(t *testing.T) {
		exifs := []ExifInfo{
			{FilePath: "image.jpg", StdOut: "Width: 100\nHeight: 200", StdErr: ""},
		}
		result := genReport(exifs)
		if !strings.Contains(result, "## 📝 Exif Report") {
			t.Error("genReport() missing report header")
		}
		if !strings.Contains(result, "### image.jpg") {
			t.Error("genReport() missing file header")
		}
		if !strings.Contains(result, "Width: 100") {
			t.Error("genReport() missing exif content")
		}
	})

	t.Run("Multiple files", func(t *testing.T) {
		exifs := []ExifInfo{
			{FilePath: "a.jpg", StdOut: "out1", StdErr: ""},
			{FilePath: "b.png", StdOut: "out2", StdErr: "err2"},
		}
		result := genReport(exifs)
		if !strings.Contains(result, "### a.jpg") || !strings.Contains(result, "### b.png") {
			t.Errorf("genReport() missing file headers: %s", result)
		}
		if !strings.Contains(result, "err2") {
			t.Error("genReport() missing stderr content")
		}
	})

	t.Run("Empty input", func(t *testing.T) {
		result := genReport([]ExifInfo{})
		if result != "## 📝 Exif Report\n" {
			t.Errorf("genReport([]) = %q, want header only", result)
		}
	})

	t.Run("Contains details tag", func(t *testing.T) {
		exifs := []ExifInfo{
			{FilePath: "test.jpg", StdOut: "data", StdErr: ""},
		}
		result := genReport(exifs)
		if !strings.Contains(result, "<details>") || !strings.Contains(result, "</details>") {
			t.Error("genReport() missing details tag")
		}
	})
}

func TestGetEnv(t *testing.T) {
	t.Run("Valid environment", func(t *testing.T) {
		t.Setenv("INPUT_TARGET_EXT", ".jpg .png")
		t.Setenv("GITHUB_TOKEN", "test-token")
		t.Setenv("INPUT_REPOSITORY", "owner/repo")
		t.Setenv("INPUT_PR_NUMBER", "42")

		env, err := getEnv()
		if err != nil {
			t.Fatalf("getEnv() returned error: %v", err)
		}
		if env.token != "test-token" {
			t.Errorf("token = %q, want %q", env.token, "test-token")
		}
		if env.owner != "owner" {
			t.Errorf("owner = %q, want %q", env.owner, "owner")
		}
		if env.repo != "repo" {
			t.Errorf("repo = %q, want %q", env.repo, "repo")
		}
		if env.prNumber != 42 {
			t.Errorf("prNumber = %d, want %d", env.prNumber, 42)
		}
		if !reflect.DeepEqual(env.targetExt, []string{".jpg", ".png"}) {
			t.Errorf("targetExt = %v, want %v", env.targetExt, []string{".jpg", ".png"})
		}
	})

	t.Run("Missing GITHUB_TOKEN", func(t *testing.T) {
		t.Setenv("GITHUB_TOKEN", "")
		t.Setenv("INPUT_REPOSITORY", "owner/repo")
		t.Setenv("INPUT_PR_NUMBER", "42")

		_, err := getEnv()
		if err == nil {
			t.Error("getEnv() should return error when GITHUB_TOKEN is missing")
		}
	})

	t.Run("Missing repository", func(t *testing.T) {
		t.Setenv("GITHUB_TOKEN", "token")
		t.Setenv("INPUT_REPOSITORY", "")
		t.Setenv("INPUT_PR_NUMBER", "42")

		_, err := getEnv()
		if err == nil {
			t.Error("getEnv() should return error when INPUT_REPOSITORY is missing")
		}
	})

	t.Run("Missing PR number", func(t *testing.T) {
		t.Setenv("GITHUB_TOKEN", "token")
		t.Setenv("INPUT_REPOSITORY", "owner/repo")
		t.Setenv("INPUT_PR_NUMBER", "")

		_, err := getEnv()
		if err == nil {
			t.Error("getEnv() should return error when INPUT_PR_NUMBER is missing")
		}
	})

	t.Run("Invalid PR number", func(t *testing.T) {
		t.Setenv("GITHUB_TOKEN", "token")
		t.Setenv("INPUT_REPOSITORY", "owner/repo")
		t.Setenv("INPUT_PR_NUMBER", "abc")

		_, err := getEnv()
		if err == nil {
			t.Error("getEnv() should return error for non-numeric PR number")
		}
	})
}
