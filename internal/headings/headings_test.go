package headings

import (
	"fmt"
	"slices"
	"testing"

	"github.com/nukopian/jdc-lab/internal/blocks"
)

func TestKey(t *testing.T) {
	nbsp := string(rune(0xA0))
	tests := []struct{ text, want string }{
		{"What You’ll Be Doing:", "what you'll be doing"},
		{"What you'll be doing", "what you'll be doing"},
		{"Key Responsibilities;", "key responsibilities"},
		{"✅ Benefits ✅", "benefits"},
		{"Skills & Experience", "skills & experience"},
		{"Full-time / Part-time", "full time / part time"},
		{"Why" + nbsp + "join" + nbsp + "us?", "why join us"},
		{"• Pension", "pension"},
		{string(rune(0xF0B7)) + "\tPension", "pension"},
		{"✔️ Holiday", "holiday"},
		{"About   the\n role", "about the role"},
		{"Café Société", "café société"},
		{"📢📢", ""},
	}
	for _, tt := range tests {
		if got := Key(tt.text); got != tt.want {
			t.Errorf("Key(%q) = %q, want %q", tt.text, got, tt.want)
		}
	}
}

// show writes parts as one line each, for comparing: the role, the text,
// and the key or value where there is one.
func show(parts []Part) []string {
	var out []string
	for _, p := range parts {
		s := fmt.Sprintf("%s %q", p.Role, p.Text)
		if p.Item {
			s = "item " + s
		}
		if p.Key != "" {
			s += fmt.Sprintf(" key=%q", p.Key)
		}
		if p.Value != "" {
			s += fmt.Sprintf(" value=%q", p.Value)
		}
		out = append(out, s)
	}
	return out
}

func TestRead(t *testing.T) {
	tests := []struct {
		name string
		html string
		want []string
	}{
		{
			"a bold heading",
			`<p><strong>The Role</strong></p><p>We build things.</p>`,
			[]string{`heading "The Role" key="the role"`, `text "We build things."`},
		},
		{
			"a heading tag",
			`<h2>About us</h2>`,
			[]string{`heading "About us" key="about us"`},
		},
		{
			"a plain heading ending with a colon or semicolon",
			`<p>Requirements:</p><p>Benefits;</p>`,
			[]string{`heading "Requirements:" key="requirements"`, `heading "Benefits;" key="benefits"`},
		},
		{
			"a plain short line without a colon is text",
			`<p>Responsibilities</p>`,
			[]string{`text "Responsibilities"`},
		},
		{
			"fields, in bold or not, on lines or in a list",
			`<p><strong>Salary:</strong> £30,000<br>Hours: 40 per week</p><ul><li>Contract type: Permanent</li></ul>`,
			[]string{
				`field "Salary: £30,000" key="salary" value="£30,000"`,
				`field "Hours: 40 per week" key="hours" value="40 per week"`,
				`item field "Contract type: Permanent" key="contract type" value="Permanent"`,
			},
		},
		{
			"a label that isn't a field's",
			`<ul><li><strong>Financial reporting:</strong> Oversee the accounts.</li></ul>`,
			[]string{`item text "Financial reporting: Oversee the accounts."`},
		},
		{
			"a label with no value is a heading",
			`<p><strong>Location:</strong></p>`,
			[]string{`heading "Location:" key="location"`},
		},
		{
			"a label as a heading, its value on the next line",
			`<p><strong>Working Hours</strong></p><p>Monday to Friday, 9am to 5pm</p><p><b>Location</b> Leeds</p>`,
			[]string{
				`field "Working Hours Monday to Friday, 9am to 5pm" key="working hours" value="Monday to Friday, 9am to 5pm"`,
				`field "Location Leeds" key="location" value="Leeds"`,
			},
		},
		{
			"a label as a heading, followed by a paragraph",
			`<p><b>Location:</b></p><p>Our office is in the centre of Leeds, a short walk from the station, with parking on site for all staff.</p>`,
			[]string{
				`heading "Location:" key="location"`,
				`text "Our office is in the centre of Leeds, a short walk from the station, with parking on site for all staff."`,
			},
		},
		{
			"a bar for a colon",
			`<p>Salary | £30,000</p>`,
			[]string{`field "Salary | £30,000" key="salary" value="£30,000"`},
		},
		{
			"a bold heading run into its text",
			`<p><strong>Key Responsibilities:</strong> You will lead the team.</p>`,
			[]string{`heading "Key Responsibilities:" key="key responsibilities"`, `text "You will lead the team."`},
		},
		{
			"an emoji in bold is no heading",
			`<p><strong>✏️ </strong>Enjoy flexibility.</p>`,
			[]string{`text "✏️ Enjoy flexibility."`},
		},
		{
			"repeats of the title and location",
			`<p><strong>Ecologist</strong></p><p><strong>Cambridge</strong></p><p><b>Leeds</b></p>`,
			[]string{`repeat "Ecologist"`, `repeat "Cambridge"`, `heading "Leeds" key="leeds"`},
		},
		{
			"bold pay, figures and sentences are no headings",
			`<p><b>£30,000 per annum</b></p><p><b>Ref 225682916</b></p><p><b>Apply today.</b></p><p><b>1</b></p>`,
			[]string{`text "£30,000 per annum"`, `text "Ref 225682916"`, `text "Apply today."`, `text "1"`},
		},
		{
			"a bold question, and a trailing ellipsis, are headings",
			`<p><b>Why join us?</b></p><p><b>And there's more...</b></p>`,
			[]string{`heading "Why join us?" key="why join us"`, `heading "And there's more..." key="and there's more"`},
		},
		{
			"long bold text is a bold sentence",
			`<p><b>We are looking for an experienced and motivated engineer to join our team</b></p>`,
			[]string{`text "We are looking for an experienced and motivated engineer to join our team"`},
		},
		{
			"bold items are text",
			`<ul><li><b>Benefits</b></li><li>Pension</li></ul>`,
			[]string{`item text "Benefits"`, `item text "Pension"`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := show(Read(blocks.Split(tt.html), "Ecologist", "Cambridge, Cambridgeshire", "Cambridge"))
			if !slices.Equal(got, tt.want) {
				t.Errorf("Read(%q)\n got: %q\nwant: %q", tt.html, got, tt.want)
			}
		})
	}
}

func TestRoleString(t *testing.T) {
	for r, want := range map[Role]string{Text: "text", Heading: "heading", Field: "field", Repeat: "repeat", 9: "unknown"} {
		if got := r.String(); got != want {
			t.Errorf("Role(%d).String() = %q, want %q", int(r), got, want)
		}
	}
}
