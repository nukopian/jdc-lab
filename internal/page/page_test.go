package page

import (
	"strings"
	"testing"

	"github.com/nukopian/jdc-lab/internal/ads"
	"github.com/nukopian/jdc-lab/internal/chunk"
	"github.com/nukopian/jdc-lab/internal/sections"
)

func TestPosted(t *testing.T) {
	for in, want := range map[string]string{
		`<p class="x" onclick="evil()">Hi <b>there</b></p>`:     `<p>Hi <b>there</b></p>`,
		`<script>alert(1)</script><p>Text</p>`:                  `<p>Text</p>`,
		`<style>p { color: red }</style><a href="x">a link</a>`: `a link`,
		`<p>5 &lt; 6 &amp; <img src=x onerror=y>more</p>`:       `<p>5 &lt; 6 &amp; more</p>`,
		`Line one<br/>Line two`:                                 `Line one<br>Line two`,
	} {
		if got := string(Posted(in)); got != want {
			t.Errorf("Posted(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestWrite(t *testing.T) {
	c := chunk.Ad{ID: "1", Title: "Driver", Tokens: 20, Chunks: []chunk.Chunk{
		{Kind: sections.Summary, Use: chunk.Embed, Text: "We <deliver> parcels.", Tokens: 12, PlacedBy: "nothing before it"},
		{Kind: chunk.Boilerplate, Use: chunk.LeaveOut, Text: "summary: Apply now", Tokens: 8, PlacedBy: "found as boilerplate"},
	}}
	var b strings.Builder
	if err := Write(&b, Ad{Ad: ads.Ad{ID: "1", Title: "Driver", HTML: "<p>We deliver</p>"}, Chunks: c, Why: "a short ad"}, "Before and after"); err != nil {
		t.Fatal(err)
	}
	page := b.String()
	for _, want := range []string{
		`We &lt;deliver&gt; parcels.`, // the chunk's text, escaped
		`<span class="from">summary</span> Apply now`,
		`embedded 12 (60%)`, `boilerplate 8 (40%)`,
		`<div class="ad"><p>We deliver</p></div>`,
		`a short ad`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("page lacks %q", want)
		}
	}
	b.Reset()
	if err := WriteIndex(&b, "Before and after", []Entry{{Ad: Ad{Ad: ads.Ad{ID: "1", Title: "Driver"}, Chunks: c}, File: "1.html"}}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), `<a href="1.html">Driver</a>`) || !strings.Contains(b.String(), "width: 60%") {
		t.Errorf("index lacks the ad's link or its bar:\n%s", b.String())
	}
}
