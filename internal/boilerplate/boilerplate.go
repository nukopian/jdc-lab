// Package boilerplate finds the text job ads repeat whatever the job: an
// agency's description of itself, its standard benefits, its legal notes.
//
// The ads say nothing of who placed them, so boilerplate is found by
// repetition: a block of text, word for word (by its key, see
// headings.Key), in ads for different jobs. "Different jobs" is the point.
// An advertiser posting one job in ten towns repeats every block of it,
// and none of that is boilerplate. So what counts is how many groups of
// similar titles the text appears under, not how many ads: "Support
// Worker" and "Residential Support Worker" are one group, "Crane Operator"
// and "Forklift Driver" two.
//
// Repetition isn't enough on its own. Related jobs from one advertiser
// share lines that describe the job, such as "Excellent ICT skills", and
// those are worth keeping. So a repeated text found mostly in the sections
// that describe the job (responsibilities, requirements, desirable) is a
// template for the job, and only one found mostly elsewhere is
// boilerplate. The verdict is the text's, over all the ads it's in, so a
// heading filed under the wrong kind in one ad doesn't change it.
//
// Footers, such as an agency's equal opportunities notice, often come
// after an ad's last heading, which is usually its requirements, and so
// are filed under it. So where a text stands counts too: one found
// mostly in the run of repeated texts that ends its ads, and talks as a
// footer does (of applying, the agency, "we" and "our"), is boilerplate,
// whatever its section, unless it's phrased as a job's requirement or
// duty is (see internal/wording): an advertiser's list of requirements
// can run to the end of its ads too. Text under Widespread
// groups of titles or more is boilerplate wherever it's found, unless
// it's phrased so and found mostly in the sections that describe the
// job: "Able to work as part of a team" is about the job, however many
// ads say it.
//
// Standard text isn't always word for word the same. Agencies write each
// ad's own title and place into it ("we're recruiting a Chef in Leeds"),
// so a text is also keyed with those as placeholders. And long texts are
// often changed in a word or two, a school's or an agency's name, so
// texts of LongWords words or more that share most of their runs of five
// words are taken as one text. Short lines are matched only as written:
// phrased alike, they're more often about the job than standard text.
package boilerplate

import (
	"hash/fnv"
	"slices"
	"strings"

	"github.com/nukopian/jdc-lab/internal/ads"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
	"github.com/nukopian/jdc-lab/internal/wording"
)

// MinWords is the fewest words a repeated text has to count: shorter ones,
// such as "Pension", are common because they're short, not because
// they're boilerplate.
const MinWords = 8

// MinGroups is the fewest groups of similar titles a text must appear
// under to count as repeated.
const MinGroups = 2

// Widespread is the number of groups of similar titles a text must appear
// under to be boilerplate wherever it's found.
const Widespread = 5

// LongWords is the fewest words a text needs to be matched to texts that
// differ from it in a few words.
const LongWords = 20

// Overlap is how much of their runs of five words two long texts must
// share to be one text: the share of all their runs that both have.
const Overlap = 0.6

// commonRun is how many texts a run of five words can be in and still
// be used to find texts like one another: one in more is too common to
// say they're alike, and would make finding them slow.
const commonRun = 300

// A Source is the ad a text is from: its title, and its places, which
// standard text often has written into it.
type Source struct {
	Title  string
	Places []string
}

// SourceOf returns an ad's Source: its title, and each part of each of its
// locations, as "Leeds, West Yorkshire" names two places.
func SourceOf(a ads.Ad) Source {
	s := Source{Title: a.Title}
	for _, l := range a.Locations() {
		for _, p := range strings.Split(l, ",") {
			if p = strings.TrimSpace(p); p != "" {
				s.Places = append(s.Places, p)
			}
		}
	}
	return s
}

// The placeholders for an ad's own title and places in a key. Keys are
// made by headings.Key, which never writes braces.
const (
	titleMark = "{title}"
	placeMark = "{place}"
)

// template returns key with the source's title and places, where they
// stand as whole words, replaced by placeholders: places first, the
// longest first, as a title such as "Driver - Leeds" holds one.
func (s Source) template(key string) string {
	var places []string
	for _, p := range s.Places {
		if k := headings.Key(p); k != "" {
			places = append(places, k)
		}
	}
	slices.SortFunc(places, func(a, b string) int { return len(b) - len(a) })
	k := " " + key + " "
	for _, p := range places {
		k = strings.ReplaceAll(k, " "+p+" ", " "+placeMark+" ")
	}
	if t := headings.Key(s.Title); t != "" {
		k = strings.ReplaceAll(k, " "+t+" ", " "+titleMark+" ")
	}
	return strings.TrimSpace(k)
}

// An Index knows, for each text, the titles of the ads it's in, and the
// kinds of section it's in there. Add every ad first, then ask about any
// text, with the ad it's from. Its zero value is ready to use.
type Index struct {
	texts  map[string]*text // key → where the text is
	long   []string         // keys of the long texts, templated, in the order added
	isLong map[string]bool
	// alike maps a long text's templated key to its group of texts alike,
	// and groups counts each key's groups of titles; both are made when
	// first asked for, and forgotten when a text is added.
	alike  map[string]string
	groups map[string]int
	// endsAdded is true when Ends has counted texts since the groups of
	// texts alike were made, so they're to be made again.
	endsAdded bool
}

// text is what's known of a text: the distinct title keys of its ads,
// its times in sections that describe the job, and in others, and its
// times in the run of repeated texts that ends an ad.
type text struct {
	titles               []string
	describes, elsewhere int
	atEnd                int
}

// Add records texts from a section of kind, in an ad from src.
func (ix *Index) Add(src Source, kind sections.Kind, texts ...string) {
	if ix.texts == nil {
		ix.texts, ix.isLong = map[string]*text{}, map[string]bool{}
	}
	t := headings.Key(src.Title)
	for _, s := range texts {
		k := headings.Key(s)
		if len(strings.Fields(k)) < MinWords {
			continue
		}
		tk := src.template(k)
		ix.record(k, t, kind)
		if tk != k {
			ix.record(tk, t, kind)
		}
		if len(strings.Fields(tk)) >= LongWords && !ix.isLong[tk] {
			ix.isLong[tk] = true
			ix.long = append(ix.long, tk)
		}
	}
	ix.alike, ix.groups = nil, nil // what was worked out before is out of date
}

// record adds one ad's use of a text, by its key.
func (ix *Index) record(key, title string, kind sections.Kind) {
	x := ix.texts[key]
	if x == nil {
		x = &text{}
		ix.texts[key] = x
	}
	if !slices.Contains(x.titles, title) {
		x.titles = append(x.titles, title)
	}
	if Describes(kind) {
		x.describes++
	} else {
		x.elsewhere++
	}
}

// Ends records which of an ad's texts, all of them in order, are in the
// run of repeated texts that ends it: from the last, back to the first
// that isn't repeated, passing over those too short to count. Call it for
// each ad, once every ad has been added.
func (ix *Index) Ends(src Source, texts ...string) {
	for i := len(texts) - 1; i >= 0; i-- {
		k := headings.Key(texts[i])
		if len(strings.Fields(k)) < MinWords {
			continue
		}
		if !ix.Repeated(src, texts[i]) {
			break
		}
		keys := []string{k}
		if tk := src.template(k); tk != k {
			keys = append(keys, tk)
		}
		for _, key := range keys {
			if x := ix.texts[key]; x != nil {
				x.atEnd++
			}
		}
	}
	ix.endsAdded = true
}

// keys returns the keys a text from src is known by: its own, its
// templated one, and its group of texts alike, those it has.
func (ix *Index) keys(src Source, s string) []string {
	k := headings.Key(s)
	ks := []string{k}
	tk := src.template(k)
	if tk != k {
		ks = append(ks, tk)
	}
	if ix.isLong[tk] {
		ix.findAlike()
		ks = append(ks, ix.alike[tk])
	}
	return ks
}

// Groups returns how many groups of similar titles a text from src
// appears under, by the key that finds the most: 0 for a text shorter
// than MinWords, or never added.
func (ix *Index) Groups(src Source, s string) int {
	n := 0
	for _, k := range ix.keys(src, s) {
		n = max(n, ix.groupsOf(k))
	}
	return n
}

// groupsOf counts a key's groups of similar titles, once.
func (ix *Index) groupsOf(key string) int {
	if n, ok := ix.groups[key]; ok {
		return n
	}
	n := 0
	if x := ix.texts[key]; x != nil {
		n = groups(x.titles)
	}
	if ix.groups == nil {
		ix.groups = map[string]int{}
	}
	ix.groups[key] = n
	return n
}

// Repeated reports whether a text from src appears under at least
// MinGroups groups of similar titles.
func (ix *Index) Repeated(src Source, s string) bool {
	return ix.Groups(src, s) >= MinGroups
}

// Boilerplate reports whether a text from src is boilerplate, by any of
// its keys: repeated and found more often outside the sections that
// describe the job than in them; or, unless it's phrased as a job's
// requirement or duty, repeated, found mostly in the runs of repeated
// text that end its ads (see Ends), and talking as a footer does (see
// FooterLine); or under Widespread groups of titles
// or more, unless it's phrased so and found mostly in the sections that
// describe the job.
func (ix *Index) Boilerplate(src Source, s string) bool {
	if ix.endsAdded { // the groups of texts alike are to count the ends too
		ix.alike, ix.groups, ix.endsAdded = nil, nil, false
	}
	job := JobLine(s)
	for _, k := range ix.keys(src, s) {
		n := ix.groupsOf(k)
		if n < MinGroups {
			continue
		}
		x := ix.texts[k]
		switch {
		case x.elsewhere > x.describes,
			!job && 2*x.atEnd > x.describes+x.elsewhere && FooterLine(s),
			n >= Widespread && !(job && x.describes >= x.elsewhere):
			return true
		}
	}
	return false
}

// footerWords are words that footers use and job lines don't: about
// applying and getting in touch, the advertiser and its notices.
var footerWords = map[string]bool{
	"apply": true, "applying": true, "application": true, "applications": true, "applicants": true,
	"contact": true, "call": true, "cv": true, "email": true, "touch": true, "interested": true, "please": true,
	"agency": true, "business": true, "recruitment": true, "consultancy": true, "consultant": true,
	"vacancy": true, "vacancies": true, "employer": true, "opportunities": true,
	"we": true, "we're": true, "our": true, "us": true,
	"privacy": true, "safeguarding": true, "ir35": true, "umbrella": true,
}

// FooterLine reports whether s talks as a footer does, with one of
// footerWords: "Interested? Call Jamie at Acme", "Acme is acting as an
// employment agency". An advertiser's list of requirements, which can run
// to the end of its ads as footers do, seldom has any.
func FooterLine(s string) bool {
	return slices.ContainsFunc(strings.Fields(headings.Key(s)), func(w string) bool { return footerWords[w] })
}

// JobLine reports whether s is phrased as a job's requirement or duty
// is, by its opening (see internal/wording): "Ability to work as part of
// a team", "Liaise with site teams".
func JobLine(s string) bool {
	k, _, ok := wording.Of(s)
	return ok && (k == sections.Requirements || k == sections.Responsibilities)
}

// Template reports whether a text from src is a template for the job:
// repeated, but not widespread, and found at least as often in the
// sections that describe the job as outside them. A tie goes this way, as
// keeping text about the job is the safer mistake.
func (ix *Index) Template(src Source, s string) bool {
	return ix.Repeated(src, s) && !ix.Boilerplate(src, s)
}

// findAlike groups the long texts that share at least Overlap of their
// runs of five words, a group taking in any text alike to one of its
// own, and records each group as a text of its own, under the key "~"
// and its first text's key, with all its texts' titles and kinds.
func (ix *Index) findAlike() {
	if ix.alike != nil {
		return
	}
	for k := range ix.texts { // groups made before, worked out again below
		if strings.HasPrefix(k, "~") {
			delete(ix.texts, k)
		}
	}
	runs := make([]map[uint64]bool, len(ix.long))
	in := map[uint64][]int{} // a run of words → the texts it's in
	for i, k := range ix.long {
		runs[i] = shingles(k)
		for r := range runs[i] {
			in[r] = append(in[r], i)
		}
	}
	group := make([]int, len(ix.long)) // union-find: each text's parent
	for i := range group {
		group[i] = i
	}
	root := func(i int) int {
		for group[i] != i {
			group[i] = group[group[i]]
			i = group[i]
		}
		return i
	}
	for i := range ix.long {
		tried := map[int]bool{}
		for r := range runs[i] {
			if len(in[r]) > commonRun {
				continue
			}
			for _, j := range in[r] {
				if j <= i || tried[j] {
					continue
				}
				tried[j] = true
				if alike(runs[i], runs[j]) {
					group[root(j)] = root(i)
				}
			}
		}
	}
	ix.alike = map[string]string{}
	for i, k := range ix.long {
		g := "~" + ix.long[root(i)]
		ix.alike[k] = g
		x, y := ix.texts[g], ix.texts[k]
		if x == nil {
			x = &text{}
			ix.texts[g] = x
		}
		for _, t := range y.titles {
			if !slices.Contains(x.titles, t) {
				x.titles = append(x.titles, t)
			}
		}
		x.describes += y.describes
		x.elsewhere += y.elsewhere
		x.atEnd += y.atEnd
	}
}

// shingles returns a key's runs of five words, each as a hash of its
// words: far smaller than the words themselves, and two runs' hashes are
// the same only if their words are, as near as matters.
func shingles(key string) map[uint64]bool {
	w := strings.Fields(key)
	m := map[uint64]bool{}
	h := fnv.New64a()
	for i := 0; i+5 <= len(w); i++ {
		h.Reset()
		h.Write([]byte(strings.Join(w[i:i+5], " ")))
		m[h.Sum64()] = true
	}
	return m
}

// alike reports whether two texts' runs of words overlap by Overlap.
func alike(a, b map[uint64]bool) bool {
	if len(a) > len(b) {
		a, b = b, a
	}
	shared := 0
	for r := range a {
		if b[r] {
			shared++
		}
	}
	all := len(a) + len(b) - shared
	return all > 0 && float64(shared) >= Overlap*float64(all)
}

// Describes reports whether a section of kind describes the job itself,
// so that text repeated in it is a template for the job, not boilerplate.
func Describes(kind sections.Kind) bool {
	switch kind {
	case sections.Responsibilities, sections.Requirements, sections.Desirable:
		return true
	}
	return false
}

// groups counts the groups of similar titles among titles: each title
// joins the first group with a title similar to it.
func groups(titles []string) int {
	var gs [][]string
	for _, t := range titles {
		i := slices.IndexFunc(gs, func(g []string) bool {
			return slices.ContainsFunc(g, func(u string) bool { return similar(t, u) })
		})
		if i < 0 {
			gs = append(gs, []string{t})
		} else {
			gs[i] = append(gs[i], t)
		}
	}
	return len(gs)
}

// SimilarTitles reports whether two titles are of the same job, as
// boilerplate groups them: see similar.
func SimilarTitles(a, b string) bool {
	return similar(headings.Key(a), headings.Key(b))
}

// similar reports whether two title keys are titles of the same job: the
// words of one are all in the other, or they share at least half their
// words between them. A title with no words is like no other.
func similar(a, b string) bool {
	if a == b {
		return true
	}
	if a == "" || b == "" {
		return false
	}
	wa, wb := words(a), words(b)
	shared := 0
	for w := range wa {
		if wb[w] {
			shared++
		}
	}
	all := len(wa) + len(wb) - shared
	return shared == len(wa) || shared == len(wb) || 2*shared >= all
}

func words(s string) map[string]bool {
	m := map[string]bool{}
	for _, w := range strings.Fields(s) {
		m[w] = true
	}
	return m
}
