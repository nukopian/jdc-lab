package boilerplate

import (
	"slices"
	"strings"
	"testing"

	"github.com/nukopian/jdc-lab/internal/ads"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
)

const (
	agency   = "A dedicated consultant will be by your side at every step of the process."
	duties   = "Prepare learning materials and assist with engaging classroom activities every day."
	location = "This is a fantastic opportunity to join a busy team based in central Leeds."
)

func TestSimilar(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"support worker", "residential support worker", true},                 // one in the other
		{"paramedic or nurse part time", "paramedic or nurse full time", true}, // most words shared
		{"crane operator", "forklift driver", false},
		{"teaching assistant", "cover supervisor", false},
		{"sen teacher", "teacher", true},
		{"", "teacher", false}, // a title with no words is like no other
		{"", "", true},
	}
	for _, tt := range tests {
		if got := similar(tt.a, tt.b); got != tt.want {
			t.Errorf("similar(%q, %q) = %t, want %t", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestIndex(t *testing.T) {
	var ix Index // the zero value is ready to use
	ix.Add(Source{Title: "Crane Operator"}, sections.Company, agency, location, "Pension")
	ix.Add(Source{Title: "Forklift Driver"}, sections.Company, agency, "Pension")
	ix.Add(Source{Title: "Teaching Assistant"}, sections.Responsibilities, duties, location)
	ix.Add(Source{Title: "SEN Teaching Assistant"}, sections.Responsibilities, duties, location)
	ix.Add(Source{Title: "Teaching Assistant"}, sections.Summary, location) // the same title again adds nothing

	tests := []struct {
		text   string
		groups int
	}{
		{agency, 2},    // two different jobs
		{duties, 1},    // two titles, but one job
		{location, 2},  // three titles, two jobs
		{"Pension", 0}, // too short to count
		{"Never seen in any ad at all, not even once, anywhere.", 0},
	}
	for _, tt := range tests {
		if got := ix.Groups(Source{}, tt.text); got != tt.groups {
			t.Errorf("Groups(%q) = %d, want %d", tt.text, got, tt.groups)
		}
	}
	if !ix.Repeated(Source{}, agency) || ix.Repeated(Source{}, duties) {
		t.Errorf("Repeated: agency %t, duties %t; want true, false", ix.Repeated(Source{}, agency), ix.Repeated(Source{}, duties))
	}
}

func TestIndexCountsAgainAfterAdd(t *testing.T) {
	var ix Index
	ix.Add(Source{Title: "Crane Operator"}, sections.Company, agency)
	if ix.Repeated(Source{}, agency) {
		t.Fatal("repeated after one ad")
	}
	ix.Add(Source{Title: "Forklift Driver"}, sections.Company, agency)
	if !ix.Repeated(Source{}, agency) {
		t.Error("not repeated after a second job's ad")
	}
}

func TestBoilerplate(t *testing.T) {
	// Each text is in three ads for different jobs, in these kinds of
	// section; the verdict is the majority's, a tie going to template.
	tests := []struct {
		kinds       []sections.Kind
		boilerplate bool
	}{
		{[]sections.Kind{sections.Company, sections.Benefits, sections.Unmapped}, true},
		{[]sections.Kind{sections.Company, sections.Company, sections.Requirements}, true}, // misfiled once
		{[]sections.Kind{sections.Responsibilities, sections.Requirements, sections.Summary}, false},
		{[]sections.Kind{sections.Responsibilities, sections.Responsibilities, sections.Desirable}, false},
	}
	jobs := []string{"Crane Operator", "Teaching Assistant", "Chef"}
	for _, tt := range tests {
		var ix Index
		for i, kind := range tt.kinds {
			ix.Add(Source{Title: jobs[i]}, kind, agency)
		}
		if got := ix.Boilerplate(Source{}, agency); got != tt.boilerplate {
			t.Errorf("in %v: Boilerplate = %t, want %t", tt.kinds, got, tt.boilerplate)
		}
		if got := ix.Template(Source{}, agency); got == tt.boilerplate {
			t.Errorf("in %v: Template = %t, want %t", tt.kinds, got, !tt.boilerplate)
		}
	}
	var ix Index
	ix.Add(Source{Title: "Chef"}, sections.Company, agency)
	if ix.Boilerplate(Source{}, agency) || ix.Template(Source{}, agency) {
		t.Error("a text in one ad is neither")
	}
}

// Text in Widespread groups of titles is boilerplate wherever it's
// found, unless it's phrased as a job's requirement or duty and found
// mostly in the sections that describe the job.
func TestWidespreadIsBoilerplate(t *testing.T) {
	const (
		notice   = "All candidates will be required to complete our standard vetting checks before starting."
		teamwork = "Able to work effectively as part of a team and on your own initiative."
	)
	var ix Index
	jobs := []string{"Chef", "Crane Operator", "Teaching Assistant", "Nurse", "Accountant"}
	for i, job := range jobs {
		ix.Add(Source{Title: job}, sections.Requirements, notice, teamwork)
		if got, want := ix.Boilerplate(Source{}, notice), i+1 >= Widespread; got != want {
			t.Errorf("a notice in %d jobs' requirements: Boilerplate = %t, want %t", i+1, got, want)
		}
		if ix.Boilerplate(Source{}, teamwork) {
			t.Errorf("a job line in %d jobs' requirements: Boilerplate = true, want false", i+1)
		}
	}
	// Found mostly outside the sections about the job, it's boilerplate.
	for _, job := range jobs {
		ix.Add(Source{Title: job}, sections.Benefits, teamwork)
		ix.Add(Source{Title: job}, sections.Company, teamwork)
	}
	if !ix.Boilerplate(Source{}, teamwork) {
		t.Error("a job line mostly outside the job's sections: Boilerplate = false, want true")
	}
}

// A repeated text found mostly in the run of repeated texts that ends its
// ads is boilerplate, whatever its section, unless it's phrased as a job
// line; one that's followed by text of the ad's own is a template still.
func TestFootersAtTheEnd(t *testing.T) {
	const (
		footer = "Acme Recruitment is acting as an employment agency in relation to this vacancy."
		// A footer word ("our") in a job line doesn't make it a footer.
		job    = "Experience of working with our clients in a busy warehouse with a forklift licence."
		own    = "This line is the advertiser's own, written for this one ad and no other."
		middle = "Our client is a well established firm with a fine reputation in the region."
	)
	var ix Index
	ads := []struct {
		src   Source
		texts []string
	}{
		{Source{Title: "Chef"}, []string{own + " Chef", middle, own + " Chef again", job, footer}},
		{Source{Title: "Crane Operator"}, []string{own + " Crane", middle, own + " Crane again", job, footer}},
	}
	for _, a := range ads {
		ix.Add(a.src, sections.Requirements, a.texts...)
	}
	for _, a := range ads {
		ix.Ends(a.src, a.texts...)
	}
	src := ads[0].src
	if !ix.Boilerplate(src, footer) {
		t.Error("a footer at the end of its ads, under requirements: Boilerplate = false, want true")
	}
	if ix.Boilerplate(src, job) {
		t.Error("a job line at the end of its ads: Boilerplate = true, want false")
	}
	if ix.Boilerplate(src, middle) || !ix.Template(src, middle) {
		t.Error("a repeated text followed by the ad's own: want a template, not boilerplate")
	}
}

// Standard text with each ad's own title and place written in is one
// text: "We're recruiting a Chef in Leeds" and "a Nurse in York".
func TestTitleAndPlaceWrittenIn(t *testing.T) {
	jobs := []Source{
		{Title: "Chef", Places: []string{"Leeds"}},
		{Title: "Crane Operator", Places: []string{"York"}},
		{Title: "Teaching Assistant", Places: []string{"Hull", "East Yorkshire"}},
		{Title: "Nurse", Places: []string{"Bath"}},
		{Title: "Accountant", Places: []string{"Derby"}},
	}
	text := func(s Source) string {
		return "We're recruiting a " + s.Title + " in " + s.Places[0] + " to join our friendly and growing team today."
	}
	var ix Index
	for _, s := range jobs {
		ix.Add(s, sections.Summary, text(s))
	}
	for _, s := range jobs {
		if got := ix.Groups(s, text(s)); got != len(jobs) {
			t.Errorf("%s: Groups = %d, want %d", s.Title, got, len(jobs))
		}
		if !ix.Boilerplate(s, text(s)) {
			t.Errorf("%s: not boilerplate", s.Title)
		}
	}
	// Without its ad, a text is only itself.
	if got := ix.Groups(Source{}, text(jobs[0])); got != 1 {
		t.Errorf("without its source: Groups = %d, want 1", got)
	}
}

// A text that names one ad's title still matches the same text in ads
// whose title it doesn't name: its own key counts as well as the
// templated one.
func TestPlainKeyStillCounts(t *testing.T) {
	const notice = "Teaching assistants and support staff are welcome to apply through our agency today."
	var ix Index
	for _, title := range []string{"Teaching Assistants", "Chef", "Nurse"} {
		ix.Add(Source{Title: title}, sections.Apply, notice)
	}
	if got := ix.Groups(Source{Title: "Teaching Assistants"}, notice); got != 3 {
		t.Errorf("Groups = %d, want 3", got)
	}
}

// Long texts with a few words changed are one text; short ones aren't,
// as short lines phrased alike are often about the job.
func TestNearDuplicates(t *testing.T) {
	long := "Our agency is committed to safeguarding and promoting the welfare of children and young people, " +
		"and all workers must complete an enhanced DBS check and provide two references before starting any placement."
	if n := len(strings.Fields(long)); n < LongWords {
		t.Fatalf("the long text has %d words, under LongWords", n)
	}
	short := "Assist with goods in and goods out activities each and every day."
	variants := []struct{ long, short string }{
		{long, short},
		{strings.Replace(long, "Our agency", "Protocol Education", 1), strings.Replace(short, "Assist", "Assisting", 1)},
		{strings.Replace(long, "two references", "three references", 1), strings.Replace(short, "activities", "processes", 1)},
	}
	jobs := []string{"Chef", "Crane Operator", "Teaching Assistant"}
	var ix Index
	for i, v := range variants {
		ix.Add(Source{Title: jobs[i]}, sections.Legal, v.long, v.short)
	}
	for i, v := range variants {
		src := Source{Title: jobs[i]}
		if got := ix.Groups(src, v.long); got != 3 {
			t.Errorf("long variant %d: Groups = %d, want 3", i, got)
		}
		if got := ix.Groups(src, v.short); got != 1 {
			t.Errorf("short variant %d: Groups = %d, want 1", i, got)
		}
	}
	// Texts sharing a little are still two.
	other := "Our agency is committed to finding every candidate a role they love, with weekly pay, " +
		"holiday pay, a dedicated consultant and free training courses throughout the whole year ahead."
	ix.Add(Source{Title: "Nurse"}, sections.Company, other)
	if got := ix.Groups(Source{Title: "Nurse"}, other); got != 1 {
		t.Errorf("a different text: Groups = %d, want 1", got)
	}
}

func TestSourceOf(t *testing.T) {
	loc, canon := "Leeds, West Yorkshire", "Leeds"
	got := SourceOf(ads.Ad{Title: "Chef", Location: &loc, LocationCanonical: &canon})
	if got.Title != "Chef" || strings.Join(got.Places, "|") != "Leeds|West Yorkshire|Leeds" {
		t.Errorf("got %+v", got)
	}
}

// Groups of texts alike are worked out again after an Add, not added to.
func TestNearDuplicatesAfterAdd(t *testing.T) {
	long := "Our agency is committed to safeguarding and promoting the welfare of children and young people, " +
		"and all workers must complete an enhanced DBS check and provide two references before starting any placement."
	other := strings.Replace(long, "two references", "three references", 1)
	var ix Index
	ix.Add(Source{Title: "Chef"}, sections.Legal, long)
	ix.Add(Source{Title: "Crane Operator"}, sections.Legal, other)
	if got := ix.Groups(Source{Title: "Chef"}, long); got != 2 {
		t.Fatalf("Groups = %d, want 2", got)
	}
	ix.Add(Source{Title: "Chef"}, sections.Company, "A short line that isn't alike to anything else here.")
	ix.Add(Source{Title: "Nurse"}, sections.Company, long)
	if got := ix.Groups(Source{Title: "Chef"}, long); got != 3 {
		t.Errorf("after more ads: Groups = %d, want 3", got)
	}
	// Two in legal and one in company: elsewhere all three times, so
	// boilerplate, not counted twice over.
	if g := ix.texts[ix.alike[headings.Key(long)]]; g.elsewhere != 3 {
		t.Errorf("the group's texts are in %d sections, want 3", g.elsewhere)
	}
}

// SimilarTitles compares titles as written, by their keys.
func TestSimilarTitles(t *testing.T) {
	if !SimilarTitles("SEN Teacher", "Teacher:") || SimilarTitles("Crane Operator", "Forklift Driver") {
		t.Error("SimilarTitles doesn't compare titles by their keys")
	}
}

// Long footers changed in a word or two from ad to ad are one text, and
// it's the group's ending runs that count: each one alone is in one ad.
func TestFootersAlikeAtTheEnd(t *testing.T) {
	footer := func(name string) string {
		return name + " Recruitment is acting as an employment agency in relation to this vacancy and will contact" +
			" shortlisted applicants within five working days of the closing date for applications to the role"
	}
	const own = "This line is the advertiser's own, written for this one ad and no other."
	var ix Index
	srcs := []Source{{Title: "Chef"}, {Title: "Crane Operator"}, {Title: "Nurse"}}
	names := []string{"Acme", "Apex", "Atlas"}
	for i, src := range srcs {
		ix.Add(src, sections.Requirements, own+" "+src.Title, footer(names[i]))
	}
	if ix.Boilerplate(srcs[0], footer(names[0])) {
		t.Fatal("before Ends: a footer filed under requirements is a template still")
	}
	for i, src := range srcs {
		ix.Ends(src, own+" "+src.Title, footer(names[i]))
	}
	if !ix.Boilerplate(srcs[0], footer(names[0])) {
		t.Error("a long footer, alike in its ads' ending runs: Boilerplate = false, want true")
	}
}

func TestWallJudgedWhole(t *testing.T) {
	short := "Acme is a disability-confident employer."
	long := "We welcome applications from everyone in every community across the region, whatever their background."
	footer := short + " " + long
	duty := "You'll plan the week's rota with the team leader."
	duties := duty + " " + long
	titles := []string{"Delivery Driver", "Head Chef", "Staff Nurse"}
	var ix Index
	for _, title := range titles {
		src := Source{Title: title}
		ix.Add(src, sections.Legal, short, long)
		ix.Wall(src, sections.Legal, footer, short, long)
	}
	src := Source{Title: "Delivery Driver"}
	for _, s := range []string{short, long} {
		if !ix.Boilerplate(src, s) {
			t.Errorf("%q, in a notices wall repeated under 3 groups of titles: not boilerplate", s)
		}
	}

	var jobs Index
	for _, title := range titles {
		src := Source{Title: title}
		jobs.Add(src, sections.Responsibilities, duty, long)
		jobs.Wall(src, sections.Responsibilities, duties, duty, long)
	}
	if jobs.Boilerplate(src, duty) {
		t.Errorf("%q, in an employer's duties repeated whole: boilerplate, want a template", duty)
	}
}

// A sentence that can be judged on its own, repeated enough to count,
// keeps its own verdict, whatever its wall's: an agency's opening line
// about the job, a template in the role, stays one inside a wall judged
// boilerplate.
func TestWallSentenceJudgedAlone(t *testing.T) {
	opening := "English teacher required for a high-achieving school in Central London."
	notice := "We are an agency acting for the school in all of its recruitment matters today."
	wall := opening + " " + notice
	titles := []string{"English Teacher", "Science Technician", "Head of Music"}
	var ix Index
	for _, title := range titles {
		src := Source{Title: title}
		ix.Add(src, sections.Role, opening)
		ix.Add(src, sections.Legal, notice)
		ix.Wall(src, sections.Legal, wall, opening, notice)
	}
	src := Source{Title: "English Teacher"}
	if ix.Boilerplate(src, opening) {
		t.Errorf("%q, a template for the job on its own: boilerplate through its wall", opening)
	}
	if !ix.Boilerplate(src, notice) {
		t.Errorf("%q: not boilerplate", notice)
	}
}

// Why names the rule that made a text boilerplate, and the key that found
// it, and nothing for a text that isn't.
func TestWhy(t *testing.T) {
	jobs := []string{"Chef", "Crane Operator", "Teaching Assistant", "Nurse", "Accountant"}

	var elsewhere Index
	for _, job := range jobs[:3] {
		elsewhere.Add(Source{Title: job}, sections.Company, agency)
		elsewhere.Add(Source{Title: job}, sections.Responsibilities, duties)
	}
	if got := elsewhere.Why(Source{}, agency); got != "elsewhere, by its own key" {
		t.Errorf("an agency's line: Why = %q", got)
	}
	if got := elsewhere.Why(Source{}, duties); got != "" {
		t.Errorf("duties, a template: Why = %q, want none", got)
	}

	const notice = "All candidates will be required to complete our standard vetting checks before starting."
	var widespread Index
	for _, job := range jobs {
		widespread.Add(Source{Title: job}, sections.Requirements, notice)
	}
	if got := widespread.Why(Source{}, notice); got != "widespread, by its own key" {
		t.Errorf("a notice in 5 jobs' requirements: Why = %q", got)
	}

	const footer = "Acme Recruitment is acting as an employment agency in relation to this vacancy."
	var footers Index
	for _, job := range jobs[:2] {
		footers.Add(Source{Title: job}, sections.Requirements, duties, footer)
	}
	for _, job := range jobs[:2] {
		footers.Ends(Source{Title: job}, duties, footer)
	}
	if got := footers.Why(Source{}, footer); got != "footer, by its own key" {
		t.Errorf("a footer under requirements: Why = %q", got)
	}

	var templated Index
	places := []string{"Leeds", "York", "Hull"}
	text := func(i int) string {
		return "Our client in " + places[i] + " is a well established firm with a fine reputation in the region."
	}
	for i, job := range jobs[:3] {
		templated.Add(Source{Title: job, Places: []string{places[i]}}, sections.Company, text(i))
	}
	if got := templated.Why(Source{Title: jobs[0], Places: places[:1]}, text(0)); got != "elsewhere, by its templated key" {
		t.Errorf("standard text with the ad's place written in: Why = %q", got)
	}

	long := "Our agency is committed to safeguarding and promoting the welfare of children and young people, " +
		"and all workers must complete an enhanced DBS check and provide two references before starting any placement."
	var alike Index
	for i, v := range []string{long, strings.Replace(long, "two", "three", 1), strings.Replace(long, "Our agency", "Acme", 1)} {
		alike.Add(Source{Title: jobs[i]}, sections.Legal, v)
	}
	if got := alike.Why(Source{Title: jobs[0]}, long); got != "elsewhere, by its alike key" {
		t.Errorf("a long notice changed a word in each ad: Why = %q", got)
	}

	short := "Acme is a disability-confident employer."
	rest := "We welcome applications from everyone in every community across the region, whatever their background."
	var walls Index
	for _, job := range jobs[:3] {
		walls.Wall(Source{Title: job}, sections.Legal, short+" "+rest, short, rest)
	}
	if got := walls.Why(Source{Title: jobs[0]}, short); got != "wall: elsewhere, by its own key" {
		t.Errorf("a short line of a notices wall: Why = %q", got)
	}
}

// Check lists each text found as boilerplate once, with where it's
// found, why, and its flags; a text that isn't boilerplate isn't there.
func TestCheck(t *testing.T) {
	const pay = "All our roles pay £12.50 per hour, paid weekly, with holiday pay on top of that rate."
	jobs := []string{"Chef", "Crane Operator", "Teaching Assistant"}
	var all []Divided
	for _, job := range jobs {
		all = append(all, Divided{
			Ad: ads.Ad{Title: job},
			Sections: []sections.Section{
				{Kind: sections.Company, Parts: []headings.Part{{Role: headings.Text, Text: agency}, {Role: headings.Text, Text: pay}}},
				{Kind: sections.Benefits, Parts: []headings.Part{{Role: headings.Text, Text: pay}}},
				{Kind: sections.Requirements, Parts: []headings.Part{{Role: headings.Text, Text: pay}}},
				{Kind: sections.Responsibilities, Parts: []headings.Part{{Role: headings.Text, Text: duties}}},
			},
		})
	}
	var ix Index
	for _, a := range all {
		src := SourceOf(a.Ad)
		for _, s := range a.Sections {
			ix.Add(src, s.Kind, Texts([]sections.Section{s})...)
		}
	}
	got := Check(all, &ix)
	if len(got) != 2 {
		t.Fatalf("got %d texts, want 2, the agency's line and its pay, not the duties: %+v", len(got), got)
	}
	// The pay line is in each ad three times, a third of it under
	// requirements.
	p := got[0]
	if p.Text != pay || p.Title != "Chef" || p.Ads != 3 || p.Chars != 9*len(pay) || p.Groups != 3 || p.Describing != 3*len(pay) || p.Embedded != 6*len(pay) {
		t.Errorf("pay: %+v", p)
	}
	if !slices.Equal(p.Kinds, []sections.Kind{sections.Requirements, sections.Benefits, sections.Company}) || p.Why != "elsewhere, by its own key" {
		t.Errorf("pay: kinds %v, why %q", p.Kinds, p.Why)
	}
	if !slices.Equal(p.Facts, []string{"pay"}) || !p.Flagged() || p.MostlyDescribing() {
		t.Errorf("pay: facts %v, flagged %t, mostly describing %t", p.Facts, p.Flagged(), p.MostlyDescribing())
	}
	if a := got[1]; a.Text != agency || a.Flagged() {
		t.Errorf("the agency's line: %+v, flagged %t", a, a.Flagged())
	}
}
