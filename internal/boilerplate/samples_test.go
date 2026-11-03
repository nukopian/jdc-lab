package boilerplate

import (
	"bufio"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/nukopian/jdc-lab/internal/headings"
)

// samplePath writes the first n ads of the latest run to a file of their
// own, and returns its path, skipping where there's no run.
func samplePath(b testing.TB, n int) string {
	in, err := os.Open("../../data/runs/latest/ads.jsonl")
	if err != nil {
		b.Skip("no run to read")
	}
	defer in.Close()
	path := filepath.Join(b.TempDir(), "ads.jsonl")
	out, err := os.Create(path)
	if err != nil {
		b.Fatal(err)
	}
	r := bufio.NewReaderSize(in, 1<<20)
	for range n {
		line, err := r.ReadBytes('\n')
		if len(line) == 0 && err != nil {
			break
		}
		if _, err := out.Write(line); err != nil {
			b.Fatal(err)
		}
	}
	if err := out.Close(); err != nil {
		b.Fatal(err)
	}
	return path
}

// BenchmarkSamples reads 5,000 ads of the latest run, and asks of each of
// their texts what cmd/boilerplate asks: whether it's boilerplate,
// whether it's repeated, and its groups of titles. It's skipped where
// there's no run.
func BenchmarkSamples(b *testing.B) {
	path := samplePath(b, 5000)
	b.ReportAllocs()
	for b.Loop() {
		all, ix, _, err := Read(path)
		if err != nil {
			b.Fatal(err)
		}
		for _, a := range all {
			src := SourceOf(a.Ad)
			for _, s := range a.Sections {
				for _, p := range s.Parts {
					if p.Role != headings.Text {
						continue
					}
					if ix.Boilerplate(src, p.Text) || ix.Repeated(src, p.Text) {
						ix.Groups(src, p.Text)
					}
				}
			}
		}
	}
}

// answers returns what cmd/boilerplate asks of each text of the ads in
// the file at path, one line a text, and the long texts alike, each
// key's group named by the first of its keys in sorting order, as the
// names Index gives them depend on the order texts are joined.
func answers(t *testing.T, path string) ([]string, map[string]string) {
	all, ix, _, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	for _, a := range all {
		src := SourceOf(a.Ad)
		for _, s := range a.Sections {
			for _, p := range s.Parts {
				if p.Role == headings.Text {
					lines = append(lines, fmt.Sprintf("%t %t %t %d %s", ix.Boilerplate(src, p.Text),
						ix.Repeated(src, p.Text), ix.Template(src, p.Text), ix.Groups(src, p.Text), p.Text))
				}
			}
		}
	}
	first := map[string]string{} // a group's name → its first key
	for k, g := range ix.alike {
		if f, ok := first[g]; !ok || k < f {
			first[g] = k
		}
	}
	groups := map[string]string{}
	for k, g := range ix.alike {
		groups[k] = first[g]
	}
	return lines, groups
}

// Finding texts alike through sorted slices changes no run of words, no
// group of texts alike, and no answer, over 3,000 ads of the latest run.
func TestFastAlikeChangesNothing(t *testing.T) {
	if testing.Short() {
		t.Skip("short")
	}
	path := samplePath(t, 3000)
	for _, k := range []string{"", "one two three four", "a b c d e", "a b c d e a b c d e f", agency + " " + duties + " " + location} {
		want := slices.Sorted(maps.Keys(shingles(k)))
		if got := shingleList(k); !slices.Equal(got, want) && len(got)+len(want) > 0 {
			t.Errorf("shingleList(%q) = %v, want %v", k, got, want)
		}
	}
	with, withGroups := answers(t, path)
	defer func() { fastAlike = true }()
	fastAlike = false
	without, withoutGroups := answers(t, path)
	if len(with) != len(without) {
		t.Fatalf("%d answers, %d without", len(with), len(without))
	}
	for i := range with {
		if with[i] != without[i] {
			t.Errorf("with: %s\nwithout: %s", with[i], without[i])
		}
	}
	if !maps.Equal(withGroups, withoutGroups) {
		t.Errorf("groups of texts alike differ: %d keys with, %d without", len(withGroups), len(withoutGroups))
	}
	joined := 0
	for k, g := range withGroups {
		if k != g {
			joined++
		}
	}
	t.Logf("%d answers, %d long texts, %d of them joined to another", len(with), len(withGroups), joined)
}
