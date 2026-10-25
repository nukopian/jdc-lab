// Command sections divides every job ad into sections (see internal/blocks,
// internal/headings and internal/sections), and measures how well the
// heading phrases in internal/sections/headings.txt cover the ads. Ads
// whose HTML is the same as an earlier one's are counted once.
//
// It writes, into OUTDIR:
//
//   - ads.jsonl: each ad's sections, one ad per line;
//   - unmapped.tsv: headings of no known kind, by how many ads use
//     them, most first, with the characters of text under them: the list
//     to work down when adding phrases;
//   - disagree.tsv: headings whose phrases point to more than one kind:
//     the kind that won, and all of them, in the order they come in the
//     heading;
//   - named.tsv: lines of text, with no cue to be headings, that started a
//     section by being exactly a section's name.
//
// And it prints a summary.
//
// From the lab's folder:
//
//	go run ./cmd/sections data/ingestion/ads.jsonl data/sections
//
// or, once installed with `go install ./cmd/...`, just sections.
package main

import (
	"bufio"
	"cmp"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/nukopian/jdc-lab/internal/ads"
	"github.com/nukopian/jdc-lab/internal/blocks"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
)

// record is one line of ads.jsonl.
type record struct {
	ID       string             `json:"id"`
	Sections []sections.Section `json:"sections"`
}

// heading is what's known of one heading key, over all the ads.
type heading struct {
	ads   map[string]bool // the ads using it
	text  string          // how it's first written
	kinds []sections.Kind // what its phrases point to
	chars int             // characters of text in the sections it starts
}

// tally is what's counted over all the ads.
type tally struct {
	ads, duplicates int
	uses, mapped    int            // headings, and those of a known kind
	named           map[string]int // lines of text that start a section by name alone
	allMapped       int            // ads whose every heading is of a known kind
	onlySummary     int            // ads with no section but the summary
	kindAds         map[sections.Kind]int
	kindChars       map[sections.Kind]int // characters of text in each kind
	chars           int
	headings        map[string]*heading
}

func main() {
	log.SetFlags(0)
	if len(os.Args) != 3 {
		log.Fatal("usage: sections ADS.jsonl OUTDIR")
	}
	source, dir := os.Args[1], os.Args[2]
	t, err := run(source, dir)
	if err != nil {
		log.Fatal(err)
	}
	if err := t.tables(dir); err != nil {
		log.Fatal(err)
	}
	t.print(dir)
}

func run(source, dir string) (t tally, err error) {
	t.kindAds, t.kindChars = map[sections.Kind]int{}, map[sections.Kind]int{}
	t.headings, t.named = map[string]*heading{}, map[string]int{}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return t, err
	}
	out, err := os.Create(filepath.Join(dir, "ads.jsonl"))
	if err != nil {
		return t, err
	}
	defer func() {
		if cerr := out.Close(); err == nil {
			err = cerr
		}
	}()
	buf := bufio.NewWriter(out)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)

	var d ads.Distinct
	for ad, err := range ads.File(source) {
		if err != nil {
			return t, err
		}
		if !d.First(ad) {
			continue
		}
		parts := headings.Read(blocks.Split(ad.HTML), ad.Title, ad.Locations()...)
		ss := sections.Divide(parts)
		t.add(ad.ID, parts, ss)
		if err := enc.Encode(record{ID: ad.ID, Sections: ss}); err != nil {
			return t, err
		}
	}
	t.duplicates = d.Duplicates
	return t, buf.Flush()
}

func (t *tally) add(id string, parts []headings.Part, ss []sections.Section) {
	t.ads++
	all := true
	for _, p := range parts {
		if p.Role != headings.Heading {
			continue
		}
		t.uses++
		if _, ok := sections.Of(p.Key); ok {
			t.mapped++
		} else {
			all = false
		}
		h := t.headings[p.Key]
		if h == nil {
			h = &heading{ads: map[string]bool{}, text: p.Text, kinds: sections.All(p.Key)}
			t.headings[p.Key] = h
		}
		h.ads[id] = true
	}
	if all {
		t.allMapped++
	}

	kinds := map[sections.Kind]bool{}
	for _, s := range ss {
		kinds[s.Kind] = true
		if s.ByName {
			t.named[s.Heading]++
		}
		for _, p := range s.Parts {
			if p.Role == headings.Text {
				t.kindChars[s.Kind] += len(p.Text)
				t.chars += len(p.Text)
				if h := t.headings[headings.Key(s.Heading)]; h != nil && s.Kind == sections.Unmapped {
					h.chars += len(p.Text)
				}
			}
		}
	}
	for k := range kinds {
		t.kindAds[k]++
	}
	if len(ss) <= 1 && (len(ss) == 0 || ss[0].Kind == sections.Summary) {
		t.onlySummary++
	}
}

// tables writes unmapped.tsv, disagree.tsv and named.tsv.
func (t tally) tables(dir string) error {
	keys := make([]string, 0, len(t.headings))
	for k := range t.headings {
		keys = append(keys, k)
	}
	slices.SortFunc(keys, func(a, b string) int {
		return cmp.Or(cmp.Compare(len(t.headings[b].ads), len(t.headings[a].ads)), cmp.Compare(a, b))
	})
	var unmapped, disagree strings.Builder
	unmapped.WriteString("ads\tchars\tkey\theading\n")
	disagree.WriteString("ads\twon\tkinds\theading\n")
	for _, k := range keys {
		h := t.headings[k]
		if _, ok := sections.Of(k); !ok {
			fmt.Fprintf(&unmapped, "%d\t%d\t%s\t%s\n", len(h.ads), h.chars, k, h.text)
		}
		if len(h.kinds) > 1 {
			names := make([]string, len(h.kinds))
			for i, kind := range h.kinds {
				names[i] = string(kind)
			}
			won, _ := sections.Of(k)
			fmt.Fprintf(&disagree, "%d\t%s\t%s\t%s\n", len(h.ads), won, strings.Join(names, ","), h.text)
		}
	}
	var named strings.Builder
	named.WriteString("lines\ttext\n")
	texts := make([]string, 0, len(t.named))
	for text := range t.named {
		texts = append(texts, text)
	}
	slices.SortFunc(texts, func(a, b string) int {
		return cmp.Or(cmp.Compare(t.named[b], t.named[a]), cmp.Compare(a, b))
	})
	for _, text := range texts {
		fmt.Fprintf(&named, "%d\t%s\n", t.named[text], text)
	}
	for name, b := range map[string]*strings.Builder{"unmapped.tsv": &unmapped, "disagree.tsv": &disagree, "named.tsv": &named} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(b.String()), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func (t tally) print(dir string) {
	fmt.Printf("%d ads (%d duplicates left out); written to %s\n\n", t.ads, t.duplicates, dir)
	fmt.Printf("headings of a known kind:        %d of %d uses (%.0f%%)\n", t.mapped, t.uses, pct(t.mapped, t.uses))
	named := 0
	for _, n := range t.named {
		named += n
	}
	fmt.Printf("lines starting one by name:      %d\n", named)
	fmt.Printf("ads with every heading known:    %d (%.0f%%)\n", t.allMapped, pct(t.allMapped, t.ads))
	fmt.Printf("ads with only a summary:         %d (%.0f%%)\n\n", t.onlySummary, pct(t.onlySummary, t.ads))
	fmt.Printf("%-17s %6s %8s\n", "section", "ads", "of text")
	for _, k := range sections.Kinds {
		fmt.Printf("%-17s %5.0f%% %7.0f%%\n", k, pct(t.kindAds[k], t.ads), pct(t.kindChars[k], t.chars))
	}
}

func pct(n, of int) float64 {
	if of == 0 {
		return 0
	}
	return 100 * float64(n) / float64(of)
}
