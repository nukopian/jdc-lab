package headings

import (
	"testing"

	"github.com/nukopian/grammar"

	"github.com/nukopian/jdc-lab/internal/grammars"
)

// TestPaths checks that every path the package reads from a match is
// one its grammar can define: a misspelt capture reads undef, as one that
// didn't match does, so it would fail silently.
func TestPaths(t *testing.T) {
	tests := []struct {
		rule  string
		paths []string
	}{
		{"shift", []string{"count"}},
		{"span", []string{"a.h.~", "b.h.~", "unit"}},
	}
	for _, tt := range tests {
		for _, p := range tt.paths {
			if err := grammars.Schedule.CheckPath(p, grammar.Rule(tt.rule)); err != nil {
				t.Errorf("%s: %s: %v", tt.rule, p, err)
			}
		}
	}
}
