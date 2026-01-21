package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const precursorAnswersFile = "precursor-answers.yaml"

// PrecursorAnswers represents the answers file stored in a proposal directory
type PrecursorAnswers struct {
	Version       int                             `yaml:"version"`
	PrecursorPath string                          `yaml:"precursor_path"`
	Inputs        map[string]PrecursorAnswerInput `yaml:"inputs"`
}

// PrecursorAnswerInput represents a single input answer
type PrecursorAnswerInput struct {
	Required bool   `yaml:"required"`
	Prompt   string `yaml:"prompt"`
	Value    string `yaml:"value"`
}

// loadPrecursorAnswers loads the answers file from a proposal directory
func loadPrecursorAnswers(proposalPath string) (*PrecursorAnswers, error) {
	answersPath := filepath.Join(proposalPath, precursorAnswersFile)

	// If file doesn't exist, return empty answers
	if !fileExists(answersPath) {
		return &PrecursorAnswers{
			Version: 1,
			Inputs:  make(map[string]PrecursorAnswerInput),
		}, nil
	}

	content, err := os.ReadFile(answersPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read answers file: %w", err)
	}

	var answers PrecursorAnswers
	if err := yaml.Unmarshal(content, &answers); err != nil {
		return nil, fmt.Errorf("failed to parse answers file: %w", err)
	}

	if answers.Inputs == nil {
		answers.Inputs = make(map[string]PrecursorAnswerInput)
	}

	return &answers, nil
}

// savePrecursorAnswers writes the answers file to a proposal directory
func savePrecursorAnswers(proposalPath string, answers *PrecursorAnswers) error {
	answersPath := filepath.Join(proposalPath, precursorAnswersFile)

	content, err := yaml.Marshal(answers)
	if err != nil {
		return fmt.Errorf("failed to serialize answers: %w", err)
	}

	if err := os.WriteFile(answersPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write answers file: %w", err)
	}

	return nil
}

// mergePrecursorAnswers merges manifest inputs with existing answers
// Preserves existing values, adds new inputs, updates prompts/required flags
func mergePrecursorAnswers(manifest *PrecursorManifest, existing *PrecursorAnswers, precursorPath string) *PrecursorAnswers {
	merged := &PrecursorAnswers{
		Version:       1,
		PrecursorPath: precursorPath,
		Inputs:        make(map[string]PrecursorAnswerInput),
	}

	// Process each input from the manifest
	for _, input := range manifest.Inputs {
		existingAnswer := existing.Inputs[input.Key]

		merged.Inputs[input.Key] = PrecursorAnswerInput{
			Required: input.Required,
			Prompt:   input.Prompt,
			Value:    existingAnswer.Value, // Preserve existing value
		}
	}

	return merged
}

// getMissingRequiredInputs returns a list of input keys that are required but have empty values
func getMissingRequiredInputs(answers *PrecursorAnswers) []string {
	var missing []string

	for key, input := range answers.Inputs {
		if input.Required && strings.TrimSpace(input.Value) == "" {
			missing = append(missing, key)
		}
	}

	return missing
}

// answersToTemplateData converts answers to a map suitable for template rendering
func answersToTemplateData(answers *PrecursorAnswers) map[string]any {
	data := make(map[string]any)

	for key, input := range answers.Inputs {
		value := input.Value

		// Special handling for comma-separated lists (common pattern)
		if strings.Contains(value, ",") {
			// Convert to slice for template iteration
			parts := strings.Split(value, ",")
			var trimmed []any
			for _, part := range parts {
				if t := strings.TrimSpace(part); t != "" {
					trimmed = append(trimmed, t)
				}
			}
			data[key] = trimmed
		} else {
			data[key] = value
		}
	}

	return data
}
