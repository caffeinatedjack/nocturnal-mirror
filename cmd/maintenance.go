package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var maintenanceCmd = &cobra.Command{
	Use:   "maintenance",
	Short: "Manage maintenance requirements",
}

var maintenanceAddCmd = &cobra.Command{
	Use:   "add <name-or-slug>",
	Short: "Create a new maintenance item",
	Args:  cobra.ExactArgs(1),
	Run:   runMaintenanceAdd,
}

var maintenanceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all maintenance items with due counts",
	Run:   runMaintenanceList,
}

var maintenanceShowCmd = &cobra.Command{
	Use:   "show <slug>",
	Short: "Show a maintenance item",
	Args:  cobra.ExactArgs(1),
	Run:   runMaintenanceShow,
}

var maintenanceDueCmd = &cobra.Command{
	Use:   "due <slug>",
	Short: "Show due requirements for a maintenance item",
	Args:  cobra.ExactArgs(1),
	Run:   runMaintenanceDue,
}

var maintenanceActionedCmd = &cobra.Command{
	Use:   "actioned <slug> <id>",
	Short: "Mark a requirement as actioned",
	Args:  cobra.ExactArgs(2),
	Run:   runMaintenanceActioned,
}

var maintenanceRemoveCmd = &cobra.Command{
	Use:   "remove <slug>",
	Short: "Remove a maintenance item",
	Args:  cobra.ExactArgs(1),
	Run:   runMaintenanceRemove,
}

func init() {
	maintenanceCmd.Long = helpText("spec-maintenance")
	maintenanceAddCmd.Long = helpText("spec-maintenance-add")
	maintenanceListCmd.Long = helpText("spec-maintenance-list")
	maintenanceShowCmd.Long = helpText("spec-maintenance-show")
	maintenanceDueCmd.Long = helpText("spec-maintenance-due")
	maintenanceActionedCmd.Long = helpText("spec-maintenance-actioned")

	maintenanceCmd.AddCommand(maintenanceAddCmd)
	maintenanceCmd.AddCommand(maintenanceListCmd)
	maintenanceCmd.AddCommand(maintenanceShowCmd)
	maintenanceCmd.AddCommand(maintenanceDueCmd)
	maintenanceCmd.AddCommand(maintenanceActionedCmd)
	maintenanceCmd.AddCommand(maintenanceRemoveCmd)

	specCmd.AddCommand(maintenanceCmd)
}

// MaintenanceRequirement represents a parsed requirement from a maintenance file.
type MaintenanceRequirement struct {
	ID           string
	Text         string
	Freq         string // daily, weekly, biweekly, monthly, quarterly, yearly, or empty (always)
	Due          bool
	LastActioned string // RFC3339 timestamp or empty
	Line         int    // 1-indexed line number in file
}

var allowedFreqs = map[string]bool{
	"daily":     true,
	"weekly":    true,
	"biweekly":  true,
	"monthly":   true,
	"quarterly": true,
	"yearly":    true,
}

// parseMaintenanceFile reads and parses a maintenance file.
func parseMaintenanceFile(filePath string, state *State, slug string) ([]MaintenanceRequirement, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var requirements []MaintenanceRequirement
	inRequirements := false
	seenIDs := make(map[string]int) // id -> line number

	// Regex to extract tokens: [id=...] [freq=...]
	idPattern := regexp.MustCompile(`\[id=([^\]]+)\]`)
	freqPattern := regexp.MustCompile(`\[freq=([^\]]+)\]`)

	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect start of Requirements section
		if strings.HasPrefix(trimmed, "## Requirements") {
			inRequirements = true
			continue
		}

		// Stop at next section
		if inRequirements && strings.HasPrefix(trimmed, "## ") {
			break
		}

		// Parse requirement lines
		if inRequirements && (strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ")) {
			// Extract ID
			idMatch := idPattern.FindStringSubmatch(trimmed)
			if len(idMatch) < 2 {
				return nil, fmt.Errorf("line %d: requirement missing [id=...]: %s", lineNum+1, trimmed)
			}
			id := strings.TrimSpace(idMatch[1])

			// Check for duplicate IDs
			if prevLine, exists := seenIDs[id]; exists {
				return nil, fmt.Errorf("line %d: duplicate id '%s' (first seen on line %d)", lineNum+1, id, prevLine)
			}
			seenIDs[id] = lineNum + 1

			// Extract frequency (optional)
			freq := ""
			freqMatch := freqPattern.FindStringSubmatch(trimmed)
			if len(freqMatch) >= 2 {
				freq = strings.TrimSpace(freqMatch[1])
				if !allowedFreqs[freq] {
					return nil, fmt.Errorf("line %d: unknown frequency '%s' (allowed: daily, weekly, biweekly, monthly, quarterly, yearly)", lineNum+1, freq)
				}
			}

			// Strip tokens to get clean text
			text := trimmed
			text = idPattern.ReplaceAllString(text, "")
			text = freqPattern.ReplaceAllString(text, "")
			text = strings.TrimSpace(text)
			// Remove leading bullet
			text = strings.TrimPrefix(text, "- ")
			text = strings.TrimPrefix(text, "* ")
			text = strings.TrimSpace(text)

			// Get last actioned time from state
			lastActioned := ""
			if state != nil && state.Maintenance != nil {
				if slugMap, ok := state.Maintenance[slug]; ok {
					if reqState, ok := slugMap[id]; ok {
						lastActioned = reqState.LastActioned
					}
				}
			}

			// Compute due status
			due := computeDue(freq, lastActioned)

			requirements = append(requirements, MaintenanceRequirement{
				ID:           id,
				Text:         text,
				Freq:         freq,
				Due:          due,
				LastActioned: lastActioned,
				Line:         lineNum + 1,
			})
		}
	}

	return requirements, nil
}

// computeDue determines if a requirement is due based on frequency and last actioned time.
func computeDue(freq string, lastActioned string) bool {
	// No freq => always due
	if freq == "" {
		return true
	}

	// Never actioned => due
	if lastActioned == "" {
		return true
	}

	// Parse last actioned time
	lastTime, err := time.Parse(time.RFC3339, lastActioned)
	if err != nil {
		// Invalid timestamp => treat as never actioned
		return true
	}

	now := time.Now()
	var nextDue time.Time

	switch freq {
	case "daily":
		nextDue = lastTime.AddDate(0, 0, 1)
	case "weekly":
		nextDue = lastTime.AddDate(0, 0, 7)
	case "biweekly":
		nextDue = lastTime.AddDate(0, 0, 14)
	case "monthly":
		nextDue = lastTime.AddDate(0, 1, 0)
	case "quarterly":
		nextDue = lastTime.AddDate(0, 3, 0)
	case "yearly":
		nextDue = lastTime.AddDate(1, 0, 0)
	default:
		// Unknown freq => always due
		return true
	}

	return now.After(nextDue) || now.Equal(nextDue)
}

// listMaintenanceFiles returns sorted maintenance file slugs.
func listMaintenanceFiles(specPath string) ([]string, error) {
	maintenancePath := filepath.Join(specPath, maintenanceDir)
	entries, err := os.ReadDir(maintenancePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var slugs []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			slugs = append(slugs, strings.TrimSuffix(entry.Name(), ".md"))
		}
	}
	return slugs, nil
}

func runMaintenanceAdd(cmd *cobra.Command, args []string) {
	name := args[0]
	slug := nameToSlug(name)

	if slug == "" {
		printError("Invalid maintenance name: must contain at least one alphanumeric character")
		return
	}

	specPath, err := checkSpecWorkspace()
	if err != nil {
		printWorkspaceError()
		return
	}

	maintenancePath := filepath.Join(specPath, maintenanceDir)
	if err := os.MkdirAll(maintenancePath, 0755); err != nil {
		printError(fmt.Sprintf("Failed to create maintenance directory: %v", err))
		return
	}

	filePath := filepath.Join(maintenancePath, slug+".md")
	if fileExists(filePath) {
		printError(fmt.Sprintf("Maintenance item '%s' already exists", slug))
		return
	}

	// Render template
	data := struct {
		Name string
		Slug string
	}{Name: name, Slug: slug}

	content, err := renderTemplate("templates/maintenance.md", data)
	if err != nil {
		printError(fmt.Sprintf("Failed to render template: %v", err))
		return
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		printError(fmt.Sprintf("Failed to create maintenance item: %v", err))
		return
	}

	printSuccess(fmt.Sprintf("Created maintenance item '%s'", slug))
	printDim(fmt.Sprintf("Location: %s", filePath))
}

func runMaintenanceList(cmd *cobra.Command, args []string) {
	specPath, err := checkSpecWorkspace()
	if err != nil {
		printWorkspaceError()
		return
	}

	slugs, err := listMaintenanceFiles(specPath)
	if err != nil {
		printError(fmt.Sprintf("Failed to list maintenance items: %v", err))
		return
	}

	if len(slugs) == 0 {
		printDim("No maintenance items found")
		printDim("Use 'nocturnal spec maintenance add <name>' to create one")
		return
	}

	state, err := loadState(specPath)
	if err != nil {
		printError(fmt.Sprintf("Failed to load state: %v", err))
		return
	}

	fmt.Println()
	fmt.Println(boldStyle.Render(fmt.Sprintf("Maintenance Items (%d)", len(slugs))))
	fmt.Println()

	for _, slug := range slugs {
		filePath := filepath.Join(specPath, maintenanceDir, slug+".md")
		reqs, err := parseMaintenanceFile(filePath, state, slug)
		if err != nil {
			printError(fmt.Sprintf("Error parsing %s: %v", slug, err))
			continue
		}

		dueCount := 0
		for _, req := range reqs {
			if req.Due {
				dueCount++
			}
		}

		dueText := fmt.Sprintf("%d/%d due", dueCount, len(reqs))
		if dueCount > 0 {
			dueText = warningStyle.Render(dueText)
		} else {
			dueText = dimStyle.Render(dueText)
		}

		fmt.Printf("  %s  %s\n", infoStyle.Render(slug), dueText)
	}
	fmt.Println()
}

func runMaintenanceShow(cmd *cobra.Command, args []string) {
	slug := args[0]
	specPath, err := checkSpecWorkspace()
	if err != nil {
		printWorkspaceError()
		return
	}

	filePath := filepath.Join(specPath, maintenanceDir, slug+".md")
	if !fileExists(filePath) {
		printError(fmt.Sprintf("Maintenance item '%s' does not exist", slug))
		return
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		printError(fmt.Sprintf("Failed to read maintenance item: %v", err))
		return
	}

	fmt.Print(string(content))
}

func runMaintenanceDue(cmd *cobra.Command, args []string) {
	slug := args[0]
	specPath, err := checkSpecWorkspace()
	if err != nil {
		printWorkspaceError()
		return
	}

	filePath := filepath.Join(specPath, maintenanceDir, slug+".md")
	if !fileExists(filePath) {
		printError(fmt.Sprintf("Maintenance item '%s' does not exist", slug))
		return
	}

	state, err := loadState(specPath)
	if err != nil {
		printError(fmt.Sprintf("Failed to load state: %v", err))
		return
	}

	reqs, err := parseMaintenanceFile(filePath, state, slug)
	if err != nil {
		printError(fmt.Sprintf("Failed to parse maintenance file: %v", err))
		return
	}

	dueReqs := []MaintenanceRequirement{}
	for _, req := range reqs {
		if req.Due {
			dueReqs = append(dueReqs, req)
		}
	}

	fmt.Println()
	fmt.Println(boldStyle.Render(fmt.Sprintf("Due Requirements: %s", slug)))
	fmt.Println()

	if len(dueReqs) == 0 {
		printDim("No requirements due")
		return
	}

	for _, req := range dueReqs {
		fmt.Printf("  %s  %s\n", successStyle.Render("["+req.ID+"]"), req.Text)
		if req.Freq != "" {
			fmt.Printf("      %s\n", dimStyle.Render("freq: "+req.Freq))
		}
		if req.LastActioned != "" {
			fmt.Printf("      %s\n", dimStyle.Render("last: "+req.LastActioned))
		}
		fmt.Println()
	}
}

func runMaintenanceActioned(cmd *cobra.Command, args []string) {
	slug := args[0]
	id := args[1]

	specPath, err := checkSpecWorkspace()
	if err != nil {
		printWorkspaceError()
		return
	}

	filePath := filepath.Join(specPath, maintenanceDir, slug+".md")
	if !fileExists(filePath) {
		printError(fmt.Sprintf("Maintenance item '%s' does not exist", slug))
		return
	}

	state, err := loadState(specPath)
	if err != nil {
		printError(fmt.Sprintf("Failed to load state: %v", err))
		return
	}

	// Parse file to validate ID exists
	reqs, err := parseMaintenanceFile(filePath, state, slug)
	if err != nil {
		printError(fmt.Sprintf("Failed to parse maintenance file: %v", err))
		return
	}

	found := false
	var reqText string
	for _, req := range reqs {
		if req.ID == id {
			found = true
			reqText = req.Text
			break
		}
	}

	if !found {
		printError(fmt.Sprintf("Requirement ID '%s' not found in maintenance item '%s'", id, slug))
		return
	}

	// Update state
	if state.Maintenance == nil {
		state.Maintenance = make(map[string]map[string]MaintenanceState)
	}
	if state.Maintenance[slug] == nil {
		state.Maintenance[slug] = make(map[string]MaintenanceState)
	}

	state.Maintenance[slug][id] = MaintenanceState{
		LastActioned: time.Now().Format(time.RFC3339),
	}

	if err := saveState(specPath, state); err != nil {
		printError(fmt.Sprintf("Failed to save state: %v", err))
		return
	}

	printSuccess(fmt.Sprintf("Marked '%s' as actioned", id))
	printDim(reqText)
}

func runMaintenanceRemove(cmd *cobra.Command, args []string) {
	slug := args[0]

	specPath, err := checkSpecWorkspace()
	if err != nil {
		printWorkspaceError()
		return
	}

	filePath := filepath.Join(specPath, maintenanceDir, slug+".md")
	if !fileExists(filePath) {
		printError(fmt.Sprintf("Maintenance item '%s' does not exist", slug))
		return
	}

	if err := os.Remove(filePath); err != nil {
		printError(fmt.Sprintf("Failed to remove maintenance item: %v", err))
		return
	}

	// Clean up state
	state, err := loadState(specPath)
	if err == nil && state.Maintenance != nil {
		delete(state.Maintenance, slug)
		_ = saveState(specPath, state) // Ignore error, file is already deleted
	}

	printSuccess(fmt.Sprintf("Removed maintenance item '%s'", slug))
}
