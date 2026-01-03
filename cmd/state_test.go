package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStateLoadSave(t *testing.T) {
	t.Parallel()

	specPath := t.TempDir()

	// Load from non-existent file returns empty state
	state, err := loadState(specPath)
	if err != nil {
		t.Fatalf("loadState error: %v", err)
	}
	if state.Version != 1 {
		t.Fatalf("expected version 1, got %d", state.Version)
	}
	if len(state.Active) != 0 {
		t.Fatalf("expected empty active list, got %v", state.Active)
	}

	// Save and reload
	state.activateProposal("test-proposal", map[string]string{
		"specification.md": "abc123",
	})

	if err := saveState(specPath, state); err != nil {
		t.Fatalf("saveState error: %v", err)
	}

	loaded, err := loadState(specPath)
	if err != nil {
		t.Fatalf("loadState after save error: %v", err)
	}

	if loaded.Primary != "test-proposal" {
		t.Fatalf("expected primary 'test-proposal', got %q", loaded.Primary)
	}
	if len(loaded.Active) != 1 || loaded.Active[0] != "test-proposal" {
		t.Fatalf("expected active ['test-proposal'], got %v", loaded.Active)
	}
	if loaded.Hashes["test-proposal"]["specification.md"] != "abc123" {
		t.Fatalf("hash mismatch")
	}
}

func TestStateActivateDeactivate(t *testing.T) {
	t.Parallel()

	state := &State{Version: 1, Active: []string{}, Hashes: make(map[string]map[string]string)}

	// Activate first proposal
	state.activateProposal("a", map[string]string{"spec.md": "hash-a"})
	if state.Primary != "a" {
		t.Fatalf("expected primary 'a', got %q", state.Primary)
	}
	if !state.isProposalActive("a") {
		t.Fatal("expected 'a' to be active")
	}

	// Activate second proposal (becomes primary)
	state.activateProposal("b", map[string]string{"spec.md": "hash-b"})
	if state.Primary != "b" {
		t.Fatalf("expected primary 'b', got %q", state.Primary)
	}
	if len(state.Active) != 2 {
		t.Fatalf("expected 2 active, got %d", len(state.Active))
	}

	// Deactivate primary
	state.deactivateProposal("b")
	if state.Primary != "a" {
		t.Fatalf("expected primary to fall back to 'a', got %q", state.Primary)
	}
	if state.isProposalActive("b") {
		t.Fatal("expected 'b' to be inactive")
	}

	// Deactivate last
	state.deactivateProposal("a")
	if state.Primary != "" {
		t.Fatalf("expected empty primary, got %q", state.Primary)
	}
}

func TestHashFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")

	// Non-existent file returns empty hash
	hash, err := hashFile(path)
	if err != nil {
		t.Fatalf("hashFile error: %v", err)
	}
	if hash != "" {
		t.Fatalf("expected empty hash for non-existent file, got %q", hash)
	}

	// Write file and hash it
	content := "# Test\n\nHello world\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	hash, err = hashFile(path)
	if err != nil {
		t.Fatalf("hashFile error: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}

	// Same content = same hash
	hash2, _ := hashFile(path)
	if hash != hash2 {
		t.Fatal("expected same hash for same content")
	}

	// Modified content = different hash
	if err := os.WriteFile(path, []byte(content+"extra"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	hash3, _ := hashFile(path)
	if hash == hash3 {
		t.Fatal("expected different hash for modified content")
	}
}

func TestVerifyProposalHashes(t *testing.T) {
	t.Parallel()

	proposalPath := t.TempDir()

	// Create initial files
	specContent := "# Spec\n"
	designContent := "# Design\n"
	if err := os.WriteFile(filepath.Join(proposalPath, "specification.md"), []byte(specContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(proposalPath, "design.md"), []byte(designContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Compute initial hashes
	hashes, err := computeProposalHashes(proposalPath)
	if err != nil {
		t.Fatalf("computeProposalHashes error: %v", err)
	}

	// Verify unchanged files
	changed, err := verifyProposalHashes(proposalPath, hashes)
	if err != nil {
		t.Fatalf("verifyProposalHashes error: %v", err)
	}
	if len(changed) != 0 {
		t.Fatalf("expected no changes, got %v", changed)
	}

	// Modify a file
	if err := os.WriteFile(filepath.Join(proposalPath, "specification.md"), []byte(specContent+"modified"), 0644); err != nil {
		t.Fatal(err)
	}

	changed, err = verifyProposalHashes(proposalPath, hashes)
	if err != nil {
		t.Fatalf("verifyProposalHashes error: %v", err)
	}
	if len(changed) != 1 || changed[0] != "specification.md" {
		t.Fatalf("expected ['specification.md'], got %v", changed)
	}
}
