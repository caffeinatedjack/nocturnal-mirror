package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// State represents the nocturnal state file.
type State struct {
	Version      int                                    `json:"version"`
	Active       []string                               `json:"active"`
	Primary      string                                 `json:"primary"`
	Hashes       map[string]map[string]string           `json:"hashes,omitempty"`
	Maintenance  map[string]map[string]MaintenanceState `json:"maintenance,omitempty"`
	GitSnapshots map[string]GitSnapshotState            `json:"git_snapshots,omitempty"`
}

// GitSnapshotState tracks git snapshots for task execution
type GitSnapshotState struct {
	SnapshotRef string `json:"snapshot_ref,omitempty"` // Git ref at snapshot time
	TaskID      string `json:"task_id"`
	Timestamp   string `json:"timestamp"` // RFC3339 timestamp
}

// MaintenanceState tracks when a maintenance requirement was last actioned.
type MaintenanceState struct {
	LastActioned string `json:"last_actioned"` // RFC3339 timestamp
}

// loadState reads the state file and returns active proposals.
func loadState(specPath string) (*State, error) {
	statePath := filepath.Join(specPath, ".nocturnal.json")
	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{
				Version:      1,
				Active:       []string{},
				Hashes:       make(map[string]map[string]string),
				Maintenance:  make(map[string]map[string]MaintenanceState),
				GitSnapshots: make(map[string]GitSnapshotState),
			}, nil
		}
		return nil, err
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	if state.Hashes == nil {
		state.Hashes = make(map[string]map[string]string)
	}
	if state.Maintenance == nil {
		state.Maintenance = make(map[string]map[string]MaintenanceState)
	}
	if state.GitSnapshots == nil {
		state.GitSnapshots = make(map[string]GitSnapshotState)
	}

	return &state, nil
}

// getActiveProposal returns the primary active proposal slug.
func getActiveProposal(specPath string) string {
	state, err := loadState(specPath)
	if err != nil {
		return ""
	}

	if state.Primary != "" {
		return state.Primary
	}

	if len(state.Active) > 0 {
		return state.Active[0]
	}

	return ""
}
