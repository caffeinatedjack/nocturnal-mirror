package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const stateFile = ".nocturnal.json"

// State represents the nocturnal state file (spec/.nocturnal.json).
type State struct {
	Version int                          `json:"version"`
	Active  []string                     `json:"active"`
	Primary string                       `json:"primary"`
	Hashes  map[string]map[string]string `json:"hashes,omitempty"`
}

// getStatePath returns the path to the state file.
func getStatePath(specPath string) string {
	return filepath.Join(specPath, stateFile)
}

// loadState reads the state file. Returns empty state if file doesn't exist.
func loadState(specPath string) (*State, error) {
	statePath := getStatePath(specPath)
	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{Version: 1, Active: []string{}, Hashes: make(map[string]map[string]string)}, nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	if state.Hashes == nil {
		state.Hashes = make(map[string]map[string]string)
	}

	return &state, nil
}

// saveState writes the state file.
func saveState(specPath string, state *State) error {
	statePath := getStatePath(specPath)
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize state: %w", err)
	}

	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
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

// verifyProposalHashes checks if current file hashes match stored hashes.
// Returns list of changed files (empty if all match).
func verifyProposalHashes(proposalPath string, storedHashes map[string]string) ([]string, error) {
	var changed []string

	for _, filename := range proposalDocFiles {
		filePath := filepath.Join(proposalPath, filename)
		currentHash, err := hashFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to hash %s: %w", filename, err)
		}

		storedHash, exists := storedHashes[filename]
		if !exists && currentHash != "" {
			// New file added since activation
			changed = append(changed, filename)
		} else if exists && currentHash != storedHash {
			// File content changed
			changed = append(changed, filename)
		}
	}

	return changed, nil
}

// isProposalActive checks if a proposal is in the active list.
func (s *State) isProposalActive(slug string) bool {
	for _, active := range s.Active {
		if active == slug {
			return true
		}
	}
	return false
}

// activateProposal adds a proposal to the active list and sets it as primary.
func (s *State) activateProposal(slug string, hashes map[string]string) {
	if !s.isProposalActive(slug) {
		s.Active = append(s.Active, slug)
	}
	s.Primary = slug
	s.Hashes[slug] = hashes
}

// deactivateProposal removes a proposal from the active list.
func (s *State) deactivateProposal(slug string) {
	var newActive []string
	for _, active := range s.Active {
		if active != slug {
			newActive = append(newActive, active)
		}
	}
	s.Active = newActive
	delete(s.Hashes, slug)

	// Update primary if needed
	if s.Primary == slug {
		if len(s.Active) > 0 {
			s.Primary = s.Active[0]
		} else {
			s.Primary = ""
		}
	}
}

// getPrimaryProposal returns the primary proposal slug and path.
func getPrimaryProposal(specPath string) (slug string, proposalPath string, err error) {
	state, err := loadState(specPath)
	if err != nil {
		return "", "", err
	}

	if state.Primary == "" {
		return "", "", nil
	}

	proposalPath = filepath.Join(specPath, proposalDir, state.Primary)
	if !fileExists(proposalPath) {
		return state.Primary, "", fmt.Errorf("primary proposal '%s' no longer exists (stale state)", state.Primary)
	}

	return state.Primary, proposalPath, nil
}

// getPrimaryProposalSlug returns just the primary proposal slug.
func getPrimaryProposalSlug(specPath string) string {
	state, err := loadState(specPath)
	if err != nil {
		return ""
	}
	return state.Primary
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

// checkProposalIntegrity verifies that a proposal's files haven't changed since activation.
// Returns changed files list and whether confirmation is required.
func checkProposalIntegrity(specPath, slug string) (changedFiles []string, requiresConfirmation bool, err error) {
	state, err := loadState(specPath)
	if err != nil {
		return nil, false, err
	}

	storedHashes, exists := state.Hashes[slug]
	if !exists {
		// No hashes stored, can't verify
		return nil, false, nil
	}

	proposalPath := filepath.Join(specPath, proposalDir, slug)
	changed, err := verifyProposalHashes(proposalPath, storedHashes)
	if err != nil {
		return nil, false, err
	}

	return changed, len(changed) > 0, nil
}
