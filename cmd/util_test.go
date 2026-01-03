package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNameToSlug(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "simple", in: "My Feature", want: "my-feature"},
		{name: "underscores", in: "API_Rate_Limiting", want: "api-rate-limiting"},
		{name: "trims_symbols", in: "Hello!!! World???", want: "hello-world"},
		{name: "collapses_hyphens", in: "a---b", want: "a-b"},
		{name: "leading_symbols", in: "---a", want: "a"},
		{name: "only_symbols", in: "---", want: ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := nameToSlug(tt.in); got != tt.want {
				t.Fatalf("nameToSlug(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseDependsOn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "markdown_field_bold",
			content: "# X\n\n**Depends on**: auth, rate-limiting\n",
			want:    []string{"auth", "rate-limiting"},
		},
		{
			name:    "plain_field",
			content: "Depends on: a, b\n",
			want:    []string{"a", "b"},
		},
		{
			name:    "none",
			content: "Depends on: none\n",
			want:    nil,
		},
		{
			name:    "empty",
			content: "Depends on:    \n",
			want:    nil,
		},
		{
			name:    "with_comment",
			content: "Depends on: auth, rate-limiting <!-- note -->\n",
			want:    []string{"auth", "rate-limiting"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := parseDependsOn(tt.content); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("parseDependsOn() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestGetMissingCompletedDependencies(t *testing.T) {
	t.Parallel()

	specPath := t.TempDir()

	proposalPath := filepath.Join(specPath, proposalDir, "feature")
	if err := os.MkdirAll(proposalPath, 0o755); err != nil {
		t.Fatalf("mkdir proposal: %v", err)
	}

	// Proposal depends on dep-a and dep-b.
	specContent := "# Feature\n\n**Depends on**: dep-a, dep-b\n"
	if err := os.WriteFile(filepath.Join(proposalPath, "specification.md"), []byte(specContent), 0o644); err != nil {
		t.Fatalf("write specification.md: %v", err)
	}

	// Only dep-a is completed.
	sectionPath := filepath.Join(specPath, sectionDir)
	if err := os.MkdirAll(sectionPath, 0o755); err != nil {
		t.Fatalf("mkdir section: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sectionPath, "dep-a.md"), []byte("# dep-a\n"), 0o644); err != nil {
		t.Fatalf("write dep-a.md: %v", err)
	}

	missing, err := getMissingCompletedDependencies(specPath, proposalPath)
	if err != nil {
		t.Fatalf("getMissingCompletedDependencies() error: %v", err)
	}

	want := []string{"dep-b"}
	if !reflect.DeepEqual(missing, want) {
		t.Fatalf("missing = %#v, want %#v", missing, want)
	}
}
