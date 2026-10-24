package sections_test

import (
	"fmt"

	"github.com/nukopian/jdc-lab/internal/blocks"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
)

// Divide gives an ad's sections: what comes before the first heading is the
// summary, and each heading of a known kind starts a section.
func ExampleDivide() {
	html := `<p>Join our friendly team in Leeds.</p>` +
		`<p><strong>What you'll be doing</strong></p><ul><li>Deliver parcels</li></ul>` +
		`<p><strong>Why join us?</strong></p><ul><li>Pension</li><li>Free parking</li></ul>`
	for _, s := range sections.Divide(headings.Read(blocks.Split(html), "Driver")) {
		fmt.Printf("%s %q: %d parts\n", s.Kind, s.Heading, len(s.Parts))
	}
	// Output:
	// summary "": 1 parts
	// responsibilities "What you'll be doing": 1 parts
	// benefits "Why join us?": 2 parts
}

// Of finds the section a heading starts from the phrases in it.
func ExampleOf() {
	for _, h := range []string{"Key Responsibilities include:", "What Our Client Offers", "Your impact"} {
		kind, ok := sections.Of(headings.Key(h))
		fmt.Printf("%q: %q %t\n", h, kind, ok)
	}
	// Output:
	// "Key Responsibilities include:": "responsibilities" true
	// "What Our Client Offers": "benefits" true
	// "Your impact": "" false
}

// All lists every kind a heading's phrases point to, for finding headings
// whose words disagree.
func ExampleAll() {
	fmt.Println(sections.All(headings.Key("What Our Client Offers")))
	// Output:
	// [company benefits]
}

// Named recognises text with no cue as a heading only if it's exactly a
// section's name.
func ExampleNamed() {
	for _, text := range []string{"Responsibilities", "Our responsibilities to you"} {
		kind, ok := sections.Named(headings.Key(text))
		fmt.Printf("%q: %q %t\n", text, kind, ok)
	}
	// Output:
	// "Responsibilities": "responsibilities" true
	// "Our responsibilities to you": "" false
}
