package cmd

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

//go:embed templates
var templateFS embed.FS

// helpText loads a help text file from the embedded templates.
func helpText(name string) string {
	content, err := templateFS.ReadFile("templates/help/" + name + ".txt")
	if err != nil {
		return ""
	}
	return string(content)
}

// renderTemplate executes a Go template with the given data and returns the result.
func renderTemplate(templatePath string, data any) (string, error) {
	content, err := templateFS.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template %s: %w", templatePath, err)
	}

	tmpl, err := template.New(filepath.Base(templatePath)).Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", templatePath, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", templatePath, err)
	}

	return buf.String(), nil
}

// readTemplate reads a template file without executing it.
func readTemplate(templatePath string) (string, error) {
	content, err := templateFS.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template %s: %w", templatePath, err)
	}
	return string(content), nil
}

var specCmd = &cobra.Command{
	Use:   "spec",
	Short: "Manage project specifications",
}

var specViewCmd = &cobra.Command{
	Use:   "view",
	Short: "View specification workspace overview",
	Run:   runSpecView,
}

var specInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a specification workspace",
	Run:   runSpecInit,
}

var specProposalCmd = &cobra.Command{
	Use:   "proposal",
	Short: "Manage proposals",
}

var specProposalAddCmd = &cobra.Command{
	Use:   "add <change-slug>",
	Short: "Create a new proposal",
	Args:  cobra.ExactArgs(1),
	Run:   runSpecProposalAdd,
}

var forceRemove bool

var specProposalRemoveCmd = &cobra.Command{
	Use:               "remove <change-slug>",
	Short:             "Remove a proposal",
	Args:              cobra.ExactArgs(1),
	Run:               runSpecProposalRemove,
	ValidArgsFunction: completeProposalNames,
}

var specProposalActivateCmd = &cobra.Command{
	Use:               "activate <change-slug>",
	Short:             "Activate a proposal",
	Args:              cobra.ExactArgs(1),
	Run:               runSpecProposalActivate,
	ValidArgsFunction: completeProposalNames,
}

var specProposalDeactivateCmd = &cobra.Command{
	Use:   "deactivate",
	Short: "Deactivate the current proposal",
	Args:  cobra.NoArgs,
	Run:   runSpecProposalDeactivate,
}

var specProposalCompleteCmd = &cobra.Command{
	Use:               "complete <change-slug>",
	Short:             "Complete and promote a proposal",
	Args:              cobra.ExactArgs(1),
	Run:               runSpecProposalComplete,
	ValidArgsFunction: completeProposalNames,
}

var specProposalValidateCmd = &cobra.Command{
	Use:               "validate <change-slug>",
	Short:             "Validate proposal documents against guidelines",
	Args:              cobra.ExactArgs(1),
	Run:               runSpecProposalValidate,
	ValidArgsFunction: completeProposalNames,
}

var specProposalListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all proposals with status and progress",
	Run:   runSpecProposalList,
}

var specProposalAbandonCmd = &cobra.Command{
	Use:               "abandon <change-slug>",
	Short:             "Abandon a proposal and archive it without promoting",
	Args:              cobra.ExactArgs(1),
	Run:               runSpecProposalAbandon,
	ValidArgsFunction: completeProposalNames,
}

var agentSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show a complete project summary for AI context",
	Run:   runAgentSummary,
}

var specRuleCmd = &cobra.Command{
	Use:   "rule",
	Short: "Manage rules",
}

var specRuleAddCmd = &cobra.Command{
	Use:   "add <rule-name>",
	Short: "Add a new rule",
	Args:  cobra.ExactArgs(1),
	Run:   runSpecRuleAdd,
}

var specRuleShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show all rules",
	Run:   runSpecRuleShow,
}

var agentCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the currently active proposal",
	Run:   runAgentCurrent,
}

var agentProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Show project rules and design",
	Run:   runAgentProject,
}

var agentSpecificationsCmd = &cobra.Command{
	Use:     "specifications",
	Aliases: []string{"specs"},
	Short:   "Show completed specifications",
	Run:     runAgentSpecifications,
}

func init() {
	specCmd.Long = helpText("spec")
	specViewCmd.Long = helpText("spec-view")
	specInitCmd.Long = helpText("spec-init")
	specProposalCmd.Long = helpText("spec-proposal")
	specProposalAddCmd.Long = helpText("spec-proposal-add")
	specProposalRemoveCmd.Long = helpText("spec-proposal-remove")
	specProposalActivateCmd.Long = helpText("spec-proposal-activate")
	specProposalDeactivateCmd.Long = helpText("spec-proposal-deactivate")
	specProposalCompleteCmd.Long = helpText("spec-proposal-complete")
	specProposalValidateCmd.Long = helpText("spec-proposal-validate")
	specProposalListCmd.Long = helpText("spec-proposal-list")
	specProposalAbandonCmd.Long = helpText("spec-proposal-abandon")
	specRuleCmd.Long = helpText("spec-rule")
	specRuleAddCmd.Long = helpText("spec-rule-add")
	specRuleShowCmd.Long = helpText("spec-rule-show")
	agentCurrentCmd.Long = helpText("agent-current")
	agentProjectCmd.Long = helpText("agent-project")
	agentSpecificationsCmd.Long = helpText("agent-specs")
	agentSummaryCmd.Long = helpText("agent-summary")

	rootCmd.AddCommand(specCmd)

	specCmd.AddCommand(specViewCmd)
	specCmd.AddCommand(specInitCmd)
	specCmd.AddCommand(specProposalCmd)
	specCmd.AddCommand(specRuleCmd)

	specProposalCmd.AddCommand(specProposalAddCmd)
	specProposalCmd.AddCommand(specProposalRemoveCmd)
	specProposalCmd.AddCommand(specProposalActivateCmd)
	specProposalCmd.AddCommand(specProposalDeactivateCmd)
	specProposalCmd.AddCommand(specProposalCompleteCmd)
	specProposalCmd.AddCommand(specProposalValidateCmd)
	specProposalCmd.AddCommand(specProposalListCmd)
	specProposalCmd.AddCommand(specProposalAbandonCmd)

	specProposalRemoveCmd.Flags().BoolVarP(&forceRemove, "force", "f", false, "Force removal even if proposal is active")

	specRuleCmd.AddCommand(specRuleAddCmd)
	specRuleCmd.AddCommand(specRuleShowCmd)

	agentCmd.AddCommand(agentCurrentCmd)
	agentCmd.AddCommand(agentProjectCmd)
	agentCmd.AddCommand(agentSpecificationsCmd)
	agentCmd.AddCommand(agentSummaryCmd)
}

var proposalDocs = []struct {
	Name string
	File string
}{
	{"Specification", "specification.md"},
	{"Design", "design.md"},
	{"Implementation", "implementation.md"},
}

// readProposalDocs reads all proposal documents (spec, design, implementation)
func readProposalDocs(proposalPath string) (string, error) {
	return readProposalDocsFiltered(proposalPath, nil)
}

// readProposalDocsFiltered reads proposal documents, optionally filtering to specific files
// If files is nil or empty, all documents are read
func readProposalDocsFiltered(proposalPath string, files []string) (string, error) {
	var buf bytes.Buffer
	first := true

	for _, doc := range proposalDocs {
		// Skip if filtering and file not in list
		if len(files) > 0 && !contains(files, doc.File) {
			continue
		}

		filePath := filepath.Join(proposalPath, doc.File)
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		if !first {
			buf.WriteString("\n---\n\n")
		}
		first = false

		buf.WriteString(fmt.Sprintf("## %s\n\n", doc.Name))
		buf.Write(content)
	}

	return buf.String(), nil
}

// contains checks if a string slice contains a value
func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

// readRulesAndProject concatenates all rules and project.md into a single string.
func readRulesAndProject(specPath string) (string, error) {
	var buf bytes.Buffer
	hasOutput := false

	rulesDirPath := filepath.Join(specPath, ruleDir)
	ruleFiles, err := listMarkdownFiles(rulesDirPath)
	if err == nil && len(ruleFiles) > 0 {
		buf.WriteString("# Rules\n\n")

		for _, filename := range ruleFiles {
			filePath := filepath.Join(rulesDirPath, filename)
			content, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}
			buf.Write(content)
			buf.WriteString("\n")
		}
		hasOutput = true
	}

	projectPath := filepath.Join(specPath, projectFile)
	if content, err := os.ReadFile(projectPath); err == nil {
		if hasOutput {
			buf.WriteString("---\n\n")
		}
		buf.WriteString("# Project Design\n\n")
		buf.Write(content)
		buf.WriteString("\n")
		hasOutput = true
	}

	if !hasOutput {
		return "", nil
	}

	return buf.String(), nil
}

// readSpecifications concatenates all completed specifications from section/.
func readSpecifications(specPath string) (string, error) {
	sectionDirPath := filepath.Join(specPath, sectionDir)
	sectionFiles, err := listMarkdownFiles(sectionDirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read section directory: %w", err)
	}

	if len(sectionFiles) == 0 {
		return "", nil
	}

	var buf bytes.Buffer
	buf.WriteString("# Specifications\n\n")

	for i, filename := range sectionFiles {
		filePath := filepath.Join(sectionDirPath, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		if i > 0 {
			buf.WriteString("\n---\n\n")
		}

		sectionName := strings.TrimSuffix(filename, ".md")
		buf.WriteString(fmt.Sprintf("## %s\n\n", sectionName))
		buf.Write(content)
	}

	return buf.String(), nil
}

// completeProposalNames provides shell completion for proposal names.
func completeProposalNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	specPath := getSpecPath()
	proposalsPath := filepath.Join(specPath, proposalDir)

	entries, err := os.ReadDir(proposalsPath)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var proposals []string
	for _, entry := range entries {
		if entry.IsDir() {
			proposals = append(proposals, entry.Name())
		}
	}

	return proposals, cobra.ShellCompDirectiveNoFileComp
}

// countRequirements counts lines containing MUST or SHALL keywords.
func countRequirements(content string) int {
	count := 0
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		upper := strings.ToUpper(line)
		if strings.Contains(upper, "MUST") || strings.Contains(upper, "SHALL") {
			count++
		}
	}
	return count
}

// getProposalProgress counts task checkboxes in implementation.md.
func getProposalProgress(proposalPath string) (total int, completed int) {
	implPath := filepath.Join(proposalPath, "implementation.md")
	content, err := os.ReadFile(implPath)
	if err != nil {
		return 0, 0
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- [ ]") {
			total++
		} else if strings.HasPrefix(trimmed, "- [x]") || strings.HasPrefix(trimmed, "- [X]") {
			total++
			completed++
		}
	}
	return total, completed
}

func runSpecView(cmd *cobra.Command, args []string) {
	specPath, err := checkSpecWorkspace()
	if err != nil {
		printWorkspaceError()
		return
	}

	fmt.Println()

	sectionDirPath := filepath.Join(specPath, sectionDir)
	sectionFiles, err := listMarkdownFiles(sectionDirPath)
	if err != nil && !os.IsNotExist(err) {
		printError(fmt.Sprintf("Failed to read section directory: %v", err))
		return
	}

	fmt.Println(boldStyle.Render("Specifications"))
	fmt.Println()

	if len(sectionFiles) == 0 {
		printDim("  No completed specifications")
	} else {
		for _, filename := range sectionFiles {
			filePath := filepath.Join(sectionDirPath, filename)
			content, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}
			name := strings.TrimSuffix(filename, ".md")
			reqCount := countRequirements(string(content))
			reqLabel := "requirements"
			if reqCount == 1 {
				reqLabel = "requirement"
			}
			fmt.Printf("  %s  %s\n", name, dimStyle.Render(fmt.Sprintf("(%d %s)", reqCount, reqLabel)))
		}
	}

	fmt.Println()

	fmt.Println(boldStyle.Render("Active Proposal"))
	fmt.Println()

	slug, proposalPath, err := getActiveProposal(specPath)
	if err != nil {
		printWarning(fmt.Sprintf("  %s", err.Error()))
	} else if slug == "" {
		printDim("  No active proposal")
	} else {
		total, completed := getProposalProgress(proposalPath)
		if total > 0 {
			percentage := (completed * 100) / total
			progressBar := renderProgressBar(completed, total, 20)
			fmt.Printf("  %s  %s %s\n", infoStyle.Render(slug), progressBar, dimStyle.Render(fmt.Sprintf("%d%% (%d/%d tasks)", percentage, completed, total)))
		} else {
			fmt.Printf("  %s  %s\n", infoStyle.Render(slug), dimStyle.Render("(no tasks)"))
		}
		// Show dependencies for active proposal
		if deps, _ := getProposalDependencies(proposalPath); len(deps) > 0 {
			fmt.Printf("  %s %s\n", dimStyle.Render("depends on:"), strings.Join(deps, ", "))
		}
	}

	fmt.Println()

	fmt.Println(boldStyle.Render("Other Proposals"))
	fmt.Println()

	proposalsPath := filepath.Join(specPath, proposalDir)
	entries, err := os.ReadDir(proposalsPath)
	if err != nil && !os.IsNotExist(err) {
		printError(fmt.Sprintf("Failed to read proposals directory: %v", err))
		return
	}

	otherProposals := []string{}
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != slug {
			otherProposals = append(otherProposals, entry.Name())
		}
	}

	if len(otherProposals) == 0 {
		printDim("  No other proposals")
	} else {
		for _, name := range otherProposals {
			propPath := filepath.Join(proposalsPath, name)
			total, completed := getProposalProgress(propPath)
			deps, _ := getProposalDependencies(propPath)

			var parts []string
			if total > 0 {
				percentage := (completed * 100) / total
				parts = append(parts, fmt.Sprintf("%d%% complete", percentage))
			}
			if len(deps) > 0 {
				parts = append(parts, fmt.Sprintf("depends on: %s", strings.Join(deps, ", ")))
			}

			if len(parts) > 0 {
				fmt.Printf("  %s  %s\n", name, dimStyle.Render("("+strings.Join(parts, ", ")+")"))
			} else {
				fmt.Printf("  %s\n", name)
			}
		}
	}

	fmt.Println()
}

// renderProgressBar creates a visual progress bar using block characters.
func renderProgressBar(completed, total, width int) string {
	if total == 0 {
		return dimStyle.Render("[" + strings.Repeat("-", width) + "]")
	}

	filled := (completed * width) / total
	empty := width - filled

	bar := successStyle.Render(strings.Repeat("█", filled)) + dimStyle.Render(strings.Repeat("░", empty))
	return "[" + bar + "]"
}

func runSpecInit(cmd *cobra.Command, args []string) {
	specPath := getSpecPath()

	if _, err := os.Stat(specPath); err == nil {
		printError("Specification workspace already exists")
		printDim("Remove specification/ directory first if you want to reinitialize")
		return
	}

	dirs := []string{
		specPath,
		filepath.Join(specPath, ruleDir),
		filepath.Join(specPath, proposalDir),
		filepath.Join(specPath, archiveDir),
		filepath.Join(specPath, sectionDir),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			printError(fmt.Sprintf("Failed to create directory %s: %v", dir, err))
			return
		}
	}

	templateFiles := []struct {
		template string
		filename string
	}{
		{"templates/project.md", "project.md"},
		{"templates/AGENTS.md", "AGENTS.md"},
		{"templates/specification guidelines.md", "specification guidelines.md"},
		{"templates/design guidelines.md", "design guidelines.md"},
		{"templates/coding guidelines.md", "coding guidelines.md"},
	}

	for _, tf := range templateFiles {
		content, err := readTemplate(tf.template)
		if err != nil {
			printError(fmt.Sprintf("Failed to read %s template: %v", tf.filename, err))
			return
		}
		filePath := filepath.Join(specPath, tf.filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			printError(fmt.Sprintf("Failed to create %s: %v", tf.filename, err))
			return
		}
	}

	printSuccess("Initialized specification workspace")
	printDim(fmt.Sprintf("Created %s/", specDir))
}

func runSpecProposalAdd(cmd *cobra.Command, args []string) {
	name := args[0]
	slug := nameToSlug(name)

	if slug == "" {
		printError("Invalid proposal name: must contain at least one alphanumeric character")
		return
	}

	specPath, err := checkSpecWorkspace()
	if err != nil {
		printWorkspaceError()
		return
	}
	proposalPath := filepath.Join(specPath, proposalDir, slug)

	if _, err := os.Stat(proposalPath); err == nil {
		printError(fmt.Sprintf("Proposal '%s' already exists", slug))
		return
	}

	if err := os.MkdirAll(proposalPath, 0755); err != nil {
		printError(fmt.Sprintf("Failed to create proposal directory: %v", err))
		return
	}

	data := struct {
		Name string
		Slug string
	}{Name: name, Slug: slug}

	templates := map[string]string{
		"specification.md":  "templates/proposal/specification.md",
		"design.md":         "templates/proposal/design.md",
		"implementation.md": "templates/proposal/implementation.md",
	}

	for filename, templatePath := range templates {
		content, err := renderTemplate(templatePath, data)
		if err != nil {
			printError(fmt.Sprintf("Failed to render %s: %v", filename, err))
			return
		}
		filePath := filepath.Join(proposalPath, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			printError(fmt.Sprintf("Failed to create %s: %v", filename, err))
			return
		}
	}

	printSuccess(fmt.Sprintf("Created proposal '%s'", slug))
	printDim(fmt.Sprintf("Location: %s/", proposalPath))
}

func runSpecProposalRemove(cmd *cobra.Command, args []string) {
	slug := args[0]
	specPath := getSpecPath()
	proposalPath, err := checkProposal(specPath, slug)
	if err != nil {
		printError(err.Error())
		return
	}
	currentPath := filepath.Join(specPath, currentSymlink)

	if !forceRemove {
		if target, err := os.Readlink(currentPath); err == nil {
			activeSlug := filepath.Base(target)
			if activeSlug == slug {
				printError(fmt.Sprintf("Proposal '%s' is currently active", slug))
				printDim("Use --force to remove anyway, or deactivate first")
				return
			}
		}
	}

	if err := os.RemoveAll(proposalPath); err != nil {
		printError(fmt.Sprintf("Failed to remove proposal: %v", err))
		return
	}

	clearActiveProposalIfMatches(specPath, slug)
	printSuccess(fmt.Sprintf("Removed proposal '%s'", slug))
}

func runSpecProposalActivate(cmd *cobra.Command, args []string) {
	slug := args[0]
	specPath, err := checkSpecWorkspace()
	if err != nil {
		printWorkspaceError()
		return
	}

	if _, err := checkProposal(specPath, slug); err != nil {
		printError(err.Error())
		return
	}
	currentPath := filepath.Join(specPath, currentSymlink)

	// Check if any other proposals depend on this one
	dependents, err := findDependentProposals(specPath, slug)
	if err != nil {
		printError(fmt.Sprintf("Failed to check dependencies: %v", err))
		return
	}
	if len(dependents) > 0 {
		printError(fmt.Sprintf("Cannot activate '%s': other proposals depend on it", slug))
		printDim(fmt.Sprintf("Dependent proposals: %s", strings.Join(dependents, ", ")))
		printDim("Complete the dependent proposals first, or remove the dependency")
		return
	}

	if _, err := os.Lstat(currentPath); err == nil {
		if err := os.Remove(currentPath); err != nil {
			printError(fmt.Sprintf("Failed to remove existing symlink: %v", err))
			return
		}
	}

	relTarget := filepath.Join(proposalDir, slug)
	if err := os.Symlink(relTarget, currentPath); err != nil {
		printError(fmt.Sprintf("Failed to create symlink: %v", err))
		return
	}

	printSuccess(fmt.Sprintf("Activated proposal '%s'", slug))
}

func runSpecProposalDeactivate(cmd *cobra.Command, args []string) {
	specPath, err := checkSpecWorkspace()
	if err != nil {
		printWorkspaceError()
		return
	}

	currentPath := filepath.Join(specPath, currentSymlink)

	slug, _, err := getActiveProposal(specPath)
	if err != nil {
		printWarning(err.Error())
		return
	}
	if slug == "" {
		printDim("No active proposal to deactivate")
		return
	}

	if err := os.Remove(currentPath); err != nil {
		printError(fmt.Sprintf("Failed to remove symlink: %v", err))
		return
	}

	printSuccess(fmt.Sprintf("Deactivated proposal '%s'", slug))
}

func runSpecProposalComplete(cmd *cobra.Command, args []string) {
	slug := args[0]
	specPath := getSpecPath()
	proposalPath, err := checkProposal(specPath, slug)
	if err != nil {
		printError(err.Error())
		return
	}

	archivePath := filepath.Join(specPath, archiveDir, slug)
	sectionPath := filepath.Join(specPath, sectionDir)

	specFile := filepath.Join(proposalPath, "specification.md")
	if !fileExists(specFile) {
		printError(fmt.Sprintf("Proposal '%s' is missing specification.md", slug))
		return
	}

	// Archive design and implementation documents
	if err := archiveProposalDocs(proposalPath, archivePath, []string{"design.md", "implementation.md"}); err != nil {
		printError(err.Error())
		return
	}

	// Promote specification to section
	specDst := filepath.Join(sectionPath, slug+".md")
	if err := copyFile(specFile, specDst); err != nil {
		printError(fmt.Sprintf("Failed to promote specification: %v", err))
		return
	}

	if err := os.RemoveAll(proposalPath); err != nil {
		printError(fmt.Sprintf("Failed to remove proposal workspace: %v", err))
		return
	}

	clearActiveProposalIfMatches(specPath, slug)
	printSuccess(fmt.Sprintf("Completed proposal '%s'", slug))
	printDim(fmt.Sprintf("Specification promoted to %s/%s.md", sectionDir, slug))
	printDim(fmt.Sprintf("Design/implementation archived to %s/%s/", archiveDir, slug))
}

func runSpecRuleAdd(cmd *cobra.Command, args []string) {
	ruleName := args[0]
	slug := nameToSlug(ruleName)

	if slug == "" {
		printError("Invalid rule name: must contain at least one alphanumeric character")
		return
	}

	specPath, err := checkSpecWorkspace()
	if err != nil {
		printWorkspaceError()
		return
	}
	rulePath := filepath.Join(specPath, ruleDir, slug+".md")

	if _, err := os.Stat(rulePath); err == nil {
		printError(fmt.Sprintf("Rule '%s' already exists", slug))
		return
	}

	data := struct{ Name string }{Name: ruleName}
	ruleContent, err := renderTemplate("templates/rule.md", data)
	if err != nil {
		printError(fmt.Sprintf("Failed to render rule template: %v", err))
		return
	}

	if err := os.WriteFile(rulePath, []byte(ruleContent), 0644); err != nil {
		printError(fmt.Sprintf("Failed to create rule: %v", err))
		return
	}

	printSuccess(fmt.Sprintf("Created rule '%s'", slug))
	printDim(fmt.Sprintf("Location: %s", rulePath))
}

func runSpecRuleShow(cmd *cobra.Command, args []string) {
	specPath, err := checkSpecWorkspace()
	if err != nil {
		printWorkspaceError()
		return
	}

	rulesDirPath := filepath.Join(specPath, ruleDir)
	ruleFiles, err := listMarkdownFiles(rulesDirPath)
	if err != nil {
		if os.IsNotExist(err) {
			printDim("No rules directory found")
			return
		}
		printError(fmt.Sprintf("Failed to read rules directory: %v", err))
		return
	}

	if len(ruleFiles) == 0 {
		printDim("No rules found")
		printDim("Use 'nocturnal spec rule add <rule-name>' to add a rule")
		return
	}

	fmt.Println()
	fmt.Println(boldStyle.Render(fmt.Sprintf("Rules (%d)", len(ruleFiles))))
	fmt.Println()

	for i, filename := range ruleFiles {
		filePath := filepath.Join(rulesDirPath, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			printError(fmt.Sprintf("Failed to read %s: %v", filename, err))
			continue
		}

		if i > 0 {
			fmt.Println(dimStyle.Render("---"))
			fmt.Println()
		}

		fmt.Println(string(content))
	}
}

func runAgentCurrent(cmd *cobra.Command, args []string) {
	specPath, err := checkSpecWorkspace()
	if err != nil {
		printWorkspaceError()
		return
	}

	slug, proposalPath, err := getActiveProposal(specPath)
	if err != nil {
		printWarning(err.Error())
		return
	}
	if slug == "" {
		printDim("No active proposal")
		return
	}

	fmt.Println(boldStyle.Render("Active proposal:"), slug)
	printDim(fmt.Sprintf("Location: %s", proposalPath))
	fmt.Println()

	for i, doc := range proposalDocs {
		filePath := filepath.Join(proposalPath, doc.File)
		content, err := os.ReadFile(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			printError(fmt.Sprintf("Failed to read %s: %v", doc.File, err))
			continue
		}

		if i > 0 {
			fmt.Println()
			fmt.Println(dimStyle.Render("---"))
			fmt.Println()
		}

		fmt.Println(boldStyle.Render(doc.Name))
		fmt.Println()
		fmt.Print(string(content))
	}
}

func runAgentProject(cmd *cobra.Command, args []string) {
	specPath, err := checkSpecWorkspace()
	if err != nil {
		printWorkspaceError()
		return
	}

	content, err := readRulesAndProject(specPath)
	if err != nil {
		printError(err.Error())
		return
	}

	if content == "" {
		printDim("No project context found (no rules or project.md)")
		return
	}

	fmt.Print(content)
}

func runAgentSpecifications(cmd *cobra.Command, args []string) {
	specPath, err := checkSpecWorkspace()
	if err != nil {
		printWorkspaceError()
		return
	}

	content, err := readSpecifications(specPath)
	if err != nil {
		printError(err.Error())
		return
	}

	if content == "" {
		printDim("No specifications found")
		printDim("Complete a proposal with 'nocturnal spec proposal complete <slug>' to create specifications")
		return
	}

	fmt.Print(content)
}

// ValidationResult holds errors and warnings from document validation.
type ValidationResult struct {
	Document string
	Errors   []string
	Warnings []string
}

// containsText checks if content contains text (case-insensitive)
func containsText(content, text string) bool {
	return strings.Contains(strings.ToLower(content), strings.ToLower(text))
}

// validateSpecification checks for required sections and normative language.
func validateSpecification(content string) ValidationResult {
	result := ValidationResult{Document: "specification.md"}

	requiredSections := []struct {
		name     string
		required bool
		hint     string
	}{
		{"Abstract", true, "Add a 2-4 sentence summary of the specification"},
		{"Introduction", true, "Add context for why this specification exists"},
		{"Requirements", true, "List requirements using MUST/SHOULD/MAY language"},
	}

	recommendedSections := []struct {
		name string
		hint string
	}{
		{"Examples", "Provide concrete, runnable examples"},
		{"Security Considerations", "Address security implications"},
		{"Error Handling", "Define error conditions and responses"},
	}

	for _, section := range requiredSections {
		if section.required && !containsHeaderWithText(content, section.name) {
			result.Errors = append(result.Errors, fmt.Sprintf("Missing required section: %s - %s", section.name, section.hint))
		}
	}

	for _, section := range recommendedSections {
		if !containsHeaderWithText(content, section.name) {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Missing recommended section: %s - %s", section.name, section.hint))
		}
	}

	if containsHeaderWithText(content, "Requirements") {
		hasNormative := containsText(content, "MUST") || containsText(content, "SHOULD") || containsText(content, "MAY")
		if !hasNormative {
			result.Warnings = append(result.Warnings, "Requirements section should use normative language (MUST/SHOULD/MAY)")
		}
	}

	if containsText(content, "<!-- ") && containsText(content, " -->") {
		result.Warnings = append(result.Warnings, "Document contains unfilled template comments")
	}

	return result
}

// validateDesign checks for required design doc sections and metadata.
func validateDesign(content string) ValidationResult {
	result := ValidationResult{Document: "design.md"}

	requiredSections := []struct {
		name string
		hint string
	}{
		{"Context", "Establish the technical landscape and constraints"},
		{"Goals and Non-Goals", "Define goals and explicitly excluded items"},
		{"Options Considered", "Document at least 2 viable approaches"},
		{"Decision", "State the chosen approach and rationale"},
		{"Detailed Design", "Describe architecture, components, data, or API design"},
		{"Cross-Cutting Concerns", "Address security, performance, reliability, testing"},
		{"Implementation Plan", "Define phased approach and milestones"},
	}

	recommendedSections := []struct {
		name string
		hint string
	}{
		{"Open Questions", "List unresolved items with owners and blocking status"},
	}

	for _, section := range requiredSections {
		if !containsHeaderWithText(content, section.name) {
			result.Errors = append(result.Errors, fmt.Sprintf("Missing required section: %s - %s", section.name, section.hint))
		}
	}

	for _, section := range recommendedSections {
		if !containsHeaderWithText(content, section.name) {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Missing recommended section: %s - %s", section.name, section.hint))
		}
	}

	hasTitle := containsText(content, "# Design:") || containsText(content, "# design:")
	if !hasTitle {
		result.Errors = append(result.Errors, "Missing metadata: Title should be 'Design: [Feature Name]'")
	}

	hasSpecRef := containsText(content, "Specification Reference") || containsText(content, "specification reference")
	if !hasSpecRef {
		result.Warnings = append(result.Warnings, "Missing metadata: Specification Reference")
	}

	hasStatus := containsText(content, "Status:") || containsText(content, "status:")
	if !hasStatus {
		result.Warnings = append(result.Warnings, "Missing metadata: Status (Draft | Review | Approved | Superseded)")
	}

	hasOption1 := containsHeaderWithText(content, "Option 1") || containsHeaderWithText(content, "Option A")
	hasOption2 := containsHeaderWithText(content, "Option 2") || containsHeaderWithText(content, "Option B")
	if hasOption1 && !hasOption2 {
		result.Warnings = append(result.Warnings, "Only one option documented - guidelines require at least 2 alternatives or justification")
	}

	if containsText(content, "<!-- ") && containsText(content, " -->") {
		result.Warnings = append(result.Warnings, "Document contains unfilled template comments")
	}

	return result
}

// validateImplementation checks for phases and task checkboxes.
func validateImplementation(content string) ValidationResult {
	result := ValidationResult{Document: "implementation.md"}

	if !containsHeaderWithText(content, "Phase") {
		result.Errors = append(result.Errors, "Missing phases - implementation should be broken into phases")
	}

	if !containsText(content, "- [ ]") && !containsText(content, "- [x]") {
		result.Warnings = append(result.Warnings, "No task checkboxes found - consider adding actionable tasks")
	}

	if containsText(content, "<!-- ") && containsText(content, " -->") {
		result.Warnings = append(result.Warnings, "Document contains unfilled template comments")
	}

	return result
}

func runSpecProposalValidate(cmd *cobra.Command, args []string) {
	slug := args[0]
	specPath, err := checkSpecWorkspace()
	if err != nil {
		printWorkspaceError()
		return
	}

	proposalPath, err := checkProposal(specPath, slug)
	if err != nil {
		printError(err.Error())
		return
	}

	fmt.Println()
	fmt.Println(boldStyle.Render(fmt.Sprintf("Validating proposal: %s", slug)))
	fmt.Println()

	var totalErrors, totalWarnings int
	var results []ValidationResult

	documents := []struct {
		filename string
		validate func(string) ValidationResult
	}{
		{"specification.md", validateSpecification},
		{"design.md", validateDesign},
		{"implementation.md", validateImplementation},
	}

	for _, doc := range documents {
		filePath := filepath.Join(proposalPath, doc.filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				results = append(results, ValidationResult{
					Document: doc.filename,
					Errors:   []string{"File not found"},
				})
				totalErrors++
				continue
			}
			printError(fmt.Sprintf("Failed to read %s: %v", doc.filename, err))
			continue
		}

		result := doc.validate(string(content))
		results = append(results, result)
		totalErrors += len(result.Errors)
		totalWarnings += len(result.Warnings)
	}

	for _, result := range results {
		hasIssues := len(result.Errors) > 0 || len(result.Warnings) > 0

		if len(result.Errors) > 0 {
			fmt.Println(errorStyle.Render(fmt.Sprintf("✗ %s", result.Document)))
		} else if len(result.Warnings) > 0 {
			fmt.Println(warningStyle.Render(fmt.Sprintf("⚠ %s", result.Document)))
		} else {
			fmt.Println(successStyle.Render(fmt.Sprintf("✓ %s", result.Document)))
		}

		for _, err := range result.Errors {
			fmt.Println(errorStyle.Render(fmt.Sprintf("    ✗ %s", err)))
		}

		for _, warn := range result.Warnings {
			fmt.Println(warningStyle.Render(fmt.Sprintf("    ⚠ %s", warn)))
		}

		if hasIssues {
			fmt.Println()
		}
	}

	fmt.Println(dimStyle.Render("---"))
	if totalErrors == 0 && totalWarnings == 0 {
		printSuccess("All documents pass validation")
	} else {
		summary := fmt.Sprintf("Validation complete: %d error(s), %d warning(s)", totalErrors, totalWarnings)
		if totalErrors > 0 {
			printError(summary)
		} else {
			printWarning(summary)
		}
	}
}

func runSpecProposalList(cmd *cobra.Command, args []string) {
	specPath, err := checkSpecWorkspace()
	if err != nil {
		printWorkspaceError()
		return
	}

	proposalsPath := filepath.Join(specPath, proposalDir)
	entries, err := os.ReadDir(proposalsPath)
	if err != nil {
		if os.IsNotExist(err) {
			printDim("No proposals found")
			return
		}
		printError(fmt.Sprintf("Failed to read proposals directory: %v", err))
		return
	}

	activeSlug := getActiveProposalSlug(specPath)

	var proposals []string
	for _, entry := range entries {
		if entry.IsDir() {
			proposals = append(proposals, entry.Name())
		}
	}

	if len(proposals) == 0 {
		printDim("No proposals found")
		printDim("Use 'nocturnal spec proposal add <name>' to create one")
		return
	}

	fmt.Println()
	fmt.Println(boldStyle.Render(fmt.Sprintf("Proposals (%d)", len(proposals))))
	fmt.Println()

	// Header
	fmt.Printf("  %-20s %-10s %-15s %s\n",
		dimStyle.Render("NAME"),
		dimStyle.Render("STATUS"),
		dimStyle.Render("PROGRESS"),
		dimStyle.Render("DEPENDENCIES"))
	fmt.Println()

	for _, name := range proposals {
		propPath := filepath.Join(proposalsPath, name)
		total, completed := getProposalProgress(propPath)
		deps, _ := getProposalDependencies(propPath)

		// Status
		status := dimStyle.Render("inactive")
		if name == activeSlug {
			status = successStyle.Render("active")
		}

		// Progress
		var progress string
		if total > 0 {
			percentage := (completed * 100) / total
			progress = fmt.Sprintf("%d%% (%d/%d)", percentage, completed, total)
		} else {
			progress = dimStyle.Render("no tasks")
		}

		// Dependencies
		var depsStr string
		if len(deps) > 0 {
			depsStr = strings.Join(deps, ", ")
		} else {
			depsStr = dimStyle.Render("-")
		}

		// Name with indicator
		displayName := name
		if name == activeSlug {
			displayName = infoStyle.Render(name)
		}

		fmt.Printf("  %-20s %-10s %-15s %s\n", displayName, status, progress, depsStr)
	}
	fmt.Println()
}

func runSpecProposalAbandon(cmd *cobra.Command, args []string) {
	slug := args[0]
	specPath, err := checkSpecWorkspace()
	if err != nil {
		printWorkspaceError()
		return
	}

	proposalPath, err := checkProposal(specPath, slug)
	if err != nil {
		printError(err.Error())
		return
	}

	archivePath := filepath.Join(specPath, archiveDir, slug)

	// Archive all proposal documents
	if err := archiveProposalDocs(proposalPath, archivePath, proposalDocFiles); err != nil {
		printError(err.Error())
		return
	}

	// Create an abandoned marker file
	abandonedPath := filepath.Join(archivePath, ".abandoned")
	if err := os.WriteFile(abandonedPath, []byte(""), 0644); err != nil {
		printWarning(fmt.Sprintf("Failed to create abandoned marker: %v", err))
	}

	// Remove the proposal directory
	if err := os.RemoveAll(proposalPath); err != nil {
		printError(fmt.Sprintf("Failed to remove proposal workspace: %v", err))
		return
	}

	clearActiveProposalIfMatches(specPath, slug)
	printSuccess(fmt.Sprintf("Abandoned proposal '%s'", slug))
	printDim(fmt.Sprintf("Archived to %s/%s/", archiveDir, slug))
}

// buildProjectSummary creates a complete project summary including rules, specs, and active proposal
func buildProjectSummary(specPath string) string {
	var buf bytes.Buffer

	// Project rules and design
	rulesContent, err := readRulesAndProject(specPath)
	if err == nil && rulesContent != "" {
		buf.WriteString(rulesContent)
	}

	// Completed specifications
	specsContent, err := readSpecifications(specPath)
	if err == nil && specsContent != "" {
		if buf.Len() > 0 {
			buf.WriteString("\n---\n\n")
		}
		buf.WriteString(specsContent)
	}

	// Active proposal
	slug, proposalPath, err := getActiveProposal(specPath)
	if err == nil && slug != "" {
		if buf.Len() > 0 {
			buf.WriteString("\n---\n\n")
		}
		buf.WriteString(fmt.Sprintf("# Active Proposal: %s\n\n", slug))

		proposalContent, err := readProposalDocs(proposalPath)
		if err == nil && proposalContent != "" {
			buf.WriteString(proposalContent)
		}
	}

	return buf.String()
}

func runAgentSummary(cmd *cobra.Command, args []string) {
	specPath, err := checkSpecWorkspace()
	if err != nil {
		printWorkspaceError()
		return
	}

	summary := buildProjectSummary(specPath)
	if summary == "" {
		printDim("No project context found")
		printDim("Add rules, project.md, specifications, or activate a proposal")
		return
	}

	fmt.Print(summary)
}
