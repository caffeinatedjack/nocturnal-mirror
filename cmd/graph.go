package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var graphFormat string

var specProposalGraphCmd = &cobra.Command{
	Use:               "graph [slug]",
	Short:             "Show proposal dependency graph",
	Args:              cobra.MaximumNArgs(1),
	Run:               runSpecProposalGraph,
	ValidArgsFunction: completeProposalNames,
}

func init() {
	specProposalGraphCmd.Long = helpText("spec-proposal-graph")
	specProposalGraphCmd.Flags().StringVarP(&graphFormat, "format", "f", "ascii", "Output format: ascii or dot")
	specProposalCmd.AddCommand(specProposalGraphCmd)
}

// ProposalNode represents a proposal in the dependency graph.
type ProposalNode struct {
	Slug         string
	Dependencies []string
	IsCompleted  bool
	IsActive     bool
}

func runSpecProposalGraph(cmd *cobra.Command, args []string) {
	specPath, err := checkSpecWorkspace()
	if err != nil {
		printWorkspaceError()
		return
	}

	nodes, err := buildDependencyGraph(specPath)
	if err != nil {
		printError(fmt.Sprintf("Failed to build graph: %v", err))
		return
	}

	if len(nodes) == 0 {
		printDim("No proposals found")
		return
	}

	// Filter to single proposal if specified
	var filterSlug string
	if len(args) > 0 {
		filterSlug = args[0]
		if _, exists := nodes[filterSlug]; !exists {
			printError(fmt.Sprintf("Proposal '%s' not found", filterSlug))
			return
		}
	}

	// Detect circular dependencies
	cycles := detectCycles(nodes)
	if len(cycles) > 0 {
		printWarning("Circular dependencies detected:")
		for _, cycle := range cycles {
			fmt.Printf("  %s\n", warningStyle.Render(strings.Join(cycle, " -> ")))
		}
		fmt.Println()
	}

	switch graphFormat {
	case "dot":
		fmt.Print(renderDotGraph(nodes, filterSlug))
	case "ascii":
		renderAsciiGraph(nodes, filterSlug)
	default:
		printError(fmt.Sprintf("Unknown format: %s (use 'ascii' or 'dot')", graphFormat))
	}
}

func buildDependencyGraph(specPath string) (map[string]*ProposalNode, error) {
	nodes := make(map[string]*ProposalNode)

	// Load state for active proposal info
	state, err := loadState(specPath)
	if err != nil {
		return nil, err
	}

	// Add completed specs as nodes (they satisfy dependencies)
	sectionPath := filepath.Join(specPath, sectionDir)
	sectionFiles, err := listMarkdownFiles(sectionPath)
	if err == nil {
		for _, filename := range sectionFiles {
			slug := strings.TrimSuffix(filename, ".md")
			nodes[slug] = &ProposalNode{
				Slug:        slug,
				IsCompleted: true,
			}
		}
	}

	// Add proposals
	proposalsPath := filepath.Join(specPath, proposalDir)
	entries, err := os.ReadDir(proposalsPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		slug := entry.Name()
		proposalPath := filepath.Join(proposalsPath, slug)
		deps, _ := getProposalDependencies(proposalPath)

		nodes[slug] = &ProposalNode{
			Slug:         slug,
			Dependencies: deps,
			IsCompleted:  false,
			IsActive:     state.isProposalActive(slug),
		}
	}

	return nodes, nil
}

func detectCycles(nodes map[string]*ProposalNode) [][]string {
	var cycles [][]string
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{}

	var dfs func(slug string) bool
	dfs = func(slug string) bool {
		visited[slug] = true
		recStack[slug] = true
		path = append(path, slug)

		node, exists := nodes[slug]
		if exists {
			for _, dep := range node.Dependencies {
				if !visited[dep] {
					if dfs(dep) {
						return true
					}
				} else if recStack[dep] {
					// Found cycle - extract it
					cycleStart := -1
					for i, s := range path {
						if s == dep {
							cycleStart = i
							break
						}
					}
					if cycleStart >= 0 {
						cycle := append([]string{}, path[cycleStart:]...)
						cycle = append(cycle, dep)
						cycles = append(cycles, cycle)
					}
					return true
				}
			}
		}

		path = path[:len(path)-1]
		recStack[slug] = false
		return false
	}

	for slug := range nodes {
		if !visited[slug] {
			dfs(slug)
		}
	}

	return cycles
}

func renderDotGraph(nodes map[string]*ProposalNode, filterSlug string) string {
	var buf strings.Builder
	buf.WriteString("digraph dependencies {\n")
	buf.WriteString("  rankdir=BT;\n")
	buf.WriteString("  node [shape=box];\n\n")

	// Collect relevant nodes
	relevantNodes := nodes
	if filterSlug != "" {
		relevantNodes = getRelevantNodes(nodes, filterSlug)
	}

	// Define node styles
	for slug, node := range relevantNodes {
		var style string
		if node.IsCompleted {
			style = "style=filled,fillcolor=lightgreen"
		} else if node.IsActive {
			style = "style=filled,fillcolor=lightblue"
		} else {
			style = "style=solid"
		}
		buf.WriteString(fmt.Sprintf("  \"%s\" [%s];\n", slug, style))
	}

	buf.WriteString("\n")

	// Define edges
	for slug, node := range relevantNodes {
		for _, dep := range node.Dependencies {
			buf.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\";\n", slug, dep))
		}
	}

	buf.WriteString("}\n")
	return buf.String()
}

func renderAsciiGraph(nodes map[string]*ProposalNode, filterSlug string) {
	fmt.Println()
	fmt.Println(boldStyle.Render("Dependency Graph"))
	fmt.Println()

	// Legend
	fmt.Printf("  %s completed  %s active  %s pending\n",
		successStyle.Render("*"),
		infoStyle.Render("*"),
		dimStyle.Render("*"))
	fmt.Println()

	// Collect relevant nodes
	relevantNodes := nodes
	if filterSlug != "" {
		relevantNodes = getRelevantNodes(nodes, filterSlug)
	}

	// Sort nodes by name
	slugs := make([]string, 0, len(relevantNodes))
	for slug := range relevantNodes {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)

	// Find nodes with no dependents (roots for display)
	dependents := make(map[string][]string)
	for slug, node := range relevantNodes {
		for _, dep := range node.Dependencies {
			dependents[dep] = append(dependents[dep], slug)
		}
	}

	// Print each node with its relationships
	for _, slug := range slugs {
		node := relevantNodes[slug]

		// Style the node name
		var styledName string
		if node.IsCompleted {
			styledName = successStyle.Render(slug)
		} else if node.IsActive {
			styledName = infoStyle.Render(slug)
		} else {
			styledName = slug
		}

		fmt.Printf("  %s\n", styledName)

		// Show dependencies (what this depends on)
		if len(node.Dependencies) > 0 {
			for i, dep := range node.Dependencies {
				prefix := "├──"
				if i == len(node.Dependencies)-1 && len(dependents[slug]) == 0 {
					prefix = "└──"
				}
				depNode, exists := nodes[dep]
				var depStatus string
				if !exists {
					depStatus = errorStyle.Render("(missing)")
				} else if depNode.IsCompleted {
					depStatus = successStyle.Render("(completed)")
				} else {
					depStatus = dimStyle.Render("(pending)")
				}
				fmt.Printf("    %s depends on: %s %s\n", dimStyle.Render(prefix), dep, depStatus)
			}
		}

		// Show dependents (what depends on this)
		if deps, ok := dependents[slug]; ok && len(deps) > 0 {
			for i, dep := range deps {
				prefix := "├──"
				if i == len(deps)-1 {
					prefix = "└──"
				}
				fmt.Printf("    %s blocks: %s\n", dimStyle.Render(prefix), dep)
			}
		}

		fmt.Println()
	}
}

// getRelevantNodes returns nodes related to the given slug (ancestors and descendants).
func getRelevantNodes(allNodes map[string]*ProposalNode, slug string) map[string]*ProposalNode {
	relevant := make(map[string]*ProposalNode)
	visited := make(map[string]bool)

	// Add ancestors (dependencies)
	var addAncestors func(s string)
	addAncestors = func(s string) {
		if visited[s] {
			return
		}
		visited[s] = true
		if node, exists := allNodes[s]; exists {
			relevant[s] = node
			for _, dep := range node.Dependencies {
				addAncestors(dep)
			}
		}
	}

	// Add descendants (dependents)
	var addDescendants func(s string)
	addDescendants = func(s string) {
		for otherSlug, node := range allNodes {
			if visited[otherSlug] {
				continue
			}
			for _, dep := range node.Dependencies {
				if dep == s {
					visited[otherSlug] = true
					relevant[otherSlug] = node
					addDescendants(otherSlug)
					break
				}
			}
		}
	}

	addAncestors(slug)
	visited = make(map[string]bool) // Reset for descendants
	visited[slug] = true
	addDescendants(slug)

	return relevant
}
