package headings_test

import (
	"fmt"

	"github.com/nukopian/jdc-lab/internal/blocks"
	"github.com/nukopian/jdc-lab/internal/headings"
)

// Read gives each block a role: here the title repeated, a field, a heading
// run into its text, and a heading on its own.
func ExampleRead() {
	html := `<p><strong>Delivery Driver</strong></p>` +
		`<p><strong>Salary:</strong> £30,000</p>` +
		`<p><strong>The role:</strong> You'll deliver parcels.</p>` +
		`<p><strong>Benefits</strong></p><ul><li>Pension</li></ul>`
	for _, p := range headings.Read(blocks.Split(html), "Delivery Driver") {
		fmt.Printf("%-8s %-28q key=%q value=%q\n", p.Role, p.Text, p.Key, p.Value)
	}
	// Output:
	// repeat   "Delivery Driver"            key="" value=""
	// field    "Salary: £30,000"            key="salary" value="£30,000"
	// heading  "The role:"                  key="the role" value=""
	// text     "You'll deliver parcels."    key="" value=""
	// heading  "Benefits"                   key="benefits" value=""
	// text     "Pension"                    key="" value=""
}

// Key is the form to match on: the same for headings written differently.
func ExampleKey() {
	fmt.Println(headings.Key("What You’ll Be Doing:"))
	fmt.Println(headings.Key("✅ what you'll be doing"))
	// Output:
	// what you'll be doing
	// what you'll be doing
}

// A Role prints as its name.
func ExampleRole_String() {
	fmt.Println(headings.Text, headings.Heading, headings.Field, headings.Repeat)
	// Output:
	// text heading field repeat
}
