package sections

import "testing"

// TestPaths checks that every path Explain reads from a match is one its
// grammar can define: a misspelt capture reads undef, as one that didn't
// match does, so it would fail silently.
func TestPaths(t *testing.T) {
	for _, p := range []string{"~"} {
		if err := finder.CheckPath(p); err != nil {
			t.Errorf("%s: %v", p, err)
		}
	}
}
