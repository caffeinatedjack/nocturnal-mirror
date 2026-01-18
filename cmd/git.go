package cmd

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// GitSnapshotManager handles git snapshots and commits for task execution
type GitSnapshotManager struct {
	specPath     string
	proposalSlug string
	taskID       string
	snapshotRef  string // The git ref at snapshot time
}

// NewGitSnapshotManager creates a new git snapshot manager for a task
func NewGitSnapshotManager(specPath, proposalSlug, taskID string) *GitSnapshotManager {
	return &GitSnapshotManager{
		specPath:     specPath,
		proposalSlug: proposalSlug,
		taskID:       taskID,
	}
}

// CreateSnapshot creates a git commit snapshot before starting a task
func (g *GitSnapshotManager) CreateSnapshot() error {
	if !isGitRepo() {
		return nil // Skip if not in a git repo
	}

	// Check if there are any changes to snapshot
	hasChanges, err := g.hasUncommittedChanges()
	if err != nil {
		return fmt.Errorf("failed to check for uncommitted changes: %w", err)
	}

	if !hasChanges {
		// No changes to snapshot, just record the current ref
		ref, err := g.getCurrentRef()
		if err != nil {
			return fmt.Errorf("failed to get current ref: %w", err)
		}
		g.snapshotRef = ref
		return nil
	}

	// Create snapshot commit
	commitMsg := fmt.Sprintf("WIP: snapshot before task %s [%s]", g.taskID, g.proposalSlug)
	if err := g.gitAddAll(); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	if err := g.gitCommit(commitMsg); err != nil {
		return fmt.Errorf("failed to create snapshot commit: %w", err)
	}

	// Record the snapshot ref
	ref, err := g.getCurrentRef()
	if err != nil {
		return fmt.Errorf("failed to get snapshot ref: %w", err)
	}
	g.snapshotRef = ref

	return nil
}

// CommitChanges creates a final commit for the completed task
func (g *GitSnapshotManager) CommitChanges(taskText string) error {
	if !isGitRepo() {
		return nil // Skip if not in a git repo
	}

	// Check if there are any changes to commit
	hasChanges, err := g.hasUncommittedChanges()
	if err != nil {
		return fmt.Errorf("failed to check for uncommitted changes: %w", err)
	}

	if !hasChanges {
		return nil // No changes to commit
	}

	// Stage all changes
	if err := g.gitAddAll(); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	// Create commit with descriptive message
	commitMsg := g.generateCommitMessage(taskText)
	if err := g.gitCommit(commitMsg); err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	return nil
}

// hasUncommittedChanges checks if there are uncommitted changes in the repo
func (g *GitSnapshotManager) hasUncommittedChanges() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(output))) > 0, nil
}

// getCurrentRef returns the current git ref (commit hash)
func (g *GitSnapshotManager) getCurrentRef() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// gitAddAll stages all changes
func (g *GitSnapshotManager) gitAddAll() error {
	cmd := exec.Command("git", "add", "-A")
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// gitCommit creates a commit with the given message
func (g *GitSnapshotManager) gitCommit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// generateCommitMessage creates a structured commit message for the task
func (g *GitSnapshotManager) generateCommitMessage(taskText string) string {
	// Clean up the task text
	taskText = strings.TrimSpace(taskText)

	// Create structured commit message
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	return fmt.Sprintf(`feat: complete task %s

%s

Proposal: %s
Completed: %s`, g.taskID, taskText, g.proposalSlug, timestamp)
}

// GetSnapshotRef returns the snapshot reference if one was created
func (g *GitSnapshotManager) GetSnapshotRef() string {
	return g.snapshotRef
}
