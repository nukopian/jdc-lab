package blocks_test

import (
	"fmt"

	"github.com/nukopian/jdc-lab/internal/blocks"
)

// Split gives the blocks a reader sees, each with its cues: here a bold
// heading, two fields on lines of their own, and a list.
func ExampleSplit() {
	html := `<p><strong>The role</strong></p>` +
		`<p><strong>Hours:</strong> 40 per week<br><strong>Contract:</strong> Permanent</p>` +
		`<ul><li>Pension</li><li>25 days' holiday</li></ul>`
	for _, b := range blocks.Split(html) {
		switch b.Kind {
		case blocks.List:
			fmt.Println(b.Kind, len(b.Items), "items:", b.Items[0].Text, "...")
		default:
			fmt.Printf("%s %q bold=%t lead=%q\n", b.Kind, b.Text, b.Cues.Bold, b.Cues.BoldLead)
		}
	}
	// Output:
	// paragraph "The role" bold=true lead=""
	// line "Hours: 40 per week" bold=false lead="Hours:"
	// line "Contract: Permanent" bold=false lead="Contract:"
	// list 2 items: Pension ...
}

// A Kind prints as its name.
func ExampleKind_String() {
	fmt.Println(blocks.Paragraph, blocks.Line, blocks.List)
	// Output:
	// paragraph line list
}
