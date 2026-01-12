package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoadSave(t *testing.T) {
	tmpDir := t.TempDir()

	// Test loading default when no file exists
	config, err := loadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loadConfig failed: %v", err)
	}
	if config.Validation.Strict != false {
		t.Error("expected Strict to be false by default")
	}
	if config.Context.MaxFileLines != 50 {
		t.Errorf("expected MaxFileLines to be 50, got %d", config.Context.MaxFileLines)
	}

	// Modify and save
	config.Validation.Strict = true
	config.Context.IncludeAffectedFiles = true
	config.Context.MaxFileLines = 100

	if err := saveConfig(tmpDir, config); err != nil {
		t.Fatalf("saveConfig failed: %v", err)
	}

	// Verify file was created
	configPath := filepath.Join(tmpDir, configFileName)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	// Load and verify
	loaded, err := loadConfig(tmpDir)
	if err != nil {
		t.Fatalf("loadConfig after save failed: %v", err)
	}
	if loaded.Validation.Strict != true {
		t.Error("expected Strict to be true after load")
	}
	if loaded.Context.IncludeAffectedFiles != true {
		t.Error("expected IncludeAffectedFiles to be true after load")
	}
	if loaded.Context.MaxFileLines != 100 {
		t.Errorf("expected MaxFileLines to be 100, got %d", loaded.Context.MaxFileLines)
	}
}

func TestParseAffectedFiles(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "markdown field bold",
			content: "**Affected files**: cmd/spec.go, cmd/util.go\n\n## Abstract",
			want:    []string{"cmd/spec.go", "cmd/util.go"},
		},
		{
			name:    "plain field",
			content: "Affected files: internal/foo.go\n\n## Abstract",
			want:    []string{"internal/foo.go"},
		},
		{
			name:    "with comment",
			content: "**Affected files**: cmd/mcp.go <!-- more files -->\n\n## Abstract",
			want:    []string{"cmd/mcp.go"},
		},
		{
			name:    "empty",
			content: "**Affected files**: \n\n## Abstract",
			want:    nil,
		},
		{
			name:    "template placeholder",
			content: "**Affected files**: <!-- list files here -->\n\n## Abstract",
			want:    nil,
		},
		{
			name:    "no field",
			content: "# My Proposal\n\n## Abstract",
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAffectedFiles(tt.content)
			if len(got) != len(tt.want) {
				t.Errorf("parseAffectedFiles() = %v, want %v", got, tt.want)
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("parseAffectedFiles()[%d] = %v, want %v", i, v, tt.want[i])
				}
			}
		})
	}
}

func TestReadAffectedFileContent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file with 10 lines
	testFile := filepath.Join(tmpDir, "test.go")
	content := "line 1\nline 2\nline 3\nline 4\nline 5\nline 6\nline 7\nline 8\nline 9\nline 10\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Read with limit larger than file
	got, truncated, err := readAffectedFileContent(testFile, 20)
	if err != nil {
		t.Fatalf("readAffectedFileContent failed: %v", err)
	}
	if truncated {
		t.Error("expected not truncated")
	}
	if got != content {
		t.Errorf("content mismatch: got %q", got)
	}

	// Read with limit smaller than file
	got, truncated, err = readAffectedFileContent(testFile, 5)
	if err != nil {
		t.Fatalf("readAffectedFileContent failed: %v", err)
	}
	if !truncated {
		t.Error("expected truncated")
	}
	expected := "line 1\nline 2\nline 3\nline 4\nline 5"
	if got != expected {
		t.Errorf("content mismatch: got %q, want %q", got, expected)
	}

	// Read non-existent file
	_, _, err = readAffectedFileContent(filepath.Join(tmpDir, "nonexistent.go"), 10)
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}
