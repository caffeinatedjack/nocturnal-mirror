package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var specStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show project statistics and metrics",
	Run:   runSpecStats,
}

func init() {
	specStatsCmd.Long = helpText("spec-stats")
	specCmd.AddCommand(specStatsCmd)
}

// Stats holds aggregated project statistics.
type Stats struct {
	// Specifications
	CompletedSpecs    int
	TotalRequirements int
	MustCount         int
	ShouldCount       int
	MayCount          int

	// Proposals
	ActiveProposals   int
	PendingProposals  int
	ArchivedTotal     int
	ArchivedCompleted int
	ArchivedAbandoned int

	// Current proposal progress
	CurrentProposal  string
	CurrentTotal     int
	CurrentCompleted int
}

func runSpecStats(cmd *cobra.Command, args []string) {
	specPath, err := checkSpecWorkspace()
	if err != nil {
		printWorkspaceError()
		return
	}

	stats, err := gatherStats(specPath)
	if err != nil {
		printError(fmt.Sprintf("Failed to gather stats: %v", err))
		return
	}

	fmt.Println()

	// Specifications section
	fmt.Println(boldStyle.Render("Specifications"))
	fmt.Println()
	fmt.Printf("  Completed: %d\n", stats.CompletedSpecs)
	if stats.TotalRequirements > 0 {
		fmt.Printf("  Requirements: %d ", stats.TotalRequirements)
		parts := []string{}
		if stats.MustCount > 0 {
			parts = append(parts, fmt.Sprintf("MUST: %d", stats.MustCount))
		}
		if stats.ShouldCount > 0 {
			parts = append(parts, fmt.Sprintf("SHOULD: %d", stats.ShouldCount))
		}
		if stats.MayCount > 0 {
			parts = append(parts, fmt.Sprintf("MAY: %d", stats.MayCount))
		}
		if len(parts) > 0 {
			fmt.Printf("%s\n", dimStyle.Render("("+strings.Join(parts, ", ")+")"))
		} else {
			fmt.Println()
		}
	} else {
		fmt.Printf("  Requirements: %s\n", dimStyle.Render("0"))
	}
	fmt.Println()

	// Proposals section
	fmt.Println(boldStyle.Render("Proposals"))
	fmt.Println()
	fmt.Printf("  Active: %d\n", stats.ActiveProposals)
	fmt.Printf("  Pending: %d\n", stats.PendingProposals)
	if stats.ArchivedTotal > 0 {
		fmt.Printf("  Archived: %d ", stats.ArchivedTotal)
		fmt.Printf("%s\n", dimStyle.Render(fmt.Sprintf("(%d completed, %d abandoned)", stats.ArchivedCompleted, stats.ArchivedAbandoned)))
	} else {
		fmt.Printf("  Archived: %s\n", dimStyle.Render("0"))
	}
	fmt.Println()

	// Progress section
	fmt.Println(boldStyle.Render("Progress"))
	fmt.Println()
	if stats.CurrentProposal != "" {
		fmt.Printf("  Current: %s\n", infoStyle.Render(stats.CurrentProposal))
		if stats.CurrentTotal > 0 {
			percentage := (stats.CurrentCompleted * 100) / stats.CurrentTotal
			progressBar := renderProgressBar(stats.CurrentCompleted, stats.CurrentTotal, 20)
			fmt.Printf("  Tasks: %s %s\n", progressBar, dimStyle.Render(fmt.Sprintf("%d/%d (%d%%)", stats.CurrentCompleted, stats.CurrentTotal, percentage)))
		} else {
			fmt.Printf("  Tasks: %s\n", dimStyle.Render("no tasks defined"))
		}
	} else {
		fmt.Printf("  Current: %s\n", dimStyle.Render("no active proposal"))
	}
	fmt.Println()
}

func gatherStats(specPath string) (*Stats, error) {
	stats := &Stats{}

	// Count completed specifications and their requirements
	sectionPath := filepath.Join(specPath, sectionDir)
	sectionFiles, err := listMarkdownFiles(sectionPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read section directory: %w", err)
	}
	stats.CompletedSpecs = len(sectionFiles)

	for _, filename := range sectionFiles {
		filePath := filepath.Join(sectionPath, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		must, should, may := countRequirementsByType(string(content))
		stats.MustCount += must
		stats.ShouldCount += should
		stats.MayCount += may
	}
	stats.TotalRequirements = stats.MustCount + stats.ShouldCount + stats.MayCount

	// Count proposals
	state, err := loadState(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	proposalsPath := filepath.Join(specPath, proposalDir)
	entries, err := os.ReadDir(proposalsPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read proposals directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if state.isProposalActive(entry.Name()) {
				stats.ActiveProposals++
			} else {
				stats.PendingProposals++
			}
		}
	}

	// Count archived proposals
	archivePath := filepath.Join(specPath, archiveDir)
	archiveEntries, err := os.ReadDir(archivePath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read archive directory: %w", err)
	}

	for _, entry := range archiveEntries {
		if entry.IsDir() {
			stats.ArchivedTotal++
			abandonedPath := filepath.Join(archivePath, entry.Name(), ".abandoned")
			if fileExists(abandonedPath) {
				stats.ArchivedAbandoned++
			} else {
				stats.ArchivedCompleted++
			}
		}
	}

	// Get current proposal progress
	if state.Primary != "" {
		stats.CurrentProposal = state.Primary
		proposalPath := filepath.Join(specPath, proposalDir, state.Primary)
		stats.CurrentTotal, stats.CurrentCompleted = getProposalProgress(proposalPath)
	}

	return stats, nil
}

// countRequirementsByType counts MUST, SHOULD, and MAY keywords in content.
func countRequirementsByType(content string) (must, should, may int) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		upper := strings.ToUpper(line)
		if strings.Contains(upper, "MUST NOT") || strings.Contains(upper, "MUST") {
			must++
		} else if strings.Contains(upper, "SHOULD NOT") || strings.Contains(upper, "SHOULD") {
			should++
		} else if strings.Contains(upper, "MAY") {
			may++
		}
	}
	return must, should, may
}
