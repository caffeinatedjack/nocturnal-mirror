package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDocFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "lib.md")

	content := "---\n# Component One\nLine 1\nLine 2\n---\n# Component Two\nSecond\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	components, err := parseDocFile(path)
	if err != nil {
		t.Fatalf("parseDocFile error: %v", err)
	}

	if len(components) != 2 {
		t.Fatalf("expected 2 components, got %d", len(components))
	}

	if components[0].Name != "Component One" {
		t.Fatalf("component[0].Name = %q", components[0].Name)
	}
	if components[0].Source != "lib.md" {
		t.Fatalf("component[0].Source = %q", components[0].Source)
	}
	if components[0].Content != "Line 1\nLine 2" {
		t.Fatalf("component[0].Content = %q", components[0].Content)
	}

	if components[1].Name != "Component Two" {
		t.Fatalf("component[1].Name = %q", components[1].Name)
	}
	if components[1].Content != "Second" {
		t.Fatalf("component[1].Content = %q", components[1].Content)
	}
}
