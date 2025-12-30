package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Manage third-party documentation in spec/third",
}

var (
	docsPath string
)

func init() {
	docsPath = filepath.Join(getSpecPath(), "third")

	docsCmd.Long = helpText("agent-docs")
	docsListCmd.Long = helpText("agent-docs-list")
	docsSearchCmd.Long = helpText("agent-docs-search")

	docsCmd.AddCommand(docsListCmd)
	docsCmd.AddCommand(docsSearchCmd)
}

// RegisterDocsCommand adds the docs subcommand to a parent command.
func RegisterDocsCommand(parent *cobra.Command) {
	parent.AddCommand(docsCmd)
}

// DocComponent represents a named section from a documentation file.
type DocComponent struct {
	Name    string
	Content string
	Source  string
}

// formatDocsListOutput formats components as a list with previews.
func formatDocsListOutput(components []*DocComponent) string {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("Found %d component(s)\n\n", len(components)))

	for _, comp := range components {
		buf.WriteString(fmt.Sprintf("# %s\n", comp.Name))
		buf.WriteString(fmt.Sprintf("  from %s\n", comp.Source))

		if preview := getContentPreview(comp.Content); preview != "" {
			buf.WriteString(fmt.Sprintf("  %s\n", preview))
		}
		buf.WriteString("\n")
	}

	return buf.String()
}

// searchDocs filters components by name (case-insensitive).
func searchDocs(components []*DocComponent, query string) []*DocComponent {
	queryLower := strings.ToLower(query)
	var matches []*DocComponent
	for _, comp := range components {
		if strings.Contains(strings.ToLower(comp.Name), queryLower) {
			matches = append(matches, comp)
		}
	}
	return matches
}

// formatDocsSearchOutput formats matched components with full content.
func formatDocsSearchOutput(matches []*DocComponent) string {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("Found %d result(s)\n\n", len(matches)))

	for _, comp := range matches {
		buf.WriteString(fmt.Sprintf("# %s\n", comp.Name))
		buf.WriteString(fmt.Sprintf("  from %s\n\n", comp.Source))
		buf.WriteString(comp.Content)
		buf.WriteString("\n\n")
	}

	return buf.String()
}

// loadDocs reads all documentation files from spec/third/.
func loadDocs() ([]*DocComponent, error) {
	info, err := os.Stat(docsPath)
	if os.IsNotExist(err) {
		return []*DocComponent{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to access docs directory: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", docsPath)
	}

	entries, err := os.ReadDir(docsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read docs directory: %w", err)
	}

	var components []*DocComponent

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(docsPath, entry.Name())
		fileComponents, err := parseDocFile(filePath)
		if err != nil {
			printError(fmt.Sprintf("Error reading %s: %v", entry.Name(), err))
			continue
		}

		components = append(components, fileComponents...)
	}

	return components, nil
}

// parseDocFile extracts components from a file. Sections are delimited by ---.
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

		if strings.TrimSpace(line) == "---" {
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

		if strings.HasPrefix(strings.TrimSpace(line), "# ") {
			currentName = strings.TrimSpace(strings.TrimPrefix(line, "#"))
			inContent = true
			continue
		}

		if inContent {
			if currentContent.Len() > 0 {
				currentContent.WriteString("\n")
			}
			currentContent.WriteString(line)
		}
	}

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

var docsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all documentation components",
	Run:   runDocsList,
}

func runDocsList(cmd *cobra.Command, args []string) {
	components, err := loadDocs()
	if err != nil {
		printError(fmt.Sprintf("Failed to load docs: %v", err))
		return
	}

	if len(components) == 0 {
		printDim("No documentation found")
		if !fileExists(docsPath) {
			fmt.Println()
			printInfo(fmt.Sprintf("Create %s directory and add documentation files", docsPath))
		}
		return
	}

	fmt.Println()
	fmt.Println(boldStyle.Render(fmt.Sprintf("Found %d component(s)", len(components))))
	fmt.Println()

	for _, comp := range components {
		fmt.Printf("%s\n", topicStyle.Render("# "+comp.Name))
		fmt.Printf("  %s\n", dimStyle.Render("from "+comp.Source))

		if preview := getContentPreview(comp.Content); preview != "" {
			fmt.Printf("  %s\n", dimStyle.Render(preview))
		}
		fmt.Println()
	}
}

var docsSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search documentation by component name",
	Args:  cobra.ExactArgs(1),
	Run:   runDocsSearch,
}

func runDocsSearch(cmd *cobra.Command, args []string) {
	components, err := loadDocs()
	if err != nil {
		printError(fmt.Sprintf("Failed to load docs: %v", err))
		return
	}

	if len(components) == 0 {
		printDim("No documentation found")
		return
	}

	matches := searchDocs(components, args[0])
	if len(matches) == 0 {
		printDim(fmt.Sprintf("No components found matching '%s'", args[0]))
		fmt.Println()
		printDim("Use 'nocturnal docs list' to see all available components")
		return
	}

	fmt.Println()
	fmt.Println(boldStyle.Render(fmt.Sprintf("Found %d result(s)", len(matches))))
	fmt.Println()

	for _, comp := range matches {
		fmt.Printf("%s\n", topicStyle.Render("# "+comp.Name))
		fmt.Printf("  %s\n", dimStyle.Render("from "+comp.Source))

		for _, line := range strings.Split(comp.Content, "\n") {
			fmt.Printf("  %s\n", line)
		}
		fmt.Println()
	}
}
