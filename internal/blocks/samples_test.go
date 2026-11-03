package blocks

import (
	"testing"

	"github.com/nukopian/jdc-lab/internal/ads"
)

// sampleHTML returns the HTML of the first n ads of the latest run, or
// nil where there's no run, as in a fresh clone.
func sampleHTML(n int) []string {
	var html []string
	for a, err := range ads.File("../../data/runs/latest/ads.jsonl") {
		if err != nil {
			return nil
		}
		html = append(html, a.HTML)
		if len(html) == n {
			break
		}
	}
	return html
}

// BenchmarkSplitSamples splits 2,000 ads of the latest run into blocks.
// It's skipped where there's no run.
func BenchmarkSplitSamples(b *testing.B) {
	html := sampleHTML(2000)
	if html == nil {
		b.Skip("no run to read")
	}
	b.ReportAllocs()
	for b.Loop() {
		for _, h := range html {
			Split(h)
		}
	}
}
