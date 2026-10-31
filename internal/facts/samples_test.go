package facts

import (
	"testing"

	"github.com/nukopian/jdc-lab/internal/ads"
	"github.com/nukopian/jdc-lab/internal/blocks"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/walls"
)

// sampleLinesOrNil returns the text of every part of the first n ads of
// the latest run, as the chunker gives them to Find, or nil where there's
// no run, as in a fresh clone.
func sampleLinesOrNil(n int) []string {
	var lines []string
	for a, err := range ads.File("../../data/runs/latest/ads.jsonl") {
		if err != nil {
			return nil
		}
		for _, p := range headings.Read(walls.Split(blocks.Split(a.HTML)), a.Title, a.Locations()...) {
			lines = append(lines, p.Text)
		}
		if n--; n == 0 {
			break
		}
	}
	return lines
}

// sampleLines is sampleLinesOrNil, skipping b where there's no run.
func sampleLines(b *testing.B, n int) []string {
	lines := sampleLinesOrNil(n)
	if lines == nil {
		b.Skip("no run to read")
	}
	return lines
}

// BenchmarkFindSamples runs Find over every line of 2,000 ads of the
// latest run.
func BenchmarkFindSamples(b *testing.B) {
	lines := sampleLines(b, 2000)
	b.ReportAllocs()
	for b.Loop() {
		for _, l := range lines {
			Find(l)
		}
	}
}
