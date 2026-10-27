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
			"labels from the second sample",
			`<p><b>Vacancy Type:</b> Permanent, Full Time</p><p>Working Arrangement: Hybrid</p><p><b>Reports To:</b> Team Leader</p><p><b>Hybrid Working:</b></p><p>3 days onsite, 2 from home</p>`,
			[]string{
				`field "Vacancy Type: Permanent, Full Time" key="vacancy type" value="Permanent, Full Time"`,
				`field "Working Arrangement: Hybrid" key="working arrangement" value="Hybrid"`,
				`field "Reports To: Team Leader" key="reports to" value="Team Leader"`,
				`field "Hybrid Working: 3 days onsite, 2 from home" key="hybrid working" value="3 days onsite, 2 from home"`,
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
			[]string{`text "Enjoy flexibility."`},
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
			"bold times are no headings",
			`<p><b>- 2pm to 10pm</b></p><p><b>Monday to Friday, 8:00am - 5:00pm</b></p>` +
				`<p><b>08:00 starts</b></p><p><b>Monday to Friday - 8.30-3.30</b></p>`,
			[]string{`text "- 2pm to 10pm"`, `text "Monday to Friday, 8:00am - 5:00pm"`,
				`text "08:00 starts"`, `text "Monday to Friday - 8.30-3.30"`},
		},
		{
			// A bare range of hours, "8-5", is a time only beside a range of
			// days, which M-F is too; with pm after it, am is implied.
			"bold bare hours beside days are no headings",
			`<p><b>Monday - Friday 8-5</b></p><p><b>M-F 8-4</b></p><p><b>Mon–Thu 7–3:30</b></p>` +
				`<p><b>9-5pm</b></p><p><b>8-5 Monday - Friday</b></p>` +
				// A range that runs back round the clock is no count, whatever
				// follows it; a word that only starts as a unit is no unit.
				`<p><b>Mon - Fri 7-3 Days</b></p><p><b>Monday to Friday 6-2 Nights</b></p>` +
				`<p><b>Monday - Friday 9-5 Minimum</b></p><p><b>Mon-Fri 8 till 5</b></p>`,
			[]string{`text "Monday - Friday 8-5"`, `text "M-F 8-4"`, `text "Mon-Thu 7-3:30"`,
				`text "9-5pm"`, `text "8-5 Monday - Friday"`,
				`text "Mon - Fri 7-3 Days"`, `text "Monday to Friday 6-2 Nights"`,
				`text "Monday - Friday 9-5 Minimum"`, `text "Mon-Fri 8 till 5"`},
		},
		{
			// A rota, or days or nights beside a range of days, is a pattern
			// of work, a fact as a time is.
			"bold shift patterns are no headings",
			`<p><b>4 on 4 off (days)</b></p><p><b>4on 3off</b></p><p><b>4 days on, 4 days off</b></p>` +
				`<p><b>Monday - Friday (Days)</b></p><p><b>Mon- Fri Nights</b></p>` +
				`<p><b>4 on - 4 off</b></p><p><b>4ON4OFF</b></p>`,
			[]string{`text "4 on 4 off (days)"`, `text "4on 3off"`, `text "4 days on, 4 days off"`,
				`text "Monday - Friday (Days)"`, `text "Mon- Fri Nights"`,
				`text "4 on - 4 off"`, `text "4ON4OFF"`},
		},
		{
			// Alone, a shift's name heads its own hours; with a figure,
			// "days" is a count; and a pattern ending in a colon heads
			// what follows.
			"bold shift names are headings",
			`<p><b>Night Shift</b></p><p><b>Days:</b></p><p><b>Shift Pattern</b></p>` +
				`<p><b>You will work a 4 on 4 off shift pattern:</b></p><p><b>Monday to Friday Nights:</b></p>`,
			[]string{`heading "Night Shift" key="night shift"`, `heading "Days:" key="days"`,
				`heading "Shift Pattern" key="shift pattern"`,
				`heading "You will work a 4 on 4 off shift pattern:" key="you will work a 4 on 4 off shift pattern"`,
				`heading "Monday to Friday Nights:" key="monday to friday nights"`},
		},
		{
			"bold bare ranges without days are headings",
			`<p><b>Ages 5-11</b></p><p><b>Monday to Friday, 2-3 days in the office</b></p>` +
				`<p><b>Mon-Sun, 10-12 hr shifts</b></p><p><b>Sunday to Thursday, 4-5 nights</b></p>` +
				// Places aren't days, however they start.
				`<p><b>Monmouth to Sunderland, 2-4 stops</b></p>`,
			[]string{`heading "Ages 5-11" key="ages 5 11"`,
				`heading "Monday to Friday, 2-3 days in the office" key="monday to friday 2 3 days in the office"`,
				`heading "Mon-Sun, 10-12 hr shifts" key="mon sun 10 12 hr shifts"`,
				`heading "Sunday to Thursday, 4-5 nights" key="sunday to thursday 4 5 nights"`,
				`heading "Monmouth to Sunderland, 2-4 stops" key="monmouth to sunderland 2 4 stops"`},
		},
		{
			// A figure with a point is a time only with am or pm, or in a
			// range: 37.50 is hours, not half past.
			"bold figures with a point aren't times",
			`<p><b>37.50 hours per week</b></p><p><b>Top 10 reasons to join</b></p>`,
			[]string{`heading "37.50 hours per week" key="37 50 hours per week"`,
				`heading "Top 10 reasons to join" key="top 10 reasons to join"`},
		},
		{
			"a bold question, and a trailing ellipsis, are headings",
			`<p><b>Why join us?</b></p><p><b>And there's more...</b></p>`,
			[]string{`heading "Why join us?" key="why join us"`, `heading "And there's more..." key="and there's more"`},
		},
		{
			"lone conjunctions between items are no headings",
			`<ul><li>A degree</li></ul><p><b>OR</b></p><p><b>Either:</b></p><p><b>Plus:</b></p><p><b>And &amp; or</b></p><p><b>And/or</b></p><p><b>Or call:</b></p>`,
			[]string{
				`item text "A degree"`, `text "OR"`, `text "Either:"`, `text "Plus:"`, `text "And & or"`, `text "And/or"`,
				`heading "Or call:" key="or call"`,
			},
		},
		{
			"long bold text is a bold sentence",
			`<p><b>We are looking for an experienced and motivated engineer to join our team</b></p>`,
			[]string{`text "We are looking for an experienced and motivated engineer to join our team"`},
		},
		{
			"a long bold line ending in a colon introduces what follows",
			`<p><b>As a Mobile Mechanic with us, you'll get benefits that go the extra mile:</b></p><p>As a Mobile Mechanic with us, you'll get benefits that go the extra mile:</p><p><b>As a Mobile Mechanic with us, you'll get benefits that go the extra mile and more besides.</b></p>`,
			[]string{
				`heading "As a Mobile Mechanic with us, you'll get benefits that go the extra mile:" key="as a mobile mechanic with us you'll get benefits that go the extra mile"`,
				`text "As a Mobile Mechanic with us, you'll get benefits that go the extra mile:"`,
				`text "As a Mobile Mechanic with us, you'll get benefits that go the extra mile and more besides."`,
			},
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

func TestIsLabel(t *testing.T) {
	for key, want := range map[string]bool{"salary": true, "working hours": true, "the role": false, "Salary": false} {
		if got := IsLabel(key); got != want {
			t.Errorf("IsLabel(%q) = %t, want %t", key, got, want)
		}
	}
}

func TestRoleString(t *testing.T) {
	for r, want := range map[Role]string{Text: "text", Heading: "heading", Field: "field", Repeat: "repeat", 9: "unknown"} {
		if got := r.String(); got != want {
			t.Errorf("Role(%d).String() = %q, want %q", int(r), got, want)
		}
	}
}

// A label cleaning added for a symbol makes a field, and the part says
// what was added.
func TestReadAddedLabel(t *testing.T) {
	parts := Read(blocks.Split(`<p>📍 Leeds</p><ul><li>💷 £30,000</li></ul>`), "Driver")
	if len(parts) != 2 {
		t.Fatalf("got %+v", parts)
	}
	for i, want := range []struct{ key, value, added string }{{"location", "Leeds", "Location: "}, {"pay", "£30,000", "Pay: "}} {
		p := parts[i]
		if p.Role != Field || p.Key != want.key || p.Value != want.value || p.Added != want.added {
			t.Errorf("part %d: got %+v, want a %s field of %q, added %q", i, p, want.key, want.value, want.added)
		}
	}
}

// A typographic dash keys as a hyphen does, so a title as written and the
// ad's text, cleaned, still match.
func TestKeyDashes(t *testing.T) {
	if a, b := Key("Teacher \u2013 Leeds"), Key("Teacher - Leeds"); a != b {
		t.Errorf("Key(en dash) = %q, Key(hyphen) = %q; want the same", a, b)
	}
	if a, b := Key("Mon\u2013Fri"), Key("Mon-Fri"); a != b {
		t.Errorf("Key(en dash) = %q, Key(hyphen) = %q; want the same", a, b)
	}
}
