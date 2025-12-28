package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/nocturnal/pkg/ui"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Manage documentation in ~/.docs directory",
	Long: `Manage documentation stored in the ~/.docs directory.

The ~/.docs directory can contain multiple documentation files. Each file
contains documentation components separated by '---' with headers starting
with '# component'.

Example file format:
    ---
    # my-component
    This is the content for my component.
    It can span multiple lines.

    ---
    # another-component
    More content here.

Commands:
    list      List all documentation components from all files
    search    Search documentation by component name
`,
}

var (
	docsPath string
)

func init() {
	rootCmd.AddCommand(docsCmd)

	// Set default docs path
	home, _ := os.UserHomeDir()
	docsPath = filepath.Join(home, ".docs")

	docsCmd.AddCommand(docsListCmd)
	docsCmd.AddCommand(docsSearchCmd)
}

// DocComponent represents a single documentation component.
type DocComponent struct {
	Name    string
	Content string
	Source  string // The file this component came from
}

// loadDocs reads and parses all documentation files in the docs directory.
func loadDocs() ([]*DocComponent, error) {
	// Check if docs directory exists
	info, err := os.Stat(docsPath)
	if os.IsNotExist(err) {
		return []*DocComponent{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to access docs directory: %w", err)
	}

	// Ensure it's a directory
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", docsPath)
	}

	// Read all files in the directory
	entries, err := os.ReadDir(docsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read docs directory: %w", err)
	}

	var components []*DocComponent

	// Process each file
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip directories
		}

		filePath := filepath.Join(docsPath, entry.Name())
		fileComponents, err := parseDocFile(filePath)
		if err != nil {
			// Log error but continue processing other files
			ui.Error(fmt.Sprintf("Error reading %s: %v", entry.Name(), err))
			continue
		}

		components = append(components, fileComponents...)
	}

	return components, nil
}

// parseDocFile parses a single documentation file and returns its components.
func parseDocFile(filePath string) ([]*DocComponent, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	sourceFile := filepath.Base(filePath)
	var components []*DocComponent
	var currentContent strings.Builder
	var currentName string
	inContent := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Check for separator
		if strings.TrimSpace(line) == "---" {
			// Save previous component if exists
			if currentName != "" {
				components = append(components, &DocComponent{
					Name:    currentName,
					Content: strings.TrimSpace(currentContent.String()),
					Source:  sourceFile,
				})
			}
			currentName = ""
			currentContent.Reset()
			inContent = false
			continue
		}

		// Check for header
		if strings.HasPrefix(strings.TrimSpace(line), "# ") {
			// This is a component name header
			currentName = strings.TrimSpace(strings.TrimPrefix(line, "#"))
			inContent = true
			continue
		}

		// Add to content if we're in a component
		if inContent {
			if currentContent.Len() > 0 {
				currentContent.WriteString("\n")
			}
			currentContent.WriteString(line)
		}
	}

	// Don't forget the last component
	if currentName != "" {
		components = append(components, &DocComponent{
			Name:    currentName,
			Content: strings.TrimSpace(currentContent.String()),
			Source:  sourceFile,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return components, nil
}

// docsListCmd lists all documentation components.
var docsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all documentation components",
	Long: `List all documentation components.

Shows the name and a preview of each documentation component.

Example:
    nocturnal docs list`,
	Run: runDocsList,
}

func runDocsList(cmd *cobra.Command, args []string) {
	components, err := loadDocs()
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to load docs: %v", err))
		return
	}

	if len(components) == 0 {
		ui.PrintDim("No documentation found")
		if _, err := os.Stat(docsPath); os.IsNotExist(err) {
			fmt.Println()
			ui.Info(fmt.Sprintf("Create %s directory and add documentation files", docsPath))
		}
		return
	}

	fmt.Println()
	fmt.Println(ui.BoldStyle.Render(fmt.Sprintf("Found %d component(s)", len(components))))
	fmt.Println()

	for _, comp := range components {
		// Show component name
		fmt.Printf("%s\n", ui.TopicStyle.Render("# "+comp.Name))
		// Show source file
		fmt.Printf("  %s\n", ui.DimStyle.Render("from "+comp.Source))

		// Show preview (first line or first 60 chars)
		preview := comp.Content
		if idx := strings.Index(preview, "\n"); idx > 0 {
			preview = preview[:idx]
		}
		if len(preview) > 60 {
			preview = preview[:57] + "..."
		}

		if preview != "" {
			fmt.Printf("  %s\n", ui.DimStyle.Render(preview))
		}
		fmt.Println()
	}
}

// docsSearchCmd searches documentation by component name.
var docsSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search documentation by component name",
	Long: `Search documentation by component name.

Searches for components whose names contain the query string and displays
the full content of matching components.

Example:
    nocturnal docs search "component"
    nocturnal docs search "api"`,
	Args: cobra.ExactArgs(1),
	Run:  runDocsSearch,
}

func runDocsSearch(cmd *cobra.Command, args []string) {
	query := strings.ToLower(args[0])

	components, err := loadDocs()
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to load docs: %v", err))
		return
	}

	if len(components) == 0 {
		ui.PrintDim("No documentation found")
		return
	}

	// Search for matches
	var matches []*DocComponent
	for _, comp := range components {
		if strings.Contains(strings.ToLower(comp.Name), query) {
			matches = append(matches, comp)
		}
	}

	if len(matches) == 0 {
		ui.PrintDim(fmt.Sprintf("No components found matching '%s'", args[0]))
		fmt.Println()
		ui.PrintDim("Use 'nocturnal docs list' to see all available components")
		return
	}

	fmt.Println()
	fmt.Println(ui.BoldStyle.Render(fmt.Sprintf("Found %d result(s)", len(matches))))
	fmt.Println()

	for _, comp := range matches {
		// Print component name
		fmt.Printf("%s\n", ui.TopicStyle.Render("# "+comp.Name))
		// Print source file
		fmt.Printf("  %s\n", ui.DimStyle.Render("from "+comp.Source))

		// Print full content with proper indentation and formatting
		lines := strings.Split(comp.Content, "\n")
		for _, line := range lines {
			fmt.Printf("  %s\n", line)
		}
		fmt.Println()
	}
}
