package cmd

import (
	"testing"
)

func TestDetectCycles(t *testing.T) {
	tests := []struct {
		name      string
		nodes     map[string]*ProposalNode
		wantCycle bool
	}{
		{
			name: "no cycles",
			nodes: map[string]*ProposalNode{
				"a": {Slug: "a", Dependencies: []string{"b"}},
				"b": {Slug: "b", Dependencies: []string{"c"}},
				"c": {Slug: "c", Dependencies: []string{}},
			},
			wantCycle: false,
		},
		{
			name: "simple cycle",
			nodes: map[string]*ProposalNode{
				"a": {Slug: "a", Dependencies: []string{"b"}},
				"b": {Slug: "b", Dependencies: []string{"a"}},
			},
			wantCycle: true,
		},
		{
			name: "three node cycle",
			nodes: map[string]*ProposalNode{
				"a": {Slug: "a", Dependencies: []string{"b"}},
				"b": {Slug: "b", Dependencies: []string{"c"}},
				"c": {Slug: "c", Dependencies: []string{"a"}},
			},
			wantCycle: true,
		},
		{
			name: "diamond - no cycle",
			nodes: map[string]*ProposalNode{
				"a": {Slug: "a", Dependencies: []string{"b", "c"}},
				"b": {Slug: "b", Dependencies: []string{"d"}},
				"c": {Slug: "c", Dependencies: []string{"d"}},
				"d": {Slug: "d", Dependencies: []string{}},
			},
			wantCycle: false,
		},
		{
			name:      "empty",
			nodes:     map[string]*ProposalNode{},
			wantCycle: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cycles := detectCycles(tt.nodes)
			gotCycle := len(cycles) > 0
			if gotCycle != tt.wantCycle {
				t.Errorf("detectCycles() found cycle = %v, want %v", gotCycle, tt.wantCycle)
			}
		})
	}
}

func TestGetRelevantNodes(t *testing.T) {
	nodes := map[string]*ProposalNode{
		"a": {Slug: "a", Dependencies: []string{"b"}},
		"b": {Slug: "b", Dependencies: []string{"c"}},
		"c": {Slug: "c", Dependencies: []string{}},
		"d": {Slug: "d", Dependencies: []string{"b"}},
		"e": {Slug: "e", Dependencies: []string{}},
	}

	// Get nodes relevant to "b" - should include a, b, c, d (not e)
	relevant := getRelevantNodes(nodes, "b")

	if _, ok := relevant["b"]; !ok {
		t.Error("expected 'b' to be in relevant nodes")
	}
	if _, ok := relevant["c"]; !ok {
		t.Error("expected 'c' to be in relevant nodes (dependency of b)")
	}
	if _, ok := relevant["a"]; !ok {
		t.Error("expected 'a' to be in relevant nodes (depends on b)")
	}
	if _, ok := relevant["d"]; !ok {
		t.Error("expected 'd' to be in relevant nodes (depends on b)")
	}
	if _, ok := relevant["e"]; ok {
		t.Error("expected 'e' NOT to be in relevant nodes (unrelated)")
	}
}
