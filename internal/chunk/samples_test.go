package chunk

import (
	"reflect"
	"testing"

	"github.com/nukopian/jdc-lab/internal/ads"
	"github.com/nukopian/jdc-lab/internal/blocks"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
	"github.com/nukopian/jdc-lab/internal/walls"
)

// sampleAd is an ad of the latest run, with its sections.
type sampleAd struct {
	a  ads.Ad
	ss []sections.Section
}

// sampleAds returns the first n ads of the latest run, their sections
// divided, or nil where there's no run, as in a fresh clone.
func sampleAds(n int) []sampleAd {
	var as []sampleAd
	for a, err := range ads.File("../../data/runs/latest/ads.jsonl") {
		if err != nil {
			return nil
		}
		as = append(as, sampleAd{a, sections.Divide(headings.Read(walls.Split(blocks.Split(a.HTML)), a.Title, a.Locations()...))})
		if len(as) == n {
			break
		}
	}
	return as
}

// BenchmarkAdOfSamples chunks 2,000 ads of the latest run, their
// sections divided beforehand, so that it times the chunker alone. It's
// skipped where there's no run.
func BenchmarkAdOfSamples(b *testing.B) {
	as := sampleAds(2000)
	if as == nil {
		b.Skip("no run to read")
	}
	none := func(string) bool { return false }
	b.ReportAllocs()
	for b.Loop() {
		for _, x := range as {
			AdOf(x.a.ID, x.a.Title, x.ss, none)
		}
	}
}

// Reading each line's facts once per ad changes no chunk: checked over
// 2,000 ads of the latest run, where there's one.
func TestReadOnceChangesNothing(t *testing.T) {
	if testing.Short() {
		t.Skip("short")
	}
	as := sampleAds(2000)
	if as == nil {
		t.Skip("no run to read")
	}
	odd := func(s string) bool { return len(s)%3 == 0 } // some text boilerplate, to reach the guess's every branch
	for _, x := range as {
		on := AdOf(x.a.ID, x.a.Title, x.ss, odd)
		readOnce = false
		off := AdOf(x.a.ID, x.a.Title, x.ss, odd)
		readOnce = true
		if !reflect.DeepEqual(on, off) {
			t.Fatalf("ad %s: reading once\n%+v\nnot\n%+v", x.a.ID, on, off)
		}
	}
}
