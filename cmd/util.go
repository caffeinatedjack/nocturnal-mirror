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
	specDir        = "spec"
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

// getProposalDependencies reads the specification.md file and extracts the "Depends on" field
func getProposalDependencies(proposalPath string) ([]string, error) {
	specPath := filepath.Join(proposalPath, "specification.md")
	content, err := os.ReadFile(specPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read specification.md: %w", err)
	}

	return parseDependsOn(string(content)), nil
}

// parseDependsOn extracts dependencies from the "**Depends on**:" field in content
func parseDependsOn(content string) []string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Match "**Depends on**:" or "Depends on:" (case-insensitive)
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "**depends on**:") || strings.HasPrefix(lower, "depends on:") {
			// Extract the value after the colon
			idx := strings.Index(trimmed, ":")
			if idx == -1 {
				continue
			}
			value := strings.TrimSpace(trimmed[idx+1:])
			// Remove any trailing comments
			if commentIdx := strings.Index(value, "<!--"); commentIdx != -1 {
				value = strings.TrimSpace(value[:commentIdx])
			}
			// Skip if empty, "none", or still contains template placeholder
			if value == "" || strings.ToLower(value) == "none" || strings.Contains(value, "<!--") {
				return nil
			}
			// Parse comma-separated list
			var deps []string
			for _, dep := range strings.Split(value, ",") {
				dep = strings.TrimSpace(dep)
				if dep != "" {
					deps = append(deps, dep)
				}
			}
			return deps
		}
	}
	return nil
}

// findDependentProposals returns a list of proposals that depend on the given slug
func findDependentProposals(specPath, targetSlug string) ([]string, error) {
	proposalsPath := filepath.Join(specPath, proposalDir)
	entries, err := os.ReadDir(proposalsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read proposals directory: %w", err)
	}

	var dependents []string
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == targetSlug {
			continue
		}

		propPath := filepath.Join(proposalsPath, entry.Name())
		deps, err := getProposalDependencies(propPath)
		if err != nil {
			continue // Skip proposals with unreadable specs
		}

		for _, dep := range deps {
			if dep == targetSlug {
				dependents = append(dependents, entry.Name())
				break
			}
		}
	}

	return dependents, nil
}
