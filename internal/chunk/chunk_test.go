package chunk

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/nukopian/jdc-lab/internal/facts"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
	"go.yaml.in/yaml/v3"
)

func text(s string) headings.Part { return headings.Part{Role: headings.Text, Text: s} }

func TestOf(t *testing.T) {
	ss := []sections.Section{
		{Kind: sections.Summary, Parts: []headings.Part{text("We deliver parcels across the north of England every day.")}},
		{Kind: sections.Responsibilities, Heading: "What you'll be doing", Parts: []headings.Part{text("Drive the van."), text("We are an agency.")}},
		{Kind: sections.Details, Heading: "Job details", Parts: []headings.Part{
			{Role: headings.Field, Text: "Salary: £30,000", Key: "salary", Value: "£30,000"},
		}},
		{Kind: sections.Company, Heading: "About Brightway", Parts: []headings.Part{text("We are a family firm.")}},
		{Kind: sections.Apply, Heading: "How to apply", ByName: true, Parts: []headings.Part{text("Apply today.")}},
	}
	boiler := func(s string) bool { return s == "We are an agency." }
	cs := Of(ss, boiler)
	if len(cs) != 5 {
		t.Fatalf("got %d chunks, want 4, the company with the summary, and one of boilerplate", len(cs))
	}
	tests := []struct {
		use, placed, text string
	}{
		{Embed, `nothing before it: summary by default; the heading "About Brightway": it starts "about", the employer's own words, with the summary`,
			"We deliver parcels across the north of England every day.\nAbout Brightway\nWe are a family firm."},
		{Embed, `the heading "What you'll be doing": "what you'll be doing" in headings.txt`, "What you'll be doing\nDrive the van."},
		{Metadata, `the heading "Job details": "job details" in headings.txt`, "Job details\nSalary: £30,000"},
		{LeaveOut, `the line "How to apply": exactly a section's name`, "How to apply\nApply today."},
	}
	for i, tt := range tests {
		c := cs[i]
		if c.Use != tt.use || c.PlacedBy != tt.placed || c.Text != tt.text {
			t.Errorf("chunk %d: got %s, %q, %q; want %s, %q, %q", i, c.Use, c.PlacedBy, c.Text, tt.use, tt.placed, tt.text)
		}
	}
	// Boilerplate leaves an embedded chunk for one of its own, at the end,
	// each line with the section it came from.
	b := cs[4]
	if b.Kind != Boilerplate || b.Use != LeaveOut || b.Text != "responsibilities: We are an agency." {
		t.Errorf("the boilerplate chunk: %+v", b)
	}
	if got := cs[2].Fields; len(got) != 1 || got[0].Label != "salary" || got[0].Value != "£30,000" {
		t.Errorf("the details' fields: %+v", got)
	}
	// Tokens: about four for every three words.
	if got := Tokens("We deliver parcels across the north of England every day."); got != 13 {
		t.Errorf("tokens: got %d, want 13 (10 words)", got)
	}
}

// An unmapped section the guess can't place goes to the boilerplate
// chunk, each line with its heading and why it's left out, and counts as
// unclassified; its fields stay, as details.
func TestOfUnmapped(t *testing.T) {
	ss := []sections.Section{{Kind: sections.Unmapped, Heading: "Our depots", Parts: []headings.Part{
		text("Leeds and York."),
		{Role: headings.Field, Text: "Pay: £16.50 per hour", Key: "pay", Value: "£16.50 per hour", Added: "Pay: "},
	}}}
	cs := Of(ss, func(string) bool { return false })
	if len(cs) != 2 {
		t.Fatalf("got %d chunks, want details and boilerplate: %+v", len(cs), cs)
	}
	d, b := cs[0], cs[1]
	if d.Kind != sections.Details || d.Use != Metadata || len(d.Fields) != 1 || d.Fields[0].Value != "£16.50 per hour" ||
		!strings.Contains(d.PlacedBy, "unmapped") {
		t.Errorf("the fields: %+v; want a details chunk of the unmapped section's fields", d)
	}
	if b.Kind != Boilerplate || b.Text != `unmapped "Our depots" (the advertiser's): Leeds and York.` || b.Unmapped != Tokens("Leeds and York.") {
		t.Errorf("the boilerplate: %+v; want the unmapped text, with its heading, counted", b)
	}
	if a := Totalled(Ad{Chunks: cs}, 20); a.Unclassified != b.Unmapped || a.Embedded != 0 {
		t.Errorf("totals: %+v", a)
	}
}

// YAML writes an ad's chunks as YAML that reads back as written, a
// chunk's text a block of its lines, whatever the text holds.
func TestYAML(t *testing.T) {
	tricky := []string{
		"📢 Calling all teachers 📢\nnext line", "✏️ Pencils", "a: b", "# not a comment", "yes", "42",
		"it's \"quoted\"", " leading space\nsecond", "control\x02character", "trailing ", "-", "",
		"Pay: £130 - £170 per day", "tab\tin it",
	}
	a := Ad{ID: "225682916", Title: "Part-Time Teacher: Supply", Tokens: 525, Embedded: 143, Unclassified: 60}
	for i, s := range tricky {
		a.Chunks = append(a.Chunks, Chunk{
			Kind: sections.Summary, Heading: s, PlacedBy: s, Use: Embed, Text: s, Tokens: i,
			Fields: []Field{{Label: "pay", Value: s}},
			Facts:  []facts.Fact{{Kind: "pay", Text: s, Values: map[string]string{"min": "50000", "period": "year", "note": s}}},
		})
	}
	var b strings.Builder
	if err := a.WriteYAML(&b); err != nil {
		t.Fatal(err)
	}
	var back Ad
	if err := yaml.Unmarshal([]byte(b.String()), &back); err != nil {
		t.Fatalf("not YAML: %v\n%s", err, b.String())
	}
	if !reflect.DeepEqual(a, back) {
		t.Errorf("read back differs.\nwrote:\n%s\ngot %+v", b.String(), back)
	}
	for _, want := range []string{
		"id: '225682916'\n", "tokens: 525\nembedded_tokens: 143\nunclassified_tokens: 60\nchunks:\n",
		"    text: |-\n      📢 Calling all teachers 📢\n      next line\n", "    use: embed\n",
	} {
		if !strings.Contains(b.String(), want) {
			t.Errorf("the YAML lacks %q; it was:\n%s", want, b.String())
		}
	}
}

// An ad's totals: all its tokens, those embedded, and those unclassified.
func TestTotals(t *testing.T) {
	cs := []Chunk{{Use: Embed, Tokens: 10}, {Kind: Boilerplate, Use: LeaveOut, Tokens: 9, Unmapped: 4}, {Use: LeaveOut, Tokens: 6}, {Use: Embed, Tokens: 1}}
	a := Totalled(Ad{Chunks: cs}, 30)
	if a.Tokens != 30 || a.Embedded != 11 || a.Unclassified != 4 {
		t.Errorf("got %d, %d, %d; want 30, 11, 4", a.Tokens, a.Embedded, a.Unclassified)
	}
}

// Boilerplate leaves embedded chunks for a chunk of its own; other
// chunks keep theirs, as part of a list such as an advertiser's perks.
func TestBoilerplateGathered(t *testing.T) {
	ss := []sections.Section{
		{Kind: sections.Benefits, Heading: "Perks", Parts: []headings.Part{text("A virtual GP."), text("A pension plan.")}},
		{Kind: sections.Requirements, Heading: "You'll need", Parts: []headings.Part{text("A licence."), text("A pension plan.")}},
		{Kind: sections.Summary, Parts: []headings.Part{text("We hire drivers."), {Role: headings.Text, Text: "A pension plan.", Item: true}}},
	}
	cs := Of(ss, func(s string) bool { return s == "A pension plan." })
	if len(cs) != 4 {
		t.Fatalf("got %d chunks, want 3 and one of boilerplate", len(cs))
	}
	if got := kind(cs, sections.Benefits).Text; got != "Perks\nA virtual GP.\nA pension plan." {
		t.Errorf("benefits: %q; want every line kept", got)
	}
	if got := kind(cs, sections.Requirements).Text; got != "You'll need\nA licence." {
		t.Errorf("requirements: %q; want the boilerplate gone", got)
	}
	if got := kind(cs, Boilerplate).Text; got != "requirements: A pension plan.\nsummary: A pension plan." {
		t.Errorf("boilerplate: %q; want each line with where it was", got)
	}
	// No boilerplate, no chunk of it.
	if got := Of(ss[:1], func(string) bool { return false }); len(got) != 1 {
		t.Errorf("got %d chunks, want 1", len(got))
	}
}

// List items are shown as entries of a list.
func TestListItems(t *testing.T) {
	ss := []sections.Section{{Kind: sections.Requirements, Heading: "You'll need", Parts: []headings.Part{
		text("Some of:"),
		{Role: headings.Text, Text: "A licence.", Item: true},
		{Role: headings.Field, Text: "Hours: 40", Key: "hours", Value: "40", Item: true},
	}}}
	c := Of(ss, func(string) bool { return false })[0]
	if c.Text != "You'll need\nSome of:\n- A licence.\n- Hours: 40" {
		t.Errorf("got %q", c.Text)
	}
	// A list's dashes aren't words, so aren't counted as tokens.
	if want := Tokens("You'll need Some of: A licence. Hours: 40"); c.Tokens != want {
		t.Errorf("tokens: got %d, want %d", c.Tokens, want)
	}
}

// The role, what the job is, is embedded, as the summary is.
func TestRoleEmbedded(t *testing.T) {
	c := Of([]sections.Section{{Kind: sections.Role, Heading: "About the role", Parts: []headings.Part{text("You will drive.")}}},
		func(string) bool { return false })[0]
	if c.Use != Embed {
		t.Errorf("got %s, want embed", c.Use)
	}
}

// Summary, role, responsibilities and requirements come first, in that
// order; the rest after, in the ad's order; the boilerplate last.
func TestOrder(t *testing.T) {
	ss := []sections.Section{
		{Kind: sections.Requirements, Heading: "You'll need", Parts: []headings.Part{text("A licence.")}},
		{Kind: sections.Benefits, Heading: "Perks", Parts: []headings.Part{text("A pension.")}},
		{Kind: sections.Role, Heading: "The role", Parts: []headings.Part{text("Driving."), text("We are an agency.")}},
		{Kind: sections.Company, Heading: "About us", Parts: []headings.Part{text("A firm.")}},
		{Kind: sections.Summary, Parts: []headings.Part{text("Hiring now.")}},
		{Kind: sections.Requirements, Heading: "Also", Parts: []headings.Part{text("A van.")}},
	}
	var got []string
	for _, c := range Of(ss, func(s string) bool { return s == "We are an agency." }) {
		got = append(got, string(c.Kind)+" "+c.Heading)
	}
	// The two requirements sections are one chunk, its headings lines in it.
	// The company joins the summary.
	want := []string{"summary ", "role The role", "requirements ", "benefits Perks", "boilerplate "}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

// kind returns the first chunk of kind k.
func kind(cs []Chunk, k sections.Kind) Chunk {
	for _, c := range cs {
		if c.Kind == k {
			return c
		}
	}
	return Chunk{}
}

// Lines of facts leave the chunks they're in for a details chunk of
// facts, each with its values; a details chunk keeps its lines, and
// gains their facts.
func TestFacts(t *testing.T) {
	ss := []sections.Section{
		{Kind: sections.Summary, Parts: []headings.Part{
			text("Join our team as a Project Manager."),
			{Role: headings.Text, Text: "40hrs pw", Item: true},
			{Role: headings.Text, Text: "50,000pa + excellent bonus and benefits", Item: true},
			text("This is a permanent role working with an estates team across the north."),
		}},
		{Kind: sections.Details, Heading: "Job details", Parts: []headings.Part{
			{Role: headings.Field, Text: "Pay: £16.50 per hour", Key: "pay", Value: "£16.50 per hour"},
		}},
	}
	cs := Of(ss, func(string) bool { return false })
	sum := kind(cs, sections.Summary)
	if sum.Text != "Join our team as a Project Manager.\nThis is a permanent role working with an estates team across the north." {
		t.Errorf("summary: %q; want its fact lines gone, the sentence with a fact in it kept", sum.Text)
	}
	// The ad's own details and the lines of facts taken from the summary
	// are one details chunk, its facts listed once: the sentence kept in
	// the summary has its fact listed too.
	d := kind(cs, sections.Details)
	if d.Use != Metadata || d.Text != "Job details\nPay: £16.50 per hour\n- 40hrs pw\n- 50,000pa + excellent bonus and benefits" ||
		len(d.Facts) != 4 || d.Facts[0].Values["period"] != "hour" || d.Facts[1].Kind != "hours" ||
		d.Facts[2].Values["min"] != "50000" || d.Facts[3].Values["value"] != "permanent" || !strings.Contains(d.PlacedBy, "facts") {
		t.Errorf("the details: %+v", d)
	}
}

// Sentences, and lines with a work pattern in them among other words,
// aren't lines of facts.
func TestFactLine(t *testing.T) {
	for line, want := range map[string]bool{
		"40hrs pw":                                               true,
		"Bromsgrove | Full-time | Permanent":                     true,
		"- Hybrid working":                                       true,
		"Mon - Fri 8:30am till 5:30pm":                           true,
		"Monday - Friday, 9:00am - 5:00pm.":                      true,
		"Contract administration and commercial risk management": false,
		"You'll work 40 hours per week on a shift pattern, including sleep-ins.":   false,
		"is hiring temporary School Cleaning Staff for schools in Chester. If you": false,
		"Canteen on site": false,
	} {
		if got := factLine(line, facts.Find(line)); got != want {
			t.Errorf("factLine(%q) = %t, want %t", line, got, want)
		}
	}
}

// Lines of a summary, and a role's list items, that say by their opening
// what they are go to a chunk of that kind; a role's prose stays.
func TestWording(t *testing.T) {
	ss := []sections.Section{
		{Kind: sections.Summary, Parts: []headings.Part{
			text("Join our team as a Project Manager."),
			{Role: headings.Text, Text: "Driving License essential", Item: true},
			{Role: headings.Text, Text: "Free parking", Item: true},
		}},
		{Kind: sections.Role, Heading: "About the Role", Parts: []headings.Part{
			text("You'll take ownership of a diverse portfolio of projects."),
			text("Lead the estates team in delivering projects."),
			{Role: headings.Text, Text: "Conduct site surveys.", Item: true},
			{Role: headings.Text, Text: "Manage tender exercises.", Item: true},
		}},
	}
	cs := Of(ss, func(s string) bool { return s == "Free parking" })
	if got := kind(cs, sections.Summary).Text; got != "Join our team as a Project Manager." {
		t.Errorf("summary: %q", got)
	}
	if got := kind(cs, sections.Role).Text; got != "About the Role\nYou'll take ownership of a diverse portfolio of projects.\nLead the estates team in delivering projects." {
		t.Errorf("role: %q; want its prose kept", got)
	}
	req, resp, ben := kind(cs, sections.Requirements), kind(cs, sections.Responsibilities), kind(cs, sections.Benefits)
	if req.Text != "- Driving License essential" || !strings.Contains(req.PlacedBy, `"Driving License"`) || req.Use != Embed {
		t.Errorf("requirements: %+v", req)
	}
	if resp.Text != "- Conduct site surveys.\n- Manage tender exercises." || !strings.Contains(resp.PlacedBy, "wording") {
		t.Errorf("responsibilities: %+v", resp)
	}
	// Placed by its wording before the boilerplate check: a benefit, not
	// boilerplate.
	if ben.Text != "- Free parking" || kind(cs, Boilerplate).Text != "" {
		t.Errorf("benefits: %+v; boilerplate %q", ben, kind(cs, Boilerplate).Text)
	}
}

// Faults found in the wash-up: a work pattern with a word or two after
// it is no line of facts, unless the line is laid out as a list of them.
func TestFactLineWashUp(t *testing.T) {
	for line, want := range map[string]bool{
		"Contract administration": false,
		"Remote team":             false,
		"Temporary staff":         false,
		"Leeds | Permanent":       true,
		"Permanent":               true,
	} {
		if got := factLine(line, facts.Find(line)); got != want {
			t.Errorf("factLine(%q) = %t, want %t", line, got, want)
		}
	}
	// A block of text that starts with a line break keeps its first line.
	if got := block("\n x\ny", "      "); !strings.Contains(got, "x") {
		t.Errorf("block lost text: %q", got)
	}
}

// The guess places an unmapped section by its lines' wording, or says
// why it's left out.
func TestGuess(t *testing.T) {
	never := func(string) bool { return false }
	always := func(string) bool { return true }
	duties := sections.Section{Kind: sections.Unmapped, Heading: "Your Background:", Parts: []headings.Part{
		text("Experience of working in a care setting"), text("You'll have a full UK licence"), text("Team player"),
	}}
	tests := []struct {
		name        string
		ss          []sections.Section
		boilerplate func(string) bool
		kind        sections.Kind
		why         string
	}{
		{"by wording", []sections.Section{duties}, never, sections.Requirements,
			`guessed requirements: 2 of its 3 lines open as requirements do ("Experience of", "You'll have")`},
		{"mostly boilerplate", []sections.Section{duties}, always, "", "mostly boilerplate"},
		{"the advertiser's", []sections.Section{{Kind: sections.Unmapped, Heading: "We're the AA",
			Parts: []headings.Part{text("A household name.")}}}, never, "", "the advertiser's"},
		{"after applying", []sections.Section{{Kind: sections.Apply, Heading: "How to apply"},
			{Kind: sections.Unmapped, Heading: "Agency:", Parts: []headings.Part{text("A note.")}}}, never, "", "after applying"},
		{"otherwise the summary", []sections.Section{{Kind: sections.Unmapped, Heading: "Junior Estimator",
			Parts: []headings.Part{text("Are you looking to build a career in estimating?")}}}, never,
			sections.Summary, "guessed summary: nothing says it isn't about the job"},
		{"a label and its value", []sections.Section{{Kind: sections.Unmapped, Heading: "IR35 Status:",
			Parts: []headings.Part{text("Inside IR35")}}}, never, sections.Details, "guessed details: a label, with a short value under it"},
		{"who are", []sections.Section{{Kind: sections.Unmapped, Heading: "Who are Acme?",
			Parts: []headings.Part{text("A household name.")}}}, never, "", "the advertiser's"},
		{"no text", []sections.Section{{Kind: sections.Unmapped, Heading: "Terms",
			Parts: []headings.Part{{Role: headings.Field, Text: "Pay: £16.50", Key: "pay", Value: "£16.50"}}}}, never, "", "no text"},
	}
	for _, tt := range tests {
		kind, why := guess(tt.ss, len(tt.ss)-1, tt.boilerplate)
		if kind != tt.kind || why != tt.why {
			t.Errorf("%s: got %q, %q; want %q, %q", tt.name, kind, why, tt.kind, tt.why)
		}
	}
	// A guessed section is a chunk of its kind, embedded, saying why.
	cs := Of([]sections.Section{duties}, never)
	if len(cs) != 1 || cs[0].Kind != sections.Requirements || cs[0].Use != Embed || !strings.HasPrefix(cs[0].PlacedBy, "guessed") {
		t.Errorf("the chunks: %+v; want one guessed requirements chunk", cs)
	}
}

// A field stays in its section's chunk, and is listed in the details too,
// as metadata, wherever it's written; its text isn't repeated there.
func TestFieldsInDetailsToo(t *testing.T) {
	contract := headings.Part{Role: headings.Field, Text: "Contract type: Permanent", Key: "contract type", Value: "Permanent"}
	pay := headings.Part{Role: headings.Field, Text: "Pay rate: £13.00 per hour", Key: "pay rate", Value: "£13.00 per hour"}
	ss := []sections.Section{
		{Kind: sections.Summary, Parts: []headings.Part{text("We deliver parcels across the north of England."), contract}},
		{Kind: sections.Benefits, Heading: "Benefits", Parts: []headings.Part{pay, text("A pension.")}},
	}
	cs := Of(ss, func(string) bool { return false })
	var details *Chunk
	for i := range cs {
		if cs[i].Kind == sections.Details {
			details = &cs[i]
		}
	}
	if details == nil {
		t.Fatalf("no details chunk in %+v", cs)
	}
	want := []Field{{"contract type", "Permanent"}, {"pay rate", "£13.00 per hour"}}
	if !reflect.DeepEqual(details.Fields, want) {
		t.Errorf("the details' fields: %+v, want %+v", details.Fields, want)
	}
	if details.Text != "" || details.Tokens != 0 {
		t.Errorf("the details repeat the fields' text: %q, %d tokens", details.Text, details.Tokens)
	}
	if !strings.Contains(cs[0].Text, "Contract type: Permanent") || len(cs[0].Fields) != 1 {
		t.Errorf("the summary lost its field: %+v", cs[0])
	}
	if !strings.Contains(cs[1].Text, "Pay rate: £13.00 per hour") || len(cs[1].Fields) != 1 {
		t.Errorf("the benefits lost their field: %+v", cs[1])
	}
}

// A fact stated in a section's prose stays there, and is listed in the
// details too: a working pattern written into the role is metadata.
func TestFactsInProseInDetailsToo(t *testing.T) {
	line := "You'll work Monday to Friday, 8am to 4pm, from our depot."
	ss := []sections.Section{
		{Kind: sections.Role, Heading: "The Role", Parts: []headings.Part{text("You'll deliver parcels to homes."), text(line)}},
	}
	cs := Of(ss, func(string) bool { return false })
	if !strings.Contains(cs[0].Text, line) {
		t.Errorf("the role lost its line: %q", cs[0].Text)
	}
	var details *Chunk
	for i := range cs {
		if cs[i].Kind == sections.Details {
			details = &cs[i]
		}
	}
	if details == nil || len(details.Facts) == 0 {
		t.Fatalf("no facts listed in the details: %+v", cs)
	}
	if details.Text != "" {
		t.Errorf("the details repeat the role's line: %q", details.Text)
	}
}

// Salary written in the benefits stays there, and is listed in the details.
func TestFactsLeftOutInDetailsToo(t *testing.T) {
	ss := []sections.Section{
		{Kind: sections.Benefits, Heading: "Benefits", Parts: []headings.Part{text("A salary of £30,000 a year, and a pension.")}},
	}
	cs := Of(ss, func(string) bool { return false })
	for _, c := range cs {
		if c.Kind == sections.Details {
			for _, f := range c.Facts {
				if f.Kind == "pay" {
					return
				}
			}
		}
	}
	t.Errorf("no pay fact listed in the details: %+v", cs)
}

// Facts in what the employer says of itself, or of applying, aren't the
// job's, and aren't listed in the details.
func TestFactsAboutTheEmployerNotInDetails(t *testing.T) {
	ss := []sections.Section{
		{Kind: sections.Company, Heading: "About us", Parts: []headings.Part{text("We place staff in temporary, contract and permanent roles.")}},
		{Kind: sections.Apply, Heading: "How to apply", Parts: []headings.Part{text("If you have not heard from us within 48 hours, please call.")}},
	}
	for _, c := range Of(ss, func(string) bool { return false }) {
		if c.Kind == sections.Details {
			t.Errorf("facts about the employer listed in the details: %+v", c.Facts)
		}
	}
}

// AdOf makes an ad's chunks with its title, the ad's title field, as a
// chunk of its own, first, embedded, and counted in the ad's tokens.
func TestAdOfTitle(t *testing.T) {
	ss := []sections.Section{
		{Kind: sections.Summary, Parts: []headings.Part{
			{Role: headings.Field, Text: "Job Title: Production Operative", Key: "job title", Value: "Production Operative"},
			text("We make parts for cars."),
		}},
	}
	a := AdOf("7", "Production Operative", ss, func(string) bool { return false })
	if len(a.Chunks) == 0 {
		t.Fatal("no chunks")
	}
	c := a.Chunks[0]
	if c.Kind != Title || c.Use != Embed || c.Text != "Production Operative" || c.Tokens != Tokens("Production Operative") {
		t.Errorf("the first chunk: %+v; want the title, embedded", c)
	}
	if a.Tokens != Whole(ss)+c.Tokens || a.Embedded != c.Tokens+Tokens("We make parts for cars.") {
		t.Errorf("totals: %d tokens, %d embedded; want the title counted in both", a.Tokens, a.Embedded)
	}
	// The title the ad writes as a field leaves the summary, as the title
	// chunk embeds it, but is listed in the details.
	if strings.Contains(a.Chunks[1].Text, "Job Title") {
		t.Errorf("the summary embeds the title again: %q", a.Chunks[1].Text)
	}
	var listed bool
	for _, c := range a.Chunks {
		if c.Kind == sections.Details && slices.Contains(c.Fields, Field{"job title", "Production Operative"}) {
			listed = true
		}
	}
	if !listed {
		t.Errorf("the job title field isn't listed in the details: %+v", a.Chunks)
	}
	if b := AdOf("8", "", ss, func(string) bool { return false }); b.Chunks[0].Kind == Title {
		t.Errorf("an ad with no title has a title chunk: %+v", b.Chunks[0])
	}
}

// A fact stated more than once is listed in the details once: "Full Time"
// and "Full-time" are one, and one the ad's own details list already isn't
// listed again.
func TestFactsListedOnce(t *testing.T) {
	ss := []sections.Section{
		{Kind: sections.Summary, Parts: []headings.Part{
			text("A Full Time role, Monday to Friday."),
			text("This is a full-time job in our depot."),
		}},
		{Kind: sections.Details, Heading: "Job details", Parts: []headings.Part{text("Monday to Friday")}},
	}
	seen := map[string]int{}
	for _, c := range Of(ss, func(string) bool { return false }) {
		if c.Kind != sections.Details {
			continue
		}
		for _, f := range c.Facts {
			seen[f.Kind+fmt.Sprint(f.Values)]++
		}
	}
	for k, n := range seen {
		if n > 1 {
			t.Errorf("%s listed %d times in the details", k, n)
		}
	}
	if len(seen) != 2 {
		t.Errorf("got %d distinct facts, want 2 (full time, Monday to Friday): %v", len(seen), seen)
	}
}

// An ad's requirements, desirable ones too, are one chunk: each section's
// heading a line over its list, "Desirable" where it has none.
func TestRequirementsOneChunk(t *testing.T) {
	item := func(s string) headings.Part { return headings.Part{Role: headings.Text, Text: s, Item: true} }
	ss := []sections.Section{
		{Kind: sections.Summary, Parts: []headings.Part{text("We make parts for cars."), text("Experience of CNC machines.")}},
		{Kind: sections.Requirements, Heading: "Essential", Parts: []headings.Part{item("A forklift licence"), item("Two years in a warehouse")}},
		{Kind: sections.Benefits, Heading: "Benefits", Parts: []headings.Part{text("A pension.")}},
		{Kind: sections.Desirable, Parts: []headings.Part{item("A first aid certificate")}},
	}
	var reqs []Chunk
	for _, c := range Of(ss, func(string) bool { return false }) {
		if c.Kind == sections.Requirements || c.Kind == sections.Desirable {
			reqs = append(reqs, c)
		}
	}
	if len(reqs) != 1 {
		t.Fatalf("got %d requirements chunks, want 1: %+v", len(reqs), reqs)
	}
	r := reqs[0]
	want := "Essential\n- A forklift licence\n- Two years in a warehouse\nExperience of CNC machines.\nDesirable\n- A first aid certificate"
	if r.Kind != sections.Requirements || r.Use != Embed || r.Text != want {
		t.Errorf("the requirements: %s, %s,\n%s\nwant:\n%s", r.Kind, r.Use, r.Text, want)
	}
	if r.Tokens != Tokens(want) || !strings.Contains(r.PlacedBy, "Essential") || !strings.Contains(r.PlacedBy, "wording") {
		t.Errorf("the requirements' tokens %d, placed by %q", r.Tokens, r.PlacedBy)
	}
}

// An ad's chunks of one kind are one chunk, wherever its sections are:
// two summaries are one, and the ad's own details and the details gathered
// from its other sections are one, each fact in them listed once.
func TestOneChunkAKind(t *testing.T) {
	ss := []sections.Section{
		{Kind: sections.Summary, Parts: []headings.Part{text("We make parts for cars.")}},
		{Kind: sections.Benefits, Heading: "Benefits", Parts: []headings.Part{text("A pension.")}},
		{Kind: sections.Summary, Heading: "The company", Parts: []headings.Part{text("We're forty years old."), text("We are a full time team in Leeds, Monday to Friday.")}},
		{Kind: sections.Details, Heading: "Job details", Parts: []headings.Part{text("Monday to Friday, 8am to 4pm")}},
	}
	kinds := map[sections.Kind]int{}
	var summary, details Chunk
	for _, c := range Of(ss, func(string) bool { return false }) {
		kinds[c.Kind]++
		switch c.Kind {
		case sections.Summary:
			summary = c
		case sections.Details:
			details = c
		}
	}
	for k, n := range kinds {
		if n > 1 {
			t.Errorf("%d %s chunks, want one", n, k)
		}
	}
	if want := "We make parts for cars.\nThe company\nWe're forty years old.\nWe are a full time team in Leeds, Monday to Friday."; summary.Text != want {
		t.Errorf("the summary:\n%s\nwant:\n%s", summary.Text, want)
	}
	seen := map[string]int{}
	for _, f := range details.Facts {
		seen[f.Kind+fmt.Sprint(f.Values)]++
	}
	for k, n := range seen {
		if n > 1 {
			t.Errorf("%s listed %d times in the details", k, n)
		}
	}
}

// The employer's own words go with the summary, embedded, under their
// heading; its standard text, repeated whatever the job, is boilerplate.
func TestCompanyWithTheSummary(t *testing.T) {
	ss := []sections.Section{
		{Kind: sections.Summary, Parts: []headings.Part{text("We're hiring a driver in Leeds.")}},
		{Kind: sections.Responsibilities, Heading: "Duties", Parts: []headings.Part{text("Drive the van.")}},
		{Kind: sections.Company, Heading: "About us", Parts: []headings.Part{
			text("We're a family firm of forty years."), text("Acme is an equal opportunities employer."),
			text("We place staff in temporary, contract and permanent roles."),
		}},
	}
	cs := Of(ss, func(s string) bool { return s == "Acme is an equal opportunities employer." })
	for _, c := range cs {
		if c.Kind == sections.Company {
			t.Errorf("a company chunk: %+v", c)
		}
		if c.Kind == sections.Details {
			t.Errorf("the company's facts listed in the details: %+v", c.Facts)
		}
	}
	s := kind(cs, sections.Summary)
	want := "We're hiring a driver in Leeds.\nAbout us\nWe're a family firm of forty years.\nWe place staff in temporary, contract and permanent roles."
	if s.Use != Embed || s.Text != want {
		t.Errorf("the summary: %s\n%s\nwant:\n%s", s.Use, s.Text, want)
	}
	if b := kind(cs, Boilerplate); !strings.Contains(b.Text, "Acme is an equal opportunities employer.") {
		t.Errorf("the company's standard text isn't boilerplate: %q", b.Text)
	}
}

// The company's lines of facts are the employer's, not the job's: they
// stay in its text, with the summary, and aren't taken to the details.
func TestCompanyFactsStay(t *testing.T) {
	ss := []sections.Section{
		{Kind: sections.Summary, Parts: []headings.Part{text("We're hiring a driver.")}},
		{Kind: sections.Company, Heading: "About us", Parts: []headings.Part{
			text("Temporary, contract and permanent roles"), text("Head office: Leeds LS1 4AB"),
		}},
	}
	cs := Of(ss, func(string) bool { return false })
	for _, c := range cs {
		if c.Kind == sections.Details {
			t.Errorf("the company's lines taken to the details: %q, %+v", c.Text, c.Facts)
		}
	}
	if s := kind(cs, sections.Summary); !strings.Contains(s.Text, "Head office: Leeds LS1 4AB") {
		t.Errorf("the company's line left its text: %q", s.Text)
	}
}

// Lines placed as requirements by their wording are requirements, not
// desirable: they come before the "Desirable" sub-list.
func TestDesirableLast(t *testing.T) {
	item := func(s string) headings.Part { return headings.Part{Role: headings.Text, Text: s, Item: true} }
	ss := []sections.Section{
		{Kind: sections.Summary, Parts: []headings.Part{text("We're hiring."), text("Experience of HGV driving essential")}},
		{Kind: sections.Requirements, Heading: "Essential", Parts: []headings.Part{item("Patience")}},
		{Kind: sections.Desirable, Heading: "Desirable", Parts: []headings.Part{item("A forklift ticket")}},
	}
	r := kind(Of(ss, func(string) bool { return false }), sections.Requirements)
	want := "Essential\n- Patience\nExperience of HGV driving essential\nDesirable\n- A forklift ticket"
	if r.Text != want {
		t.Errorf("the requirements:\n%s\nwant:\n%s", r.Text, want)
	}
}
