package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// write puts a sample's summaries and unmapped table into run/sample.
func write(t *testing.T, dir, run, sample, summary, unmapped string) {
	t.Helper()
	d := filepath.Join(dir, run, sample)
	if err := os.MkdirAll(filepath.Join(d, "sections"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, st := range stages {
		text := ""
		if st.dir == "sections" {
			text = summary
		}
		if err := os.WriteFile(filepath.Join(d, st.dir+".txt"), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	tsv := "ads\tchars\tkey\theading\n" + unmapped
	if err := os.WriteFile(filepath.Join(d, "sections", "unmapped.tsv"), []byte(tsv), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCompare(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "before", "a", "headings of a known kind:  70 of 100 uses (70%)\n", "39\t81961\tand\tand\n")
	write(t, dir, "after", "a", "headings of a known kind:  75 of 100 uses (75%)\n", "")
	// A sample new to after, as a new export's is, and one gone from it.
	write(t, dir, "after", "b", "", "")
	write(t, dir, "before", "c", "", "")

	var out strings.Builder
	if err := compare(&out, dir, "before", "after"); err != nil {
		t.Fatalf("compare: %v\n%s", err, out.String())
	}
	got := out.String()
	for _, want := range []string{
		"== a\n",
		"  - headings of a known kind:  70 of 100 uses (70%)\n  + headings of a known kind:  75 of 100 uses (75%)\n",
		"newly mapped: 1 headings\n",
		"== b: new in after, not in before\n",
		"== c: in before, gone from after\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("compare's output lacks %q; it was:\n%s", want, got)
		}
	}
}
