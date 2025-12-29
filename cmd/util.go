package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func copyFile(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, content, 0644)
}

func listMarkdownFiles(dirPath string) ([]string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)
	return files, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func cwdPath(elem ...string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return filepath.Join(elem...)
	}
	parts := append([]string{cwd}, elem...)
	return filepath.Join(parts...)
}

const (
	specDir        = "specification"
	ruleDir        = "rule"
	proposalDir    = "proposal"
	archiveDir     = "archive"
	sectionDir     = "section"
	currentSymlink = "current"
	projectFile    = "project.md"
	agentsFile     = "AGENTS.md"
)

func getSpecPath() string {
	return cwdPath(specDir)
}

func requireSpecWorkspace() string {
	specPath := getSpecPath()
	if !fileExists(specPath) {
		printError("Specification workspace not initialized")
		printDim("Run 'nocturnal spec init' first")
		return ""
	}
	return specPath
}

func checkSpecWorkspace() (string, error) {
	specPath := getSpecPath()
	if !fileExists(specPath) {
		return "", fmt.Errorf("specification workspace not initialized. Run 'nocturnal spec init' first")
	}
	return specPath, nil
}

func requireProposal(slug string) string {
	specPath := requireSpecWorkspace()
	if specPath == "" {
		return ""
	}

	proposalPath := filepath.Join(specPath, proposalDir, slug)
	if !fileExists(proposalPath) {
		printError(fmt.Sprintf("Proposal '%s' does not exist", slug))
		return ""
	}
	return proposalPath
}

func getActiveProposalSlug(specPath string) string {
	currentPath := filepath.Join(specPath, currentSymlink)
	target, err := os.Readlink(currentPath)
	if err != nil {
		return ""
	}
	return filepath.Base(target)
}

func getActiveProposal(specPath string) (slug string, proposalPath string, err error) {
	currentPath := filepath.Join(specPath, currentSymlink)

	target, err := os.Readlink(currentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", nil
		}
		return "", "", fmt.Errorf("failed to read current symlink: %w", err)
	}

	slug = filepath.Base(target)
	proposalPath = filepath.Join(specPath, proposalDir, slug)

	if !fileExists(proposalPath) {
		return slug, "", fmt.Errorf("active proposal '%s' no longer exists (stale symlink)", slug)
	}

	return slug, proposalPath, nil
}

func clearActiveProposalIfMatches(specPath, slug string) {
	if getActiveProposalSlug(specPath) == slug {
		os.Remove(filepath.Join(specPath, currentSymlink))
	}
}

const todoFileName = "TODO.md"

func getTodoPath() string {
	return cwdPath(todoFileName)
}

func readTodoFile() (string, error) {
	todoPath := getTodoPath()
	content, err := os.ReadFile(todoPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("TODO.md not found in current directory")
		}
		return "", fmt.Errorf("failed to read TODO.md: %w", err)
	}
	return string(content), nil
}

func writeTodoFile(todoList TodoList) (string, int, error) {
	if len(todoList.Todos) == 0 {
		return "", 0, nil
	}

	todoPath := getTodoPath()
	content := generateTodoContent(todoList.Todos)
	if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
		return "", 0, fmt.Errorf("failed to write TODO.md: %w", err)
	}

	return todoPath, len(todoList.Todos), nil
}

func getContentPreview(content string) string {
	preview := content
	if idx := strings.Index(preview, "\n"); idx > 0 {
		preview = preview[:idx]
	}
	if len(preview) > 60 {
		preview = preview[:57] + "..."
	}
	return preview
}

func nameToSlug(name string) string {
	slug := strings.ToLower(name)

	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	var result strings.Builder
	prevHyphen := false
	for _, c := range slug {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			result.WriteRune(c)
			prevHyphen = false
		} else if c == '-' && !prevHyphen && result.Len() > 0 {
			result.WriteRune(c)
			prevHyphen = true
		}
	}

	return strings.TrimSuffix(result.String(), "-")
}
