package view_test

import (
	"log"
	"os"

	"github.com/nukopian/jdc-lab/internal/ads"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
	"github.com/nukopian/jdc-lab/internal/view"
)

// Render shows an ad's sections, with what each part is.
func ExampleRender() {
	ad := ads.Ad{ID: "1", Title: "Driver"}
	ss := []sections.Section{
		{Kind: sections.Summary, Parts: []headings.Part{{Role: headings.Text, Text: "We deliver."}}},
		{Kind: sections.Benefits, Heading: "Benefits", Parts: []headings.Part{
			{Role: headings.Text, Text: "A pension.", Item: true},
			{Role: headings.Text, Text: "A van.", Item: true},
		}},
	}
	own := func(string) view.Verdict { return view.Own }
	if err := view.Render(os.Stdout, ad, ss, own, view.Lost{}); err != nil {
		log.Fatal(err)
	}
	// Output:
	// ## Driver · 1
	//
	// ### summary · no heading, placed by default
	//
	// We deliver.
	//
	// ### benefits · “Benefits”
	//
	// - A pension.
	// - A van.
}
