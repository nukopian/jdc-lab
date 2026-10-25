// Command heading-uses shows how the ads use one heading: how it's spelt,
// what kind of section it starts, what section comes before it, and the
// first lines under it, ad by ad. It's for deciding what an unmapped
// heading should be (see docs/PIPELINE.md, "Improving it").
//
// The heading is given by its key, as the tables in data/headings and
// data/sections list it, or as written: either way it's matched by key,
// so "Please Note:" finds "please note". From the lab's folder:
//
//	go run ./cmd/heading-uses data/ingestion/ads.jsonl "please note"
//
// or, once installed with `go install ./cmd/...`, just heading-uses.
package main

import (
	"cmp"
	"fmt"
	"log"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/nukopian/jdc-lab/internal/ads"
	"github.com/nukopian/jdc-lab/internal/blocks"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
)

// use is one ad's use of the heading.
type use struct {
	id, spelling string
	kind         sections.Kind
	before       string // the section before: its kind and heading
	lines        []string
}

func main() {
	log.SetFlags(0)
	if len(os.Args) != 3 {
		log.Fatal("usage: heading-uses ADS.jsonl HEADING")
	}
	source, key := os.Args[1], headings.Key(os.Args[2])
	uses, err := find(source, key)
	if err != nil {
		log.Fatal(err)
	}
	report(key, uses)
}

// find returns each ad's first use of a heading whose key is key. Ads whose
// HTML is the same as an earlier one's are left out.
func find(source, key string) ([]use, error) {
	var uses []use
	var d ads.Distinct
	for ad, err := range ads.File(source) {
		if err != nil {
			return nil, err
		}
		if !d.First(ad) {
			continue
		}
		ss := sections.Divide(headings.Read(blocks.Split(ad.HTML), ad.Title, ad.Locations()...))
		for i, s := range ss {
			if s.Heading == "" || headings.Key(s.Heading) != key {
				continue
			}
			u := use{id: ad.ID, spelling: s.Heading, kind: s.Kind, before: "(nothing)"}
			if i > 0 {
				u.before = fmt.Sprintf("%s %q", ss[i-1].Kind, ss[i-1].Heading)
			}
			for _, p := range s.Parts[:min(2, len(s.Parts))] {
				u.lines = append(u.lines, p.Text)
			}
			uses = append(uses, u)
			break
		}
	}
	return uses, nil
}

func report(key string, uses []use) {
	if len(uses) == 0 {
		fmt.Printf("%q: in no ad\n", key)
		return
	}
	kinds := map[sections.Kind]int{}
	for _, u := range uses {
		kinds[u.kind]++
	}
	fmt.Printf("%q: in %d ads, starting %s\n\n", key, len(uses), counts(kinds))

	spellings, befores := map[string]int{}, map[string]int{}
	for _, u := range uses {
		spellings[fmt.Sprintf("%q", u.spelling)]++
		befores[u.before]++
	}
	fmt.Println("spelt:     ", counts(spellings))
	fmt.Println("comes after:", counts(befores))
	fmt.Println()
	for _, u := range uses {
		fmt.Printf("%s  %q\n", u.id, u.spelling)
		for _, l := range u.lines {
			fmt.Printf("    %s\n", clip(l, 110))
		}
	}
}

// counts writes a tally, most first, as `3 a, 2 b`; the first five only.
func counts[K cmp.Ordered](m map[K]int) string {
	keys := slices.SortedFunc(maps.Keys(m), func(a, b K) int {
		return cmp.Or(cmp.Compare(m[b], m[a]), cmp.Compare(a, b))
	})
	var parts []string
	for _, k := range keys[:min(5, len(keys))] {
		parts = append(parts, fmt.Sprintf("%d %v", m[k], k))
	}
	if len(keys) > 5 {
		parts = append(parts, "…")
	}
	return strings.Join(parts, ", ")
}

// clip shortens s to at most n characters, marking where it was cut.
func clip(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}
