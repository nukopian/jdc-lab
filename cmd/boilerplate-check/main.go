// Command boilerplate-check lists every text of the ads found as
// boilerplate, to check the verdict (see internal/boilerplate's Check):
// with the rule that decided, and flags for where something important is
// likeliest, a fact, a job line, or the sections that describe the job.
//
// It writes OUTDIR/check.tsv, most characters first, and prints a
// summary: the texts and their characters by rule, and flagged.
//
// From the lab's folder:
//
//	go run ./cmd/boilerplate-check data/runs/latest/ads.jsonl data/runs/latest/boilerplate-check
//
// or, once installed with `go install ./cmd/...`, just boilerplate-check.
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/nukopian/jdc-lab/internal/boilerplate"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) != 3 {
		log.Fatal("usage: boilerplate-check ADS.jsonl OUTDIR")
	}
	source, dir := os.Args[1], os.Args[2]
	all, ix, _, err := boilerplate.Read(source)
	if err != nil {
		log.Fatal(err)
	}
	checked := boilerplate.Check(all, &ix)
	if err := write(filepath.Join(dir, "check.tsv"), checked); err != nil {
		log.Fatal(err)
	}
	summarise(checked, len(all), dir)
}

func write(path string, cs []boilerplate.Checked) (err error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}()
	w := bufio.NewWriter(f)
	fmt.Fprintln(w, "chars\tads\tgroups\tdescribing\tembedded\tkinds\twhy\tfacts\tjob\tflagged\ttitle\ttext")
	for _, c := range cs {
		var kinds []string
		for _, k := range c.Kinds {
			kinds = append(kinds, string(k))
		}
		fmt.Fprintf(w, "%d\t%d\t%d\t%d\t%d\t%s\t%s\t%s\t%t\t%t\t%s\t%s\n", c.Chars, c.Ads, c.Groups, c.Describing, c.Embedded,
			strings.Join(kinds, ","), c.Why, strings.Join(c.Facts, ","), c.Job, c.Flagged(), c.Title, c.Text)
	}
	return w.Flush()
}

// summarise prints the texts and their characters: in all, by rule, and
// by flag.
func summarise(cs []boilerplate.Checked, ads int, dir string) {
	type tally struct{ texts, chars int }
	total := tally{}
	byRule := map[string]*tally{}
	flags := map[string]*tally{}
	add := func(m map[string]*tally, k string, c boilerplate.Checked) {
		if m[k] == nil {
			m[k] = &tally{}
		}
		m[k].texts++
		m[k].chars += c.Chars
	}
	for _, c := range cs {
		total.texts++
		total.chars += c.Chars
		rule, _, _ := strings.Cut(c.Why, ",")
		add(byRule, rule, c)
		if len(c.Facts) > 0 {
			add(flags, "holds a fact", c)
		}
		if c.Job {
			add(flags, "a job line", c)
		}
		if c.MostlyDescribing() {
			add(flags, "mostly describing the job", c)
		}
		if c.Flagged() {
			add(flags, "any of them", c)
		}
	}
	pct := func(n int) float64 { return 100 * float64(n) / float64(max(total.chars, 1)) }
	fmt.Printf("%d ads; %d texts found as boilerplate, %d characters in all; written to %s\n\n",
		ads, total.texts, total.chars, dir)
	print := func(title string, m map[string]*tally) {
		fmt.Printf("%-28s %7s %12s\n", title, "texts", "of chars")
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		slices.SortFunc(keys, func(a, b string) int { return m[b].chars - m[a].chars })
		for _, k := range keys {
			fmt.Printf("%-28s %7d %11.1f%%\n", k, m[k].texts, pct(m[k].chars))
		}
		fmt.Println()
	}
	print("by rule", byRule)
	print("flagged", flags)
}
