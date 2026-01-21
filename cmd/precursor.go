package cmd

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// PrecursorManifest represents the precursor.yaml file
type PrecursorManifest struct {
	Version int              `yaml:"version"`
	ID      string           `yaml:"id"`
	Desc    string           `yaml:"description"`
	Inputs  []PrecursorInput `yaml:"inputs"`
}

// PrecursorInput represents a single input parameter in the manifest
type PrecursorInput struct {
	Key      string `yaml:"key"`
	Prompt   string `yaml:"prompt"`
	Required bool   `yaml:"required"`
}

// PrecursorBundle provides access to precursor files from either a directory or zip
type PrecursorBundle struct {
	path      string
	isZip     bool
	zipReader *zip.ReadCloser
	manifest  *PrecursorManifest
}

// LoadPrecursorBundle opens a precursor from a directory or zip file
func LoadPrecursorBundle(path string) (*PrecursorBundle, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to access precursor path: %w", err)
	}

	bundle := &PrecursorBundle{path: path}

	if info.IsDir() {
		// Directory precursor
		bundle.isZip = false
		manifest, err := loadManifestFromDir(path)
		if err != nil {
			return nil, err
		}
		bundle.manifest = manifest
	} else {
		// Zip precursor
		bundle.isZip = true
		zipReader, err := zip.OpenReader(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open zip: %w", err)
		}
		bundle.zipReader = zipReader

		// Verify precursor.yaml is at zip root
		manifest, err := loadManifestFromZip(zipReader)
		if err != nil {
			zipReader.Close()
			return nil, err
		}
		bundle.manifest = manifest
	}

	return bundle, nil
}

// Close closes the precursor bundle (important for zip files)
func (b *PrecursorBundle) Close() error {
	if b.zipReader != nil {
		return b.zipReader.Close()
	}
	return nil
}

// GetManifest returns the parsed manifest
func (b *PrecursorBundle) GetManifest() *PrecursorManifest {
	return b.manifest
}

// ReadFile reads a file from the precursor bundle by relative path
func (b *PrecursorBundle) ReadFile(relPath string) ([]byte, error) {
	if b.isZip {
		return b.readFileFromZip(relPath)
	}
	return b.readFileFromDir(relPath)
}

// ListThirdPartyDocs returns paths of all files under third/
func (b *PrecursorBundle) ListThirdPartyDocs() ([]string, error) {
	if b.isZip {
		return b.listThirdPartyDocsFromZip()
	}
	return b.listThirdPartyDocsFromDir()
}

// HasTemplate checks if a template file exists in the precursor
func (b *PrecursorBundle) HasTemplate(name string) bool {
	templatePath := filepath.Join("templates", name)
	_, err := b.ReadFile(templatePath)
	return err == nil
}

// readFileFromDir reads a file from a directory precursor
func (b *PrecursorBundle) readFileFromDir(relPath string) ([]byte, error) {
	fullPath := filepath.Join(b.path, relPath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found in precursor: %s", relPath)
		}
		return nil, fmt.Errorf("failed to read file %s: %w", relPath, err)
	}
	return content, nil
}

// readFileFromZip reads a file from a zip precursor
func (b *PrecursorBundle) readFileFromZip(relPath string) ([]byte, error) {
	// Normalize path separators for zip entries (always forward slash)
	zipPath := filepath.ToSlash(relPath)

	for _, file := range b.zipReader.File {
		if file.Name == zipPath {
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open zip entry %s: %w", zipPath, err)
			}
			defer rc.Close()

			content, err := io.ReadAll(rc)
			if err != nil {
				return nil, fmt.Errorf("failed to read zip entry %s: %w", zipPath, err)
			}
			return content, nil
		}
	}

	return nil, fmt.Errorf("file not found in precursor zip: %s", relPath)
}

// listThirdPartyDocsFromDir lists all .md files under third/ in a directory precursor
func (b *PrecursorBundle) listThirdPartyDocsFromDir() ([]string, error) {
	thirdPath := filepath.Join(b.path, "third")
	if !fileExists(thirdPath) {
		return nil, nil
	}

	entries, err := os.ReadDir(thirdPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read third/ directory: %w", err)
	}

	var docs []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			docs = append(docs, filepath.Join("third", entry.Name()))
		}
	}
	return docs, nil
}

// listThirdPartyDocsFromZip lists all .md files under third/ in a zip precursor
func (b *PrecursorBundle) listThirdPartyDocsFromZip() ([]string, error) {
	var docs []string
	for _, file := range b.zipReader.File {
		if strings.HasPrefix(file.Name, "third/") && strings.HasSuffix(file.Name, ".md") {
			docs = append(docs, file.Name)
		}
	}
	return docs, nil
}

// loadManifestFromDir loads precursor.yaml from a directory
func loadManifestFromDir(dirPath string) (*PrecursorManifest, error) {
	manifestPath := filepath.Join(dirPath, "precursor.yaml")
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("precursor.yaml not found at directory root: %s", dirPath)
		}
		return nil, fmt.Errorf("failed to read precursor.yaml: %w", err)
	}

	var manifest PrecursorManifest
	if err := yaml.Unmarshal(content, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse precursor.yaml: %w", err)
	}

	return &manifest, nil
}

// loadManifestFromZip loads precursor.yaml from a zip file
func loadManifestFromZip(zipReader *zip.ReadCloser) (*PrecursorManifest, error) {
	// Find precursor.yaml at zip root (exact match, no directory prefix)
	var manifestFile *zip.File
	var nestedPath string

	for _, file := range zipReader.File {
		if file.Name == "precursor.yaml" {
			manifestFile = file
			break
		}
		// Check if precursor.yaml exists but is nested
		if strings.HasSuffix(file.Name, "/precursor.yaml") || strings.Contains(file.Name, "/precursor.yaml") {
			nestedPath = file.Name
		}
	}

	if manifestFile == nil {
		if nestedPath != "" {
			return nil, fmt.Errorf("precursor.yaml must be at zip root (no wrapping folder). Found: %s", nestedPath)
		}
		return nil, fmt.Errorf("precursor.yaml not found at zip root")
	}

	rc, err := manifestFile.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open precursor.yaml from zip: %w", err)
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("failed to read precursor.yaml from zip: %w", err)
	}

	var manifest PrecursorManifest
	if err := yaml.Unmarshal(content, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse precursor.yaml: %w", err)
	}

	return &manifest, nil
}

// renderTemplateFromString renders a Go template from a string with the given data
func renderTemplateFromString(name, templateContent string, data any) (string, error) {
	// Create template with helper functions
	tmpl, err := template.New(name).Funcs(template.FuncMap{
		"contains": func(slice []any, item any) bool {
			for _, v := range slice {
				if v == item {
					return true
				}
			}
			return false
		},
		"get": func(m map[string]any, key string) any {
			return m[key]
		},
	}).Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", name, err)
	}

	return buf.String(), nil
}

// validatePrecursorStructure validates that a precursor bundle has the required structure
func validatePrecursorStructure(bundle *PrecursorBundle) error {
	// Check manifest
	manifest := bundle.GetManifest()
	if manifest.Version != 1 {
		return fmt.Errorf("unsupported precursor version: %d (expected 1)", manifest.Version)
	}

	// Check templates exist and parse
	requiredTemplates := []string{"specification.md.tmpl", "design.md.tmpl", "implementation.md.tmpl"}
	for _, tmplName := range requiredTemplates {
		content, err := bundle.ReadFile(filepath.Join("templates", tmplName))
		if err != nil {
			// Templates are optional if we fall back to embedded defaults
			continue
		}

		// Try to parse the template with minimal data
		testData := struct {
			Name   string
			Slug   string
			Inputs map[string]any
		}{
			Name:   "test",
			Slug:   "test",
			Inputs: make(map[string]any),
		}

		if _, err := renderTemplateFromString(tmplName, string(content), testData); err != nil {
			return fmt.Errorf("template %s failed to parse: %w", tmplName, err)
		}
	}

	return nil
}
