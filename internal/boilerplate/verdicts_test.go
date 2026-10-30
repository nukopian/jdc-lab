package boilerplate

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestVerdicts(t *testing.T) {
	path := filepath.Join(t.TempDir(), "repeated.tsv")
	tsv := "groups\tads\tchars\tkinds\tverdict\ttext\n" +
		"9\t40\t9000\tlegal\tboilerplate\tWe are an Equal Opportunities employer.\n" +
		"2\t3\t300\trequirements\ttemplate\tFull UK driving licence.\n" +
		"6\t12\t800\tcompany\tboilerplate\tAbout Us: We Place Drivers\n"
	if err := os.WriteFile(path, []byte(tsv), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Verdicts(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"about us we place drivers", "we are an equal opportunities employer"}
	if keys := slices.Sorted(maps.Keys(got)); !slices.Equal(keys, want) {
		t.Errorf("Verdicts keys = %q, want %q", keys, want)
	}
}
