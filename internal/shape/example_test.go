package shape_test

import (
	"os"

	"github.com/nukopian/jdc-lab/internal/ads"
	"github.com/nukopian/jdc-lab/internal/shape"
)

// Write shows an ad at the stages asked for: here its blocks, the role
// each part has, and its sections.
func ExampleWrite() {
	a := ads.Ad{ID: "1", Title: "Driver", HTML: `<p><strong>The role</strong></p><p>Deliver parcels.</p>` +
		`<p><strong>Salary:</strong> £30,000</p><p>Requirements:</p><ul><li>A licence</li></ul>`}
	notBoilerplate := func(string) bool { return false }
	if err := shape.Write(os.Stdout, a, notBoilerplate, "blocks", "headings", "sections"); err != nil {
		panic(err)
	}
	// Output:
	// Driver · 1
	//
	// ## blocks: 5, as a reader sees them, with their cues
	//
	//   1 paragraph The role  [bold]
	//   2 paragraph Deliver parcels.
	//   3 paragraph Salary: £30,000  [bold lead "Salary:"]
	//   4 paragraph Requirements:
	//   5 list of 1
	//       - A licence
	//
	// ## headings: 5 parts, each with its role
	//
	//   1 heading    The role  key "the role"
	//   2 text       Deliver parcels.
	//   3 field      Salary: £30,000  key "salary"  value "£30,000"
	//   4 heading    Requirements:  key "requirements"
	//   5 text     - A licence
	//
	// ## sections: 2, each a heading and what follows it
	//
	// [1] role · "The role"
	//     | Deliver parcels.
	//     field Salary: £30,000
	//
	// [2] requirements · "Requirements:"
	//     | - A licence
}
