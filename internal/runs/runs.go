// Package runs compares two runs of the pipeline over the same samples of
// ads, as cmd/pipeline keeps them: how each stage's summary moved, and
// which headings a change mapped or unmapped. It also writes how long each
// stage took.
package runs

import (
	"bytes"
	"cmp"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
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

// Ads returns how many ads the sample has: its lines, the last counted
// whether it ends with a newline or not. It reads the file a piece at a
// time, as the combined sample can be large.
func (s Sample) Ads() (int, error) {
	f, err := os.Open(s.Path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	n, last := 0, byte('\n')
	buf := make([]byte, 64<<10)
	for {
		k, err := f.Read(buf)
		if k > 0 {
			n += bytes.Count(buf[:k], []byte("\n"))
			last = buf[k-1]
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	if last != '\n' {
		n++
	}
	return n, nil
}

// A Timing is how one stage ran on one sample: its time on the clock
// (Wall), its time on the processors (CPU, which is more than Wall when it
// uses several), and the most memory it held at once, in bytes.
type Timing struct {
	Sample, Stage string
	Ads           int
	Wall, CPU     time.Duration
	PeakRSS       int64
}

// rate returns the ads per second, or 0 for a stage too quick to time.
func (t Timing) rate() float64 {
	if t.Wall <= 0 {
		return 0
	}
	return float64(t.Ads) / t.Wall.Seconds()
}

// WriteTimings writes the timings to w as a table for reading: the names
// aligned left, and the figures, of fixed widths, right.
func WriteTimings(w io.Writer, ts []Timing) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "sample\tstage\t    ads\t    wall\t     cpu\t  ads/s\t    peak")
	for _, t := range ts {
		fmt.Fprintf(tw, "%s\t%s\t%7d\t%7.2fs\t%7.2fs\t%7.0f\t%5d MB\n",
			t.Sample, t.Stage, t.Ads, t.Wall.Seconds(), t.CPU.Seconds(), t.rate(), t.PeakRSS>>20)
	}
	return tw.Flush()
}

// WriteTimingsTSV writes the timings to w as tab-separated values, for jq,
// awk or a notebook: seconds and megabytes, with a header.
func WriteTimingsTSV(w io.Writer, ts []Timing) error {
	if _, err := fmt.Fprintln(w, "sample\tstage\tads\twall_s\tcpu_s\tads_per_s\tpeak_mb"); err != nil {
		return err
	}
	for _, t := range ts {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%d\t%.3f\t%.3f\t%.0f\t%d\n",
			t.Sample, t.Stage, t.Ads, t.Wall.Seconds(), t.CPU.Seconds(), t.rate(), t.PeakRSS>>20); err != nil {
			return err
		}
	}
	return nil
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
