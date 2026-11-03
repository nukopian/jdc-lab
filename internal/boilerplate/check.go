package boilerplate

import (
	"cmp"
	"slices"

	"github.com/nukopian/jdc-lab/internal/facts"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
)

// A Checked is a text found as boilerplate, with what's needed to check
// the verdict: where it's found, the rule that decided (see Why), and
// three flags for where something important is likeliest.
type Checked struct {
	Text string
	// Title is the title of the first ad it's in.
	Title string
	// Ads and Chars are the ads it's in and its characters over all of
	// them; Groups its groups of similar titles.
	Ads, Chars, Groups int
	// Kinds are the kinds of section it's in, in sections.Kinds' order,
	// Describing its characters in the sections that describe the job,
	// and Embedded in the sections whose text is embedded, those and the
	// summary and company's: what the verdict leaves out of search, as the
	// other sections are left out whatever their text.
	Kinds                []sections.Kind
	Describing, Embedded int
	// Why is the rule that made it boilerplate, in the first ad it's in.
	Why string
	// Facts are the kinds of fact it holds, each once, in the order found.
	Facts []string
	// Job says it's phrased as a job's requirement or duty (see JobLine).
	Job bool
}

// Flagged reports whether something important is likelier in the text:
// it holds a fact, it's phrased as a job line, or most of its characters
// are in the sections that describe the job.
func (c Checked) Flagged() bool {
	return len(c.Facts) > 0 || c.Job || c.MostlyDescribing()
}

// MostlyDescribing reports whether most of a text's characters are in
// the sections that describe the job.
func (c Checked) MostlyDescribing() bool {
	return 2*c.Describing > c.Chars
}

// Check returns every distinct text of the ads found as boilerplate, by
// its key, the most characters first.
func Check(all []Divided, ix *Index) []Checked {
	byKey := map[string]*Checked{}
	kinds := map[string]map[sections.Kind]bool{}
	for _, a := range all {
		src := SourceOf(a.Ad)
		seen := map[string]bool{}
		for _, s := range a.Sections {
			for _, p := range s.Parts {
				if p.Role != headings.Text {
					continue
				}
				why := ix.Why(src, p.Text)
				if why == "" {
					continue
				}
				k := headings.Key(p.Text)
				c := byKey[k]
				if c == nil {
					c = &Checked{Text: p.Text, Title: a.Ad.Title, Groups: ix.Groups(src, p.Text), Why: why, Job: JobLine(p.Text)}
					for _, f := range facts.Find(p.Text) {
						if !slices.Contains(c.Facts, f.Kind) {
							c.Facts = append(c.Facts, f.Kind)
						}
					}
					byKey[k] = c
					kinds[k] = map[sections.Kind]bool{}
				}
				if !seen[k] {
					c.Ads++
					seen[k] = true
				}
				c.Chars += len(p.Text)
				if Describes(s.Kind) {
					c.Describing += len(p.Text)
				}
				if Describes(s.Kind) || s.Kind == sections.Summary || s.Kind == sections.Company {
					c.Embedded += len(p.Text)
				}
				kinds[k][s.Kind] = true
			}
		}
	}
	out := make([]Checked, 0, len(byKey))
	for k, c := range byKey {
		for _, kind := range sections.Kinds {
			if kinds[k][kind] {
				c.Kinds = append(c.Kinds, kind)
			}
		}
		out = append(out, *c)
	}
	slices.SortFunc(out, func(a, b Checked) int {
		return cmp.Or(cmp.Compare(b.Chars, a.Chars), cmp.Compare(a.Text, b.Text))
	})
	return out
}
