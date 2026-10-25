package runs_test

import (
	"fmt"

	"github.com/nukopian/jdc-lab/internal/runs"
)

// Changed pairs two summaries' lines by what they say, and gives those
// whose figures moved.
func ExampleChanged() {
	before := "headings of a known kind:  5230 of 6913 uses (76%)\nads with only a summary:  96 (5%)\n"
	after := "headings of a known kind:  5301 of 6913 uses (77%)\nads with only a summary:  96 (5%)\n"
	for _, c := range runs.Changed(before, after) {
		fmt.Printf("%s\n  → %s\n", c.Before, c.After)
	}
	// Output:
	// headings of a known kind:  5230 of 6913 uses (76%)
	//   → headings of a known kind:  5301 of 6913 uses (77%)
}

// Moved gives the headings a change mapped and those it unmapped.
func ExampleMoved() {
	before := map[string]runs.Heading{
		"and":         {Key: "and", Ads: 39, Chars: 81961},
		"please note": {Key: "please note", Ads: 24, Chars: 6872},
	}
	after := map[string]runs.Heading{
		"please note": {Key: "please note", Ads: 24, Chars: 6872},
	}
	mapped, unmapped := runs.Moved(before, after)
	for _, h := range mapped {
		fmt.Printf("mapped %q: %d ads, %d chars\n", h.Key, h.Ads, h.Chars)
	}
	fmt.Println(len(unmapped), "unmapped")
	// Output:
	// mapped "and": 39 ads, 81961 chars
	// 0 unmapped
}
