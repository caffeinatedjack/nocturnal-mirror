package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// copyFile copies a file from src to dst with 0644 permissions.
func copyFile(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, content, 0644)
}

// listMarkdownFiles returns sorted .md filenames in a directory.
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

// fileExists returns true if the path exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// cwdPath joins path elements with the current working directory.
func cwdPath(elem ...string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return filepath.Join(elem...)
	}
	parts := append([]string{cwd}, elem...)
	return filepath.Join(parts...)
}

const (
	specDir     = "spec"
	ruleDir     = "rule"
	proposalDir = "proposal"
	archiveDir  = "archive"
	sectionDir  = "section"
	projectFile = "project.md"
	agentsFile  = "AGENTS.md"
)

var proposalDocFiles = []string{"specification.md", "design.md", "implementation.md"}

// getSpecPath returns the path to the spec/ directory.
func getSpecPath() string {
	return cwdPath(specDir)
}

// checkSpecWorkspace returns the spec path or an error if not initialized.
func checkSpecWorkspace() (string, error) {
	specPath := getSpecPath()
	if !fileExists(specPath) {
		return "", fmt.Errorf("specification workspace not initialized. Run 'nocturnal spec init' first")
	}
	return specPath, nil
}

// checkProposal returns the proposal path or an error if it doesn't exist.
func checkProposal(specPath, slug string) (string, error) {
	proposalPath := filepath.Join(specPath, proposalDir, slug)
	if !fileExists(proposalPath) {
		return "", fmt.Errorf("proposal '%s' does not exist", slug)
	}
	return proposalPath, nil
}

// printWorkspaceError prints the standard workspace not initialized error
func printWorkspaceError() {
	printError("Specification workspace not initialized")
	printDim("Run 'nocturnal spec init' first")
}

// getActiveProposalSlug returns the primary active proposal name, or empty if none.
// Deprecated: use getPrimaryProposalSlug from state.go instead.
func getActiveProposalSlug(specPath string) string {
	return getPrimaryProposalSlug(specPath)
}

// getActiveProposal returns the primary active proposal's slug and path.
// Deprecated: use getPrimaryProposal from state.go instead.
func getActiveProposal(specPath string) (slug string, proposalPath string, err error) {
	return getPrimaryProposal(specPath)
}

// clearActiveProposalIfMatches removes a proposal from active state if it matches.
// Deprecated: use clearProposalIfMatches from state.go instead.
func clearActiveProposalIfMatches(specPath, slug string) {
	_ = clearProposalIfMatches(specPath, slug)
}

// archiveProposalDocs copies proposal documents to the archive directory
func archiveProposalDocs(proposalPath, archivePath string, files []string) error {
	if err := os.MkdirAll(archivePath, 0755); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}

	for _, filename := range files {
		src := filepath.Join(proposalPath, filename)
		if fileExists(src) {
			dst := filepath.Join(archivePath, filename)
			if err := copyFile(src, dst); err != nil {
				return fmt.Errorf("failed to archive %s: %w", filename, err)
			}
		}
	}
	return nil
}

// getContentPreview returns the first line of content, truncated to 60 chars.
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

// nameToSlug converts a name to a URL-safe lowercase slug.
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

// containsHeaderWithText checks if content has a markdown header containing the given text (case-insensitive)
func containsHeaderWithText(content, text string) bool {
	lowerText := strings.ToLower(text)
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			if strings.Contains(strings.ToLower(trimmed), lowerText) {
				return true
			}
		}
	}
	return false
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

// getMissingCompletedDependencies returns dependencies that are not completed.
// A dependency is considered completed when it exists in spec/section/<dep>.md.
func getMissingCompletedDependencies(specPath, proposalPath string) ([]string, error) {
	deps, err := getProposalDependencies(proposalPath)
	if err != nil {
		return nil, err
	}

	var missing []string
	for _, dep := range deps {
		completedSpecPath := filepath.Join(specPath, sectionDir, dep+".md")
		if !fileExists(completedSpecPath) {
			missing = append(missing, dep)
		}
	}

	sort.Strings(missing)
	return missing, nil
}
