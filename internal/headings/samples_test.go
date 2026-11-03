package headings_test

import (
	"testing"

	"github.com/nukopian/jdc-lab/internal/ads"
	"github.com/nukopian/jdc-lab/internal/blocks"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/walls"
)

// sampleAd is an ad of the latest run, with its blocks split.
type sampleAd struct {
	a  ads.Ad
	bs []blocks.Block
}

// sampleAds returns the first n ads of the latest run, their blocks
// split, or nil where there's no run, as in a fresh clone.
func sampleAds(n int) []sampleAd {
	var as []sampleAd
	for a, err := range ads.File("../../data/runs/latest/ads.jsonl") {
		if err != nil {
			return nil
		}
		as = append(as, sampleAd{a, walls.Split(blocks.Split(a.HTML))})
		if len(as) == n {
			break
		}
	}
	return as
}

// sampleTexts returns the title and the text of every part of the first
// n ads of the latest run, or nil where there's no run.
func sampleTexts(n int) []string {
	var texts []string
	for _, x := range sampleAds(n) {
		texts = append(texts, x.a.Title)
		for _, p := range headings.Read(x.bs, x.a.Title, x.a.Locations()...) {
			texts = append(texts, p.Text)
		}
	}
	return texts
}

// BenchmarkKeySamples keys the title and every part of 2,000 ads of the
// latest run. It's skipped where there's no run.
func BenchmarkKeySamples(b *testing.B) {
	texts := sampleTexts(2000)
	if texts == nil {
		b.Skip("no run to read")
	}
	b.ReportAllocs()
	for b.Loop() {
		for _, s := range texts {
			headings.Key(s)
		}
	}
}

// BenchmarkReadSamples reads the parts of 2,000 ads of the latest run,
// their blocks split beforehand. It's skipped where there's no run.
func BenchmarkReadSamples(b *testing.B) {
	as := sampleAds(2000)
	if as == nil {
		b.Skip("no run to read")
	}
	b.ReportAllocs()
	for b.Loop() {
		for _, x := range as {
			headings.Read(x.bs, x.a.Title, x.a.Locations()...)
		}
	}
}
