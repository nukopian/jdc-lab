package facts

import (
	"testing"

	"github.com/nukopian/grammar"

	"github.com/nukopian/jdc-lab/internal/grammars"
)

// TestPaths checks that every path Find reads from a match is one its
// grammar can define: a misspelt capture reads undef, as one that didn't
// match does, so it would fail silently.
func TestPaths(t *testing.T) {
	// fact reads ~ and from from every match.
	tests := []struct {
		name  string
		g     *grammar.Grammar
		paths []string
	}{
		{"pay", grammars.Pay, []string{"~", "from", "cur", "cur[0].~", "min", "max", "min.k", "max.k", "upto",
			"period", "period.~", "min.num.~", "max.num.~"}},
		{"hours", grammars.Hours, []string{"~", "from", "num.~", "per.week", "per.day"}},
		{"pattern", grammars.Pattern, []string{"~", "from", "made"}},
		{"schedule", grammars.Schedule, []string{"~", "from", "days", "days.mf", "days.first.~", "days.last.~",
			"times", "times.a.~", "times.b.~", "rota", "rota.~"}},
		{"postcode", grammars.Postcode, []string{"~", "from"}},
	}
	for _, tt := range tests {
		for _, p := range tt.paths {
			if err := tt.g.CheckPath(p); err != nil {
				t.Errorf("%s: %s: %v", tt.name, p, err)
			}
		}
	}
}
