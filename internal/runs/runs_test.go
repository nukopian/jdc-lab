package runs

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestChanged(t *testing.T) {
	before := `1890 ads (110 duplicates left out)

headings of a known kind:        5230 of 6913 uses (76%)
  the top 25  cover      31% of uses
  the top 50  cover      39% of uses
summary              90%      28%
company              17%       5%
`
	after := `1890 ads (110 duplicates left out)

headings of a known kind:        5301 of 6913 uses (77%)
  the top 25  cover      31% of uses
  the top 50  cover      40% of uses
summary              90%      28%
footers               3%       1%
`
	want := []Change{
		{"headings of a known kind:        5230 of 6913 uses (76%)", "headings of a known kind:        5301 of 6913 uses (77%)"},
		// Alike but for their figures, the top 50 pairs with the top 50.
		{"the top 50  cover      39% of uses", "the top 50  cover      40% of uses"},
		{"", "footers               3%       1%"},
		{"company              17%       5%", ""},
	}
	if got := Changed(before, after); !slices.Equal(got, want) {
		t.Errorf("Changed:\n got %q\nwant %q", got, want)
	}
	if got := Changed(before, before); len(got) != 0 {
		t.Errorf("Changed of a summary and itself = %q, want none", got)
	}
}

func TestUnmappedAndMoved(t *testing.T) {
	dir := t.TempDir()
	write := func(name, tsv string) map[string]Heading {
		t.Helper()
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(tsv), 0o644); err != nil {
			t.Fatal(err)
		}
		hs, err := Unmapped(p)
		if err != nil {
			t.Fatal(err)
		}
		return hs
	}
	before := write("before.tsv", "ads\tchars\tkey\theading\n"+
		"39\t81961\tand\tand\n"+
		"34\t18972\twho thrives here\tWho thrives here\n"+
		"24\t6872\tplease note\tPlease note\n")
	after := write("after.tsv", "ads\tchars\tkey\theading\n"+
		"24\t6872\tplease note\tPlease note\n"+
		"2\t500\tor\tOR\n")

	if got := before["and"]; got != (Heading{Key: "and", Text: "and", Ads: 39, Chars: 81961}) {
		t.Errorf(`Unmapped: "and" = %+v`, got)
	}
	mapped, unmapped := Moved(before, after)
	keys := func(hs []Heading) []string {
		var ks []string
		for _, h := range hs {
			ks = append(ks, h.Key)
		}
		return ks
	}
	if got, want := keys(mapped), []string{"and", "who thrives here"}; !slices.Equal(got, want) {
		t.Errorf("Moved: mapped = %q, want %q", got, want)
	}
	if got, want := keys(unmapped), []string{"or"}; !slices.Equal(got, want) {
		t.Errorf("Moved: unmapped = %q, want %q", got, want)
	}
}

func TestSamples(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"ads-20261025.jsonl", "ads-20261024.jsonl", "extract.xml"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := Samples(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []Sample{
		{"ads-20261024", filepath.Join(dir, "ads-20261024.jsonl")},
		{"ads-20261025", filepath.Join(dir, "ads-20261025.jsonl")},
	}
	if !slices.Equal(got, want) {
		t.Errorf("Samples = %q, want %q", got, want)
	}
}

func TestSampleAds(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ads.jsonl")
	if err := os.WriteFile(path, []byte("{}\n{}\n{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if n, err := (Sample{Name: "ads", Path: path}).Ads(); err != nil || n != 3 {
		t.Errorf("got %d, %v; want 3 ads", n, err)
	}
}

func TestTimings(t *testing.T) {
	ts := []Timing{
		{Sample: "ads-20261024", Stage: "blocks", Ads: 2000, Wall: 500 * time.Millisecond, CPU: 750 * time.Millisecond, PeakRSS: 40 << 20},
		// A stage too quick to time gives no rate, not a division by zero.
		{Sample: "ads-20261024", Stage: "sections", Ads: 2000},
	}
	var tsv, table strings.Builder
	if err := WriteTimingsTSV(&tsv, ts); err != nil {
		t.Fatal(err)
	}
	want := "sample\tstage\tads\twall_s\tcpu_s\tads_per_s\tpeak_mb\n" +
		"ads-20261024\tblocks\t2000\t0.500\t0.750\t4000\t40\n" +
		"ads-20261024\tsections\t2000\t0.000\t0.000\t0\t0\n"
	if tsv.String() != want {
		t.Errorf("TSV:\ngot  %q\nwant %q", tsv.String(), want)
	}
	if err := WriteTimings(&table, ts); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"ads/s", "ads-20261024  blocks", "0.50s", "4000", "40 MB"} {
		if !strings.Contains(table.String(), want) {
			t.Errorf("table lacks %q; it was:\n%s", want, table.String())
		}
	}
}
