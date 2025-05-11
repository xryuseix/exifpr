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

func getGitHubClient(token string) (*github.Client, context.Context) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return client, ctx
}

func filterExt(files []string, targetExt []string) []string {
	var filtered []string
	for _, file := range files {
		ext := filepath.Ext(file)
		if slices.Contains(targetExt, strings.ToLower(ext)) {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

func findFiles(env Env) ([]string, error) {
	client, ctx := getGitHubClient(env.token)

	var allFiles []string
	opt := &github.ListOptions{PerPage: 100}
	for {
		files, resp, err := client.PullRequests.ListFiles(ctx, env.owner, env.repo, env.prNumber, opt)
		if err != nil {
			fmt.Printf("Error getting PR files: %v\n", err)
			return []string{}, err
		}
		var aFiles []*github.CommitFile
		aFiles = append(aFiles, files...)
		for _, file := range aFiles {
			if file.GetStatus() != "added" {
				continue
			}
			allFiles = append(allFiles, file.GetFilename())
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return filterExt(allFiles, env.targetExt), nil
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
	targetExt []string
	token     string
	owner     string
	repo      string
	prNumber  int
}

func getEnv() (Env, error) {
	targetExt := sanitizeExt(os.Getenv("INPUT_TARGET_EXT"))

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

	return Env{targetExt, token, owner, repo, prNumberInt}, nil
}

func commentToPR(report string, env Env) error {
	client, ctx := getGitHubClient(env.token)
	comment := &github.IssueComment{
		Body: github.String(report),
	}
	_, _, err := client.Issues.CreateComment(ctx, env.owner, env.repo, env.prNumber, comment)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	fmt.Println("Starting...")
	env, err := getEnv()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	files, err := findFiles(env)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
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
	err = commentToPR(report, env)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
