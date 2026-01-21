package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var precursorCmd = &cobra.Command{
	Use:   "precursor",
	Short: "Manage proposal precursors",
}

var precursorInitCmd = &cobra.Command{
	Use:   "init <name>",
	Short: "Initialize a new precursor bundle",
	Args:  cobra.ExactArgs(1),
	Run:   runPrecursorInit,
}

var precursorValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a precursor bundle structure",
	Run:   runPrecursorValidate,
}

var precursorPackCmd = &cobra.Command{
	Use:   "pack",
	Short: "Pack a precursor directory into a zip file",
	Run:   runPrecursorPack,
}

var precursorUnpackCmd = &cobra.Command{
	Use:   "unpack",
	Short: "Unpack a precursor zip into a directory",
	Run:   runPrecursorUnpack,
}

var (
	precursorOutPath  string
	precursorPath     string
	precursorInPath   string
	overwriteProposal bool
)

func init() {
	precursorCmd.Long = helpText("precursor")
	precursorInitCmd.Long = helpText("precursor-init")
	precursorValidateCmd.Long = helpText("precursor-validate")
	precursorPackCmd.Long = helpText("precursor-pack")
	precursorUnpackCmd.Long = helpText("precursor-unpack")

	precursorInitCmd.Flags().StringVar(&precursorOutPath, "out", "", "Output path (directory or .zip)")
	precursorInitCmd.MarkFlagRequired("out")

	precursorValidateCmd.Flags().StringVar(&precursorPath, "path", "", "Path to precursor (directory or .zip)")
	precursorValidateCmd.MarkFlagRequired("path")

	precursorPackCmd.Flags().StringVar(&precursorInPath, "in", "", "Input directory")
	precursorPackCmd.Flags().StringVar(&precursorOutPath, "out", "", "Output zip file")
	precursorPackCmd.MarkFlagRequired("in")
	precursorPackCmd.MarkFlagRequired("out")

	precursorUnpackCmd.Flags().StringVar(&precursorInPath, "in", "", "Input zip file")
	precursorUnpackCmd.Flags().StringVar(&precursorOutPath, "out", "", "Output directory")
	precursorUnpackCmd.MarkFlagRequired("in")
	precursorUnpackCmd.MarkFlagRequired("out")

	precursorCmd.AddCommand(precursorInitCmd)
	precursorCmd.AddCommand(precursorValidateCmd)
	precursorCmd.AddCommand(precursorPackCmd)
	precursorCmd.AddCommand(precursorUnpackCmd)

	rootCmd.AddCommand(precursorCmd)
}

func runPrecursorInit(cmd *cobra.Command, args []string) {
	name := args[0]
	slug := nameToSlug(name)

	outPath := precursorOutPath
	isZipOutput := strings.HasSuffix(strings.ToLower(outPath), ".zip")

	var workDir string
	var cleanup func()

	if isZipOutput {
		// Create temporary directory for scaffolding
		tmpDir, err := os.MkdirTemp("", "precursor-*")
		if err != nil {
			printError(fmt.Sprintf("Failed to create temp directory: %v", err))
			return
		}
		workDir = tmpDir
		cleanup = func() { os.RemoveAll(tmpDir) }
		defer cleanup()
	} else {
		workDir = outPath
	}

	// Create directory structure
	if err := os.MkdirAll(workDir, 0755); err != nil {
		printError(fmt.Sprintf("Failed to create directory: %v", err))
		return
	}

	templatesDir := filepath.Join(workDir, "templates")
	thirdDir := filepath.Join(workDir, "third")

	for _, dir := range []string{templatesDir, thirdDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			printError(fmt.Sprintf("Failed to create directory %s: %v", dir, err))
			return
		}
	}

	// Create precursor.yaml
	manifestContent := fmt.Sprintf(`version: 1
id: %s
description: Description of this precursor

inputs:
  - key: service_name
    prompt: "Service name?"
    required: true
  - key: dependencies
    prompt: "Comma-separated list of dependencies?"
    required: true
  - key: notes
    prompt: "Additional notes or constraints?"
    required: false
`, slug)

	if err := os.WriteFile(filepath.Join(workDir, "precursor.yaml"), []byte(manifestContent), 0644); err != nil {
		printError(fmt.Sprintf("Failed to write precursor.yaml: %v", err))
		return
	}

	// Create template files
	templates := map[string]string{
		"specification.md.tmpl": fmt.Sprintf(`# {{.Name}}

**Depends on**: <!-- comma-separated list of proposal slugs, or "none" -->
**Affected files**: <!-- comma-separated list of files/paths -->

## Abstract

This specification defines {{.Name}} for service {{.Inputs.service_name}}.

Dependencies: {{.Inputs.dependencies}}

## 1. Introduction

<!-- Context for why this specification exists -->

## 2. Requirements Notation

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in RFC 2119.

## 3. Requirements

<!-- List requirements using normative language (MUST/SHOULD/MAY) -->

## 4. Examples

<!-- Provide concrete examples -->
`),
		"design.md.tmpl": `# Design: {{.Name}}

**Status**: Draft
**Service**: {{.Inputs.service_name}}

## Context

<!-- Technical landscape and constraints -->

## Goals and Non-Goals

### Goals

<!-- What this design aims to achieve -->

### Non-Goals

<!-- What is explicitly out of scope -->

## Options Considered

### Option 1

<!-- First approach -->

### Option 2

<!-- Alternative approach -->

## Decision

<!-- Chosen approach and rationale -->

## Detailed Design

<!-- Architecture, components, APIs -->

## Cross-Cutting Concerns

### Security

<!-- Security considerations -->

### Performance

<!-- Performance implications -->

### Testing

<!-- Testing strategy -->

## Implementation Plan

See implementation.md for phased breakdown.
`,
		"implementation.md.tmpl": `# Implementation Plan: {{.Name}}

**Service**: {{.Inputs.service_name}}

## Dependencies

{{.Inputs.dependencies}}

## Phases

### Phase 1: Setup and Preparation

**Goal**: Prepare the codebase and environment for implementation.

**Tasks**:
- [ ] Task 1.1: Review and understand requirements
- [ ] Task 1.2: Set up development environment
- [ ] Task 1.3: Create feature branch

**Milestone**: Ready to begin implementation

### Phase 2: Core Implementation

**Goal**: Implement the core functionality.

**Tasks**:
- [ ] Task 2.1: Implement core logic
- [ ] Task 2.2: Add error handling
- [ ] Task 2.3: Write unit tests

**Milestone**: Core functionality complete and tested

### Phase 3: Integration and Testing

**Goal**: Integrate with existing systems and perform comprehensive testing.

**Tasks**:
- [ ] Task 3.1: Integration testing
- [ ] Task 3.2: End-to-end testing
- [ ] Task 3.3: Documentation updates

**Milestone**: Ready for deployment
`,
	}

	for filename, content := range templates {
		path := filepath.Join(templatesDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			printError(fmt.Sprintf("Failed to write %s: %v", filename, err))
			return
		}
	}

	// Create README.md
	readmeContent := fmt.Sprintf(`# %s Precursor

This precursor bundle provides a template for %s.

## Contents

- precursor.yaml: Manifest defining required inputs
- templates/: Proposal document templates
- third/: Third-party documentation (add .md files here)

## Usage

To use this precursor in a project:

`+"```bash"+`
nocturnal spec proposal add <proposal-name> --precursor-path path/to/this/precursor.zip
`+"```"+`

Fill in the generated precursor-answers.yaml, then rerun with --overwrite:

`+"```bash"+`
nocturnal spec proposal add <proposal-name> \
  --precursor-path path/to/this/precursor.zip \
  --overwrite
`+"```"+`
`, name, name)

	if err := os.WriteFile(filepath.Join(workDir, "README.md"), []byte(readmeContent), 0644); err != nil {
		printError(fmt.Sprintf("Failed to write README.md: %v", err))
		return
	}

	if isZipOutput {
		// Pack the temp directory into the output zip
		if err := packPrecursorZip(workDir, outPath); err != nil {
			printError(fmt.Sprintf("Failed to pack zip: %v", err))
			return
		}
		printSuccess(fmt.Sprintf("Created precursor zip: %s", outPath))
	} else {
		printSuccess(fmt.Sprintf("Created precursor directory: %s", outPath))
	}

	printDim(fmt.Sprintf("Edit precursor.yaml and templates/ to customize for your use case"))
}

func runPrecursorValidate(cmd *cobra.Command, args []string) {
	bundle, err := LoadPrecursorBundle(precursorPath)
	if err != nil {
		printError(fmt.Sprintf("Failed to load precursor: %v", err))
		return
	}
	defer bundle.Close()

	fmt.Println()
	fmt.Println(boldStyle.Render(fmt.Sprintf("Validating precursor: %s", precursorPath)))
	fmt.Println()

	if err := validatePrecursorStructure(bundle); err != nil {
		printError(fmt.Sprintf("Validation failed: %v", err))
		return
	}

	manifest := bundle.GetManifest()
	printSuccess("Precursor structure is valid")
	fmt.Println()
	printInfo(fmt.Sprintf("ID: %s", manifest.ID))
	printInfo(fmt.Sprintf("Version: %d", manifest.Version))
	printInfo(fmt.Sprintf("Inputs: %d", len(manifest.Inputs)))

	// List templates
	fmt.Println()
	fmt.Println(boldStyle.Render("Templates"))
	for _, tmplName := range []string{"specification.md.tmpl", "design.md.tmpl", "implementation.md.tmpl"} {
		if bundle.HasTemplate(tmplName) {
			printSuccess(fmt.Sprintf("  ✓ %s", tmplName))
		} else {
			printDim(fmt.Sprintf("  - %s (will use default)", tmplName))
		}
	}

	// List third-party docs
	docs, err := bundle.ListThirdPartyDocs()
	if err != nil {
		printWarning(fmt.Sprintf("Failed to list third-party docs: %v", err))
	} else if len(docs) > 0 {
		fmt.Println()
		fmt.Println(boldStyle.Render(fmt.Sprintf("Third-Party Docs (%d)", len(docs))))
		for _, doc := range docs {
			printInfo(fmt.Sprintf("  %s", doc))
		}
	}

	fmt.Println()
}

func runPrecursorPack(cmd *cobra.Command, args []string) {
	if !fileExists(precursorInPath) {
		printError(fmt.Sprintf("Input directory does not exist: %s", precursorInPath))
		return
	}

	info, err := os.Stat(precursorInPath)
	if err != nil || !info.IsDir() {
		printError(fmt.Sprintf("Input path must be a directory: %s", precursorInPath))
		return
	}

	// Validate before packing
	bundle, err := LoadPrecursorBundle(precursorInPath)
	if err != nil {
		printError(fmt.Sprintf("Failed to load precursor: %v", err))
		return
	}
	bundle.Close()

	if err := validatePrecursorStructure(bundle); err != nil {
		printError(fmt.Sprintf("Precursor validation failed: %v", err))
		printDim("Fix validation errors before packing")
		return
	}

	if err := packPrecursorZip(precursorInPath, precursorOutPath); err != nil {
		printError(fmt.Sprintf("Failed to pack zip: %v", err))
		return
	}

	printSuccess(fmt.Sprintf("Packed precursor to: %s", precursorOutPath))
}

func runPrecursorUnpack(cmd *cobra.Command, args []string) {
	if !fileExists(precursorInPath) {
		printError(fmt.Sprintf("Input zip does not exist: %s", precursorInPath))
		return
	}

	if fileExists(precursorOutPath) {
		printError(fmt.Sprintf("Output directory already exists: %s", precursorOutPath))
		return
	}

	zipReader, err := zip.OpenReader(precursorInPath)
	if err != nil {
		printError(fmt.Sprintf("Failed to open zip: %v", err))
		return
	}
	defer zipReader.Close()

	// Create output directory
	if err := os.MkdirAll(precursorOutPath, 0755); err != nil {
		printError(fmt.Sprintf("Failed to create output directory: %v", err))
		return
	}

	// Extract all files
	for _, file := range zipReader.File {
		if err := extractZipFile(file, precursorOutPath); err != nil {
			printError(fmt.Sprintf("Failed to extract %s: %v", file.Name, err))
			return
		}
	}

	printSuccess(fmt.Sprintf("Unpacked precursor to: %s", precursorOutPath))
}

// packPrecursorZip creates a zip file from a directory, placing all files at zip root
func packPrecursorZip(srcDir, dstZip string) error {
	// Create zip file
	zipFile, err := os.Create(dstZip)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Walk the source directory
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the source directory itself
		if path == srcDir {
			return nil
		}

		// Get relative path from source directory
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// Convert to forward slashes for zip
		zipPath := filepath.ToSlash(relPath)

		if info.IsDir() {
			// Create directory entry
			_, err := zipWriter.Create(zipPath + "/")
			return err
		}

		// Create file entry
		writer, err := zipWriter.Create(zipPath)
		if err != nil {
			return err
		}

		// Copy file content
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
}

// extractZipFile extracts a single file from a zip archive
func extractZipFile(file *zip.File, destDir string) error {
	// Construct destination path
	destPath := filepath.Join(destDir, file.Name)

	// Prevent zip slip vulnerability
	if !strings.HasPrefix(destPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
		return fmt.Errorf("illegal file path: %s", file.Name)
	}

	if file.FileInfo().IsDir() {
		return os.MkdirAll(destPath, file.Mode())
	}

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	// Open source file
	srcFile, err := file.Open()
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination file
	destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy content
	_, err = io.Copy(destFile, srcFile)
	return err
}
