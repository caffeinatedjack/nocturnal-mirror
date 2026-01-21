package tui

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	proposalDir = "proposal"
	archiveDir  = "archive"
	sectionDir  = "section"
	stateFile   = ".nocturnal.json"
)

var proposalDocFiles = []string{"specification.md", "design.md", "implementation.md"}

// Helper functions copied/adapted from cmd package

// fileExists returns true if the path exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// copyFile copies a file from src to dst with 0644 permissions.
func copyFile(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, content, 0644)
}

// getStatePath returns the path to the state file.
func getStatePath(specPath string) string {
	return filepath.Join(specPath, stateFile)
}

// saveState writes the state file.
func saveState(specPath string, state *State) error {
	statePath := getStatePath(specPath)
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize state: %w", err)
	}
	return os.WriteFile(statePath, data, 0644)
}

// hashFile computes SHA256 hash of a file's contents.
func hashFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	h := sha256.Sum256(content)
	return hex.EncodeToString(h[:]), nil
}

// computeProposalHashes computes hashes for all proposal documents.
func computeProposalHashes(proposalPath string) (map[string]string, error) {
	hashes := make(map[string]string)

	for _, filename := range proposalDocFiles {
		filePath := filepath.Join(proposalPath, filename)
		hash, err := hashFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to hash %s: %w", filename, err)
		}
		if hash != "" {
			hashes[filename] = hash
		}
	}

	return hashes, nil
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

// parseDependsOn extracts dependencies from the "**Depends on**:" field in content
func parseDependsOn(content string) []string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "**depends on**:") || strings.HasPrefix(lower, "depends on:") {
			idx := strings.Index(trimmed, ":")
			if idx == -1 {
				continue
			}
			value := strings.TrimSpace(trimmed[idx+1:])
			if commentIdx := strings.Index(value, "<!--"); commentIdx != -1 {
				value = strings.TrimSpace(value[:commentIdx])
			}
			if value == "" || strings.ToLower(value) == "none" || strings.Contains(value, "<!--") {
				return nil
			}
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

// getMissingCompletedDependencies returns dependencies that are not completed.
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

// clearProposalIfMatches removes a proposal from active/primary if it matches.
func clearProposalIfMatches(specPath, slug string) error {
	state, err := loadState(specPath)
	if err != nil {
		return err
	}

	if state.isProposalActive(slug) {
		state.deactivateProposal(slug)
		return saveState(specPath, state)
	}
	return nil
}

// State methods

func (s *State) isProposalActive(slug string) bool {
	for _, active := range s.Active {
		if active == slug {
			return true
		}
	}
	return false
}

func (s *State) activateProposal(slug string, hashes map[string]string) {
	if !s.isProposalActive(slug) {
		s.Active = append(s.Active, slug)
	}
	s.Primary = slug
	s.Hashes[slug] = hashes
}

func (s *State) deactivateProposal(slug string) {
	var newActive []string
	for _, active := range s.Active {
		if active != slug {
			newActive = append(newActive, active)
		}
	}
	s.Active = newActive
	delete(s.Hashes, slug)

	if s.Primary == slug {
		if len(s.Active) > 0 {
			s.Primary = s.Active[0]
		} else {
			s.Primary = ""
		}
	}
}

// ActivateProposal activates a proposal by slug.
func ActivateProposal(specPath, slug string) tea.Cmd {
	return func() tea.Msg {
		proposalPath := filepath.Join(specPath, "proposal", slug)

		// Check if proposal exists
		if _, err := os.Stat(proposalPath); os.IsNotExist(err) {
			return ErrorMsg{Err: fmt.Errorf("proposal '%s' not found", slug)}
		}

		// Check dependencies
		missing, err := getMissingCompletedDependencies(specPath, proposalPath)
		if err != nil {
			return ErrorMsg{Err: fmt.Errorf("failed to check dependencies: %w", err)}
		}
		if len(missing) > 0 {
			return ErrorMsg{Err: fmt.Errorf("missing dependencies: %s", strings.Join(missing, ", "))}
		}

		// Compute hashes
		hashes, err := computeProposalHashes(proposalPath)
		if err != nil {
			return ErrorMsg{Err: fmt.Errorf("failed to compute hashes: %w", err)}
		}

		// Load state and activate
		state, err := loadState(specPath)
		if err != nil {
			return ErrorMsg{Err: fmt.Errorf("failed to load state: %w", err)}
		}

		state.activateProposal(slug, hashes)

		if err := saveState(specPath, state); err != nil {
			return ErrorMsg{Err: fmt.Errorf("failed to save state: %w", err)}
		}

		return SuccessMsg{Message: fmt.Sprintf("Activated proposal: %s", slug)}
	}
}

// DeactivateProposal deactivates the current proposal.
func DeactivateProposal(specPath string) tea.Cmd {
	return func() tea.Msg {
		state, err := loadState(specPath)
		if err != nil {
			return ErrorMsg{Err: fmt.Errorf("failed to load state: %w", err)}
		}

		if state.Primary == "" {
			return ErrorMsg{Err: fmt.Errorf("no active proposal")}
		}

		slug := state.Primary
		state.deactivateProposal(slug)

		if err := saveState(specPath, state); err != nil {
			return ErrorMsg{Err: fmt.Errorf("failed to save state: %w", err)}
		}

		return SuccessMsg{Message: fmt.Sprintf("Deactivated proposal: %s", slug)}
	}
}

// CompleteProposal completes a proposal by slug.
func CompleteProposal(specPath, slug string) tea.Cmd {
	return func() tea.Msg {
		proposalPath := filepath.Join(specPath, "proposal", slug)

		// Check if proposal exists
		if _, err := os.Stat(proposalPath); os.IsNotExist(err) {
			return ErrorMsg{Err: fmt.Errorf("proposal '%s' not found", slug)}
		}

		archivePath := filepath.Join(specPath, "archive", slug)
		sectionPath := filepath.Join(specPath, "section")

		specFile := filepath.Join(proposalPath, "specification.md")
		if _, err := os.Stat(specFile); os.IsNotExist(err) {
			return ErrorMsg{Err: fmt.Errorf("proposal '%s' is missing specification.md", slug)}
		}

		// Archive design and implementation documents
		if err := archiveProposalDocs(proposalPath, archivePath, []string{"design.md", "implementation.md"}); err != nil {
			return ErrorMsg{Err: err}
		}

		// Promote specification to section
		specDst := filepath.Join(sectionPath, slug+".md")
		if err := copyFile(specFile, specDst); err != nil {
			return ErrorMsg{Err: fmt.Errorf("failed to promote specification: %w", err)}
		}

		if err := os.RemoveAll(proposalPath); err != nil {
			return ErrorMsg{Err: fmt.Errorf("failed to remove proposal workspace: %w", err)}
		}

		clearProposalIfMatches(specPath, slug)

		return SuccessMsg{Message: fmt.Sprintf("Completed proposal: %s", slug)}
	}
}

// ValidateProposal validates a proposal by slug.
func ValidateProposal(specPath, slug string) tea.Cmd {
	return func() tea.Msg {
		proposalPath := filepath.Join(specPath, "proposal", slug)

		// Check if proposal exists
		if _, err := os.Stat(proposalPath); os.IsNotExist(err) {
			return ErrorMsg{Err: fmt.Errorf("proposal '%s' not found", slug)}
		}

		// Check for required files
		errors := []string{}
		warnings := []string{}

		// Check specification.md
		specFile := filepath.Join(proposalPath, "specification.md")
		if _, err := os.Stat(specFile); os.IsNotExist(err) {
			errors = append(errors, "missing specification.md")
		}

		// Check implementation.md
		implFile := filepath.Join(proposalPath, "implementation.md")
		if _, err := os.Stat(implFile); os.IsNotExist(err) {
			warnings = append(warnings, "missing implementation.md (recommended)")
		}

		// Check design.md
		designFile := filepath.Join(proposalPath, "design.md")
		if _, err := os.Stat(designFile); os.IsNotExist(err) {
			warnings = append(warnings, "missing design.md (optional)")
		}

		// Check dependencies
		missing, err := getMissingCompletedDependencies(specPath, proposalPath)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("failed to check dependencies: %v", err))
		} else if len(missing) > 0 {
			warnings = append(warnings, fmt.Sprintf("missing dependencies: %s", strings.Join(missing, ", ")))
		}

		if len(errors) > 0 {
			return ErrorMsg{Err: fmt.Errorf("validation failed:\n  - %s", strings.Join(errors, "\n  - "))}
		}

		msg := fmt.Sprintf("Proposal '%s' is valid", slug)
		if len(warnings) > 0 {
			msg += fmt.Sprintf("\nWarnings:\n  - %s", strings.Join(warnings, "\n  - "))
		}

		return SuccessMsg{Message: msg}
	}
}

// DeleteProposal deletes a proposal by slug.
func DeleteProposal(specPath, slug string, force bool) tea.Cmd {
	return func() tea.Msg {
		proposalPath := filepath.Join(specPath, "proposal", slug)

		// Check if proposal exists
		if _, err := os.Stat(proposalPath); os.IsNotExist(err) {
			return ErrorMsg{Err: fmt.Errorf("proposal '%s' not found", slug)}
		}

		// Check if proposal is active
		state, err := loadState(specPath)
		if err != nil {
			return ErrorMsg{Err: fmt.Errorf("failed to load state: %w", err)}
		}

		if !force && state.isProposalActive(slug) {
			return ErrorMsg{Err: fmt.Errorf("proposal '%s' is active; deactivate first or use force", slug)}
		}

		// Remove proposal directory
		if err := os.RemoveAll(proposalPath); err != nil {
			return ErrorMsg{Err: fmt.Errorf("failed to remove proposal: %w", err)}
		}

		// Clear from state if active
		clearProposalIfMatches(specPath, slug)

		return SuccessMsg{Message: fmt.Sprintf("Deleted proposal: %s", slug)}
	}
}
