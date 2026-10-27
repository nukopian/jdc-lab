package view_test

import (
	"log"
	"os"

	"github.com/nukopian/jdc-lab/internal/ads"
	"github.com/nukopian/jdc-lab/internal/chunk"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
	"github.com/nukopian/jdc-lab/internal/view"
)

// Render shows an ad's chunks, as the chunker made them.
func ExampleRender() {
	ad := ads.Ad{ID: "1", Title: "Driver"}
	ss := []sections.Section{
		{Kind: sections.Summary, Parts: []headings.Part{{Role: headings.Text, Text: "We deliver."}}},
		{Kind: sections.Benefits, Heading: "Benefits", Parts: []headings.Part{
			{Role: headings.Text, Text: "A pension.", Item: true},
			{Role: headings.Text, Text: "A van.", Item: true},
		}},
	}
	c := chunk.Totalled(chunk.Ad{ID: ad.ID, Title: ad.Title, Chunks: chunk.Of(ss, func(string) bool { return false })}, 8)
	if err := view.Render(os.Stdout, ad, c, view.Lost{}); err != nil {
		log.Fatal(err)
	}
	// Output:
	// ## Driver · 1
	//
	// *about 8 tokens: 3 embedded (37%), 0 unclassified (0%)*
	//
	// ### summary · embed · 3 tokens
	//
	// *placed by nothing before it: summary by default*
	//
	// We deliver.
	//
	// ### benefits · “Benefits” · leave out · 7 tokens
	//
	// *placed by the heading "Benefits": "benefits" in headings.txt*
	//
	// - A pension.
	// - A van.
}
