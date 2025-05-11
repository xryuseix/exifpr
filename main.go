package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/google/go-github/v65/github"
	"golang.org/x/oauth2"
)

func sanitizeExt(extEnv string) []string {
	exts := strings.Split(extEnv, " ")
	var sanitized []string
	for _, ext := range exts {
		if ext == "" {
			continue
		}
		if !strings.HasPrefix(ext, ".") {
			ext = fmt.Sprintf(".%s", ext)
		}
		sanitized = append(sanitized, strings.ToLower(ext))
	}
	return sanitized
}

func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func findFiles(aFileStr string, exts []string) []string {
	var files []string
	for _, path := range strings.Split(aFileStr, "\n") {
		if path == "" {
			continue
		}
		path = strings.TrimSpace(path)
		ext := filepath.Ext(path)
		isDir := IsFile(path)
		if !isDir && slices.Contains(exts, strings.ToLower(ext)) {
			files = append(files, path)
		}
	}
	return files
}

func getExifInfo(path string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("exiftool", path)

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return cmd.String(), stdout.String(), err
}

type ExifInfo struct {
	FilePath string
	StdOut   string
	StdErr   string
}

func genReport(exifs []ExifInfo) string {
	report := "## 📝 Exif Report\n"
	for _, exif := range exifs {
		report += fmt.Sprintf("### %s\n", exif.FilePath)
		report += "<details>\n<summary>Exif Data</summary>\n\n"
		report += "```\n"
		report += exif.StdOut
		report += "\n"
		report += exif.StdErr
		report += "```\n\n"
	}
	return report
}

type Env struct {
	token    string
	owner    string
	repo     string
	prNumber int
}

func getEnv() (Env, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return Env{}, fmt.Errorf("no GitHub Token present")
	}

	repository := os.Getenv("INPUT_REPOSITORY")
	if repository == "" {
		return Env{}, fmt.Errorf("no repository present")
	}
	repoPath := strings.Split(repository, "/")
	owner, repo := repoPath[0], repoPath[1]

	prNumber := os.Getenv("INPUT_PR_NUMBER")
	if prNumber == "" {
		return Env{}, fmt.Errorf("no PR number present")
	}
	prNumberInt, err := strconv.Atoi(prNumber)
	if err != nil {
		return Env{}, fmt.Errorf("error converting PR number to integer: %v", err)
	}

	return Env{token, owner, repo, prNumberInt}, nil
}

func commentToPR(report string) error {
	env, err := getEnv()
	if err != nil {
		return err
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: env.token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	comment := &github.IssueComment{
		Body: github.String(report),
	}
	_, _, err = client.Issues.CreateComment(ctx, env.owner, env.repo, env.prNumber, comment)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	fmt.Println("Starting...")
	extEnv := sanitizeExt(os.Getenv("INPUT_TARGET_EXT"))
	files := findFiles(os.Getenv("INPUT_ADDED_FILES"), extEnv)
	if len(files) == 0 {
		fmt.Println("No files found")
		os.Exit(0)
	}
	var exifs []ExifInfo
	for _, file := range files {
		stdout, stderr, err := getExifInfo(file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		exifs = append(exifs, ExifInfo{file, stdout, stderr})
	}
	report := genReport(exifs)
	if report == "" {
		fmt.Println("No content to report")
		os.Exit(0)
	}
	err := commentToPR(report)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
