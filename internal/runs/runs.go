// Package runs compares two runs of the pipeline over the same samples of
// ads, as cmd/pipeline keeps them: how each stage's summary moved, and
// which headings a change mapped or unmapped.
package runs

import (
	"cmp"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// A Sample is a file of ads, as JSON Lines, named after the file.
type Sample struct {
	Name, Path string
}

// Samples returns the samples in dir: each .jsonl file, by name.
func Samples(dir string) ([]Sample, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	if err != nil {
		return nil, err
	}
	samples := make([]Sample, len(paths))
	for i, p := range paths {
		samples[i] = Sample{Name: strings.TrimSuffix(filepath.Base(p), ".jsonl"), Path: p}
	}
	return samples, nil
}

// A Change is a line of a summary as it was and as it is: Before is empty
// for a line that's new, After for one that's gone.
type Change struct {
	Before, After string
}

// figures are what a summary line's label leaves out: numbers, with their
// percentages and brackets.
var figures = regexp.MustCompile(`[\d.,%()]+`)

// labels returns each line of a summary, trimmed, with its label: what it
// says without its figures, and which time it says it, so that lines
// alike but for their figures ("the top 25 cover 31% of uses", "the top
// 50 cover…") pair in order.
func labels(summary string) (lines, keys []string) {
	times := map[string]int{}
	for line := range strings.Lines(summary) {
		l := strings.Join(strings.Fields(figures.ReplaceAllString(line, " ")), " ")
		times[l]++
		lines = append(lines, strings.TrimSpace(line))
		keys = append(keys, l+"\x00"+strconv.Itoa(times[l]))
	}
	return lines, keys
}

// Changed returns the lines of two summaries whose figures differ, paired
// by what they say: in after's order, then the lines gone from after.
func Changed(before, after string) []Change {
	oldLines, oldKeys := labels(before)
	newLines, newKeys := labels(after)
	was := map[string]string{}
	for i, k := range oldKeys {
		was[k] = oldLines[i]
	}
	var changes []Change
	for i, k := range newKeys {
		if old := was[k]; old != newLines[i] {
			changes = append(changes, Change{Before: old, After: newLines[i]})
		}
		delete(was, k)
	}
	for i, k := range oldKeys {
		if _, ok := was[k]; ok {
			changes = append(changes, Change{Before: oldLines[i]})
		}
	}
	return changes
}

// A Heading is a row of unmapped.tsv: a heading of no known kind, the ads
// using it and the characters of text under it.
type Heading struct {
	Key, Text  string
	Ads, Chars int
}

// Unmapped reads an unmapped.tsv, by key.
func Unmapped(path string) (map[string]Heading, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	hs := map[string]Heading{}
	for i, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		f := strings.Split(line, "\t")
		if i == 0 || len(f) < 4 {
			continue
		}
		ads, _ := strconv.Atoi(f[0])
		chars, _ := strconv.Atoi(f[1])
		hs[f[2]] = Heading{Key: f[2], Text: f[3], Ads: ads, Chars: chars}
	}
	return hs, nil
}

// Moved returns the headings unmapped before but not after, which a
// change mapped, and those unmapped after but not before, which it
// unmapped; each with the most text first.
func Moved(before, after map[string]Heading) (mapped, unmapped []Heading) {
	for k, h := range before {
		if _, ok := after[k]; !ok {
			mapped = append(mapped, h)
		}
	}
	for k, h := range after {
		if _, ok := before[k]; !ok {
			unmapped = append(unmapped, h)
		}
	}
	byChars := func(a, b Heading) int { return cmp.Or(cmp.Compare(b.Chars, a.Chars), cmp.Compare(a.Key, b.Key)) }
	slices.SortFunc(mapped, byChars)
	slices.SortFunc(unmapped, byChars)
	return mapped, unmapped
}
