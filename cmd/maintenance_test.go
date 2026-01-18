package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseMaintenanceFile(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		wantErr       bool
		wantErrMsg    string
		wantReqCount  int
		wantFirstID   string
		wantFirstFreq string
	}{
		{
			name: "valid requirements with frequencies",
			content: `# Maintenance: Test

## Requirements
- Run tests [id=test] [freq=weekly]
- Update deps [id=deps] [freq=monthly]
`,
			wantReqCount:  2,
			wantFirstID:   "test",
			wantFirstFreq: "weekly",
		},
		{
			name: "requirement without frequency (always due)",
			content: `# Maintenance: Test

## Requirements
- Do something [id=always]
`,
			wantReqCount:  1,
			wantFirstID:   "always",
			wantFirstFreq: "",
		},
		{
			name: "missing id",
			content: `# Maintenance: Test

## Requirements
- Run tests [freq=weekly]
`,
			wantErr:    true,
			wantErrMsg: "missing [id=...]",
		},
		{
			name: "duplicate id",
			content: `# Maintenance: Test

## Requirements
- First [id=dup]
- Second [id=dup]
`,
			wantErr:    true,
			wantErrMsg: "duplicate id",
		},
		{
			name: "unknown frequency",
			content: `# Maintenance: Test

## Requirements
- Test [id=test] [freq=hourly]
`,
			wantErr:    true,
			wantErrMsg: "unknown frequency",
		},
		{
			name: "tokens in any order",
			content: `# Maintenance: Test

## Requirements
- First [freq=daily] [id=first]
- Second [id=second] [freq=weekly]
`,
			wantReqCount: 2,
			wantFirstID:  "first",
		},
		{
			name: "stops at next section",
			content: `# Maintenance: Test

## Requirements
- First [id=first]

## Notes
- Not a requirement [id=ignored]
`,
			wantReqCount: 1,
			wantFirstID:  "first",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "test.md")
			if err := os.WriteFile(filePath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			// Parse with empty state
			state := &State{Maintenance: make(map[string]map[string]MaintenanceState)}
			reqs, err := parseMaintenanceFile(filePath, state, "test")

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErrMsg)
				}
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErrMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(reqs) != tt.wantReqCount {
				t.Fatalf("expected %d requirements, got %d", tt.wantReqCount, len(reqs))
			}

			if tt.wantReqCount > 0 {
				if reqs[0].ID != tt.wantFirstID {
					t.Errorf("expected first ID %q, got %q", tt.wantFirstID, reqs[0].ID)
				}
				if tt.wantFirstFreq != "" && reqs[0].Freq != tt.wantFirstFreq {
					t.Errorf("expected first freq %q, got %q", tt.wantFirstFreq, reqs[0].Freq)
				}
			}
		})
	}
}

func TestComputeDue(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		freq         string
		lastActioned string
		wantDue      bool
	}{
		{
			name:    "no freq is always due",
			freq:    "",
			wantDue: true,
		},
		{
			name:         "never actioned is due",
			freq:         "weekly",
			lastActioned: "",
			wantDue:      true,
		},
		{
			name:         "weekly - actioned yesterday is not due",
			freq:         "weekly",
			lastActioned: now.AddDate(0, 0, -1).Format(time.RFC3339),
			wantDue:      false,
		},
		{
			name:         "weekly - actioned 8 days ago is due",
			freq:         "weekly",
			lastActioned: now.AddDate(0, 0, -8).Format(time.RFC3339),
			wantDue:      true,
		},
		{
			name:         "daily - actioned yesterday is due",
			freq:         "daily",
			lastActioned: now.AddDate(0, 0, -1).Format(time.RFC3339),
			wantDue:      true,
		},
		{
			name:         "monthly - actioned 20 days ago is not due",
			freq:         "monthly",
			lastActioned: now.AddDate(0, 0, -20).Format(time.RFC3339),
			wantDue:      false,
		},
		{
			name:         "monthly - actioned 35 days ago is due",
			freq:         "monthly",
			lastActioned: now.AddDate(0, 0, -35).Format(time.RFC3339),
			wantDue:      true,
		},
		{
			name:         "invalid timestamp is due",
			freq:         "weekly",
			lastActioned: "invalid",
			wantDue:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeDue(tt.freq, tt.lastActioned)
			if got != tt.wantDue {
				t.Errorf("computeDue() = %v, want %v", got, tt.wantDue)
			}
		})
	}
}

func TestMaintenanceStateLoad(t *testing.T) {
	// Test that state loads correctly with and without maintenance field
	tmpDir := t.TempDir()

	t.Run("load empty state", func(t *testing.T) {
		state, err := loadState(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state.Maintenance == nil {
			t.Error("expected Maintenance to be initialized")
		}
	})

	t.Run("load state without maintenance field", func(t *testing.T) {
		// Write old-style state
		stateContent := `{"version":1,"active":[],"primary":""}`
		statePath := filepath.Join(tmpDir, stateFile)
		if err := os.WriteFile(statePath, []byte(stateContent), 0644); err != nil {
			t.Fatalf("failed to write state: %v", err)
		}

		state, err := loadState(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state.Maintenance == nil {
			t.Error("expected Maintenance to be initialized for old state")
		}
	})

	t.Run("save and load state with maintenance", func(t *testing.T) {
		state := &State{
			Version:     1,
			Active:      []string{},
			Primary:     "",
			Hashes:      make(map[string]map[string]string),
			Maintenance: make(map[string]map[string]MaintenanceState),
		}

		state.Maintenance["test"] = map[string]MaintenanceState{
			"req1": {LastActioned: "2026-01-18T10:00:00Z"},
		}

		if err := saveState(tmpDir, state); err != nil {
			t.Fatalf("failed to save state: %v", err)
		}

		loaded, err := loadState(tmpDir)
		if err != nil {
			t.Fatalf("failed to load state: %v", err)
		}

		if loaded.Maintenance["test"]["req1"].LastActioned != "2026-01-18T10:00:00Z" {
			t.Error("maintenance state not preserved")
		}
	})
}
