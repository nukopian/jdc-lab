package sections

import (
	"fmt"
	"slices"
	"testing"

	"github.com/nukopian/jdc-lab/internal/blocks"
	"github.com/nukopian/jdc-lab/internal/headings"
)

func TestOf(t *testing.T) {
	tests := []struct {
		heading string
		want    Kind // "" for none
	}{
		{"What you’ll be doing", Responsibilities},
		{"Key Responsibilities include:", Responsibilities},
		{"Main Duties", Responsibilities},
		{"Day-to-day", Responsibilities},
		{"Why join us?", Benefits},
		{"What's on Offer", Benefits},
		{"In return for the above you can expect to receive:", Benefits},
		{"Essential Criteria", Requirements},
		{"What We're Looking For", Requirements},
		{"About You", Requirements},
		{"Desirable skills", Desirable},
		{"Skills & Experience", Requirements},
		{"The opportunity", Summary},
		// What the job is, apart from the ad's opening and introductions.
		{"The Role", Role},
		{"Job Description", Role},
		{"Overview", Summary},
		{"About the Role", Role},
		{"About Us", Company},
		{"About Brightway Recruitment", Company},
		{"Why Brightway?", Benefits},
		// A heading starting "why" is a pitch, whatever phrases are in it.
		{"Why You'll Love This Role:", Benefits},
		{"Why Choose Day-to-Day Supply?", Benefits},
		{"Why Should You Apply?", Benefits},
		{"Why consider the role?", Benefits},
		{"How to Apply", Apply},
		{"Interested?", Apply},
		{"Equal Opportunities Statement", Legal},
		{"UK applicants only – no sponsorship available", Legal},
		{"Safeguarding and DBS Compliance", Legal},
		// Phrases of more than one kind: the kind whose rule comes first wins.
		{"What Our Client Offers:", Benefits},
		{"Company Benefits", Benefits},
		{"You will have access to an attractive benefits package including;", Benefits},
		{"Why Join Our Client?", Benefits},
		{"To be successful in this role, you will have:", Requirements},
		{"Company Description", Company},
		{"Day-to-day of the role", Responsibilities},
		{"As a Teaching Assistant, you will:", Responsibilities},
		{"You will have:", Requirements},
		{"You'll be someone who:", Requirements},
		{"JOB PURPOSE:", Role},
		// Details: its phrases, then any field's label.
		{"Job Details", Details},
		{"Location & travel", Details},
		{"Salary |", Details},
		{"Salary & Benefits", Benefits},
		{"Working Days", Details},
		// Shifts, and a shift's name heading its own hours.
		{"Shifts Available:", Details},
		{"Night Shift", Details},
		{"Rotation:", Details},
		{"Nights:", Details},
		{"As a Shift Engineer, you will:", Responsibilities},
		{"Shift Manager - About the Role", Role},
		{"Tasks", Responsibilities},
		{"Candidates will also require:", Requirements},
		{"What's in it for me?", Benefits},
		// From the second sample.
		{"WHAT IS IN IT FOR YOU?", Benefits},
		{"What you can expect:", Benefits},
		{"Who thrives here?", Requirements},
		{"What to expect", Responsibilities},
		{"THE BUSINESS", Company},
		{"Our Business", Company},
		{"The School:", Company},
		{"The school is looking for someone who:", Requirements},
		{"The school would like to hear from teachers who:", Requirements},
		{"The school is seeking an SEN Teacher who has:", Requirements},
		{"What the School Offers:", Benefits},
		{"Main Purpose of Job", Role},
		{"You must have:", Requirements},
		{"What We Need From You", Requirements},
		{"Personal Qualities", Requirements},
		{"Licence/Certification:", Requirements},
		{"To be eligible for this role, the potential 1:1 Tutor must:", Requirements},
		{"Recruitment Process", Apply},
		{"What you need:", Requirements},
		{"What you need to do now", Apply},
		// Legal's name isn't one of its phrases: in a heading, it's more often a
		// job's field than a notice.
		{"Legal Information:", Legal},
		{"The Legal Bit:", Legal},
		{"As a Temporary Legal Secretary, your day-to-day tasks will include:", Responsibilities},
		{"About Brightway Legal Recruitment", Company},
		{"Temp Legal Secretary", ""},
		{"Software Development", ""},
		{"Your impact", ""},
	}
	for _, tt := range tests {
		got, ok := Of(headings.Key(tt.heading))
		if got != tt.want || ok != (tt.want != "") {
			t.Errorf("Of(%q) = %q, %t; want %q", tt.heading, got, ok, tt.want)
		}
	}
}

func TestAll(t *testing.T) {
	tests := []struct {
		heading string
		want    []Kind
	}{
		{"About you, and what's on offer", []Kind{Requirements, Benefits}},
		// Phrases that overlap both count: "in this role you will", and
		// "you will need" past its end.
		{"In this role you will need", []Kind{Responsibilities, Requirements}},
		// A phrase wholly inside a longer one doesn't: "profile" is
		// requirements, but here it's part of "company profile".
		{"Company Profile", []Kind{Company}},
	}
	for _, tt := range tests {
		if got := All(headings.Key(tt.heading)); !slices.Equal(got, tt.want) {
			t.Errorf("All(%q) = %q, want %q", tt.heading, got, tt.want)
		}
	}
}

func TestNamed(t *testing.T) {
	for text, want := range map[string]Kind{
		"Responsibilities":              Responsibilities,
		"Benefits":                      Benefits,
		"Our responsibilities to you":   "",
		"We offer a competitive salary": "",
	} {
		got, ok := Named(headings.Key(text))
		if got != want || ok != (want != "") {
			t.Errorf("Named(%q) = %q, %t; want %q", text, got, ok, want)
		}
	}
}

// TestRules checks headings.txt: every kind but Unmapped has phrases, and
// no phrase means two kinds.
func TestRules(t *testing.T) {
	if amb := phrases.Ambiguous(); len(amb) > 0 {
		t.Errorf("phrases with more than one kind: %v", amb)
	}
	// That every kind has a rule, init checks. A rule written with =>
	// leaves out its kind's name.
	if k, ok := Named("legal"); ok {
		t.Errorf(`"legal" is a phrase, of %q`, k)
	}
	if k, ok := Named("benefits"); !ok || k != Benefits {
		t.Errorf(`"benefits" = %q, %t; want benefits`, k, ok)
	}
}

func TestDivide(t *testing.T) {
	html := `<p>We're a friendly team.</p>` +
		`<p><strong>Key Responsibilities</strong></p><ul><li>Drive</li></ul>` +
		`<p><strong>Software Development</strong></p><p>Some Go.</p>` +
		`<p><strong>Salary:</strong> £30,000</p>` +
		`<p>Benefits</p><ul><li>Pension</li></ul>`
	var got []string
	for _, s := range Divide(headings.Read(blocks.Split(html), "Driver")) {
		line := fmt.Sprintf("%s %q:", s.Kind, s.Heading)
		if s.ByName {
			line = "by name " + line
		}
		for _, p := range s.Parts {
			line += fmt.Sprintf(" %s(%s)", p.Role, p.Text)
		}
		got = append(got, line)
	}
	want := []string{
		`summary "": text(We're a friendly team.)`,
		`responsibilities "Key Responsibilities": text(Drive)`,
		`unmapped "Software Development": text(Some Go.) field(Salary: £30,000)`,
		`by name benefits "Benefits": text(Pension)`,
	}
	if !slices.Equal(got, want) {
		t.Errorf("\n got: %q\nwant: %q", got, want)
	}
}

func TestDivideWithNoSummary(t *testing.T) {
	ss := Divide(headings.Read(blocks.Split(`<p><b>About us</b></p><p>We care.</p>`), ""))
	if len(ss) != 1 || ss[0].Kind != Company {
		t.Errorf("got %+v, want one company section", ss)
	}
}

func TestExplain(t *testing.T) {
	for key, want := range map[string]string{
		"why join us":          `it starts "why"`,
		"what you'll be doing": `"what you'll be doing" in headings.txt`,
		"desirable skills":     `"desirable" in headings.txt, over "skills" (requirements)`,
		"about brightway":      `it starts "about"`,
		"working hours":        `"working hours" in headings.txt`,
		"our depots":           "no phrase in headings.txt",
	} {
		if got := Explain(key); got != want {
			t.Errorf("Explain(%q) = %q, want %q", key, got, want)
		}
	}
}

// A short plain line with a section's phrase, followed by a list, starts
// that section: a sub-heading such as "Highly Desirable" among the
// requirements, or "Key Skills". Not followed by a list, or longer than a
// heading's six words, or a sentence, it's text.
func TestDivideLineBeforeAList(t *testing.T) {
	html := `<p><strong>What We're Looking For</strong></p>` +
		`<p>Essential Experience &amp; Qualifications</p><ul><li>Five years in post</li></ul>` +
		`<p>Highly Desirable</p><ul><li>Higher Education</li></ul>` +
		`<p>Desirable but not essential.</p><ul><li>A van</li></ul>` +
		`<p>Key Skills</p><ul><li>Excel</li></ul>` +
		`<p>Further desirable things we would love you to have</p><ul><li>A boat</li></ul>` +
		`<p>Desirable</p><p>Not a list after it, but the line's a section's name.</p>`
	var got []string
	for _, s := range Divide(headings.Read(blocks.Split(html), "Manager")) {
		line := fmt.Sprintf("%s %q:", s.Kind, s.Heading)
		for _, p := range s.Parts {
			line += fmt.Sprintf(" %s(%s)", p.Role, p.Text)
		}
		got = append(got, line)
	}
	want := []string{
		`requirements "What We're Looking For":`,
		`requirements "Essential Experience & Qualifications": text(Five years in post)`,
		`desirable "Highly Desirable": text(Higher Education) text(Desirable but not essential.) text(A van)`,
		`requirements "Key Skills": text(Excel) text(Further desirable things we would love you to have) text(A boat)`,
		`desirable "Desirable": text(Not a list after it, but the line's a section's name.)`,
	}
	if !slices.Equal(got, want) {
		t.Errorf("\n got: %q\nwant: %q", got, want)
	}
}
