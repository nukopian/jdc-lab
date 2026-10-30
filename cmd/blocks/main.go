// Command blocks splits each job ad into blocks (see internal/blocks), writes
// them to OUTDIR/ads.jsonl, one ad per line, and prints a summary: how many
// blocks of each kind, and how many carry each cue.
//
// From the lab's folder:
//
//	go run ./cmd/blocks data/ingestion/ads-20261024.jsonl data/blocks
//
// or, once installed with `go install ./cmd/...`, just blocks.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/nukopian/jdc-lab/internal/ads"
	"github.com/nukopian/jdc-lab/internal/blocks"
)

// record is one line of the output.
type record struct {
	ID     string         `json:"id"`
	Blocks []blocks.Block `json:"blocks"`
}

// tally counts what was found, for the summary.
type tally struct {
	ads, empty         int
	kinds              map[blocks.Kind]int
	items              int
	heading, bold, led int
	gap                int
}

func main() {
	log.SetFlags(0)
	if len(os.Args) != 3 {
		log.Fatal("usage: blocks ADS.jsonl OUTDIR")
	}
	source, target := os.Args[1], filepath.Join(os.Args[2], "ads.jsonl")
	t, err := run(source, target)
	if err != nil {
		os.Remove(target) // half an output is worse than none
		log.Fatal(err)
	}
	t.print(target)
}

func run(source, target string) (t tally, err error) {
	t.kinds = map[blocks.Kind]int{}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return t, err
	}
	out, err := os.Create(target)
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

	for ad, err := range ads.File(source) {
		if err != nil {
			return t, err
		}
		bs := blocks.Split(ad.HTML)
		t.add(bs)
		if err := enc.Encode(record{ID: ad.ID, Blocks: bs}); err != nil {
			return t, err
		}
	}
	return t, buf.Flush()
}

func (t *tally) add(bs []blocks.Block) {
	t.ads++
	if len(bs) == 0 {
		t.empty++
	}
	cues := func(c blocks.Cues) {
		if c.Heading {
			t.heading++
		}
		if c.Bold {
			t.bold++
		}
		if c.BoldLead != "" {
			t.led++
		}
		if c.Gap {
			t.gap++
		}
	}
	for _, b := range bs {
		t.kinds[b.Kind]++
		cues(b.Cues)
		for _, it := range b.Items {
			t.items++
			cues(it.Cues)
		}
	}
}

func (t tally) print(target string) {
	fmt.Printf("%d ads written to %s", t.ads, target)
	if t.empty > 0 {
		fmt.Printf(" (%d with no blocks)", t.empty)
	}
	fmt.Println()
	fmt.Println()
	for _, k := range []blocks.Kind{blocks.Paragraph, blocks.Line, blocks.List} {
		fmt.Printf("%-10s %6d\n", k, t.kinds[k])
	}
	fmt.Printf("%-10s %6d\n\n", "items", t.items)
	fmt.Println("cues, over paragraphs, lines and items:")
	fmt.Printf("%-10s %6d\n", "heading", t.heading)
	fmt.Printf("%-10s %6d\n", "bold", t.bold)
	fmt.Printf("%-10s %6d\n", "bold lead", t.led)
	fmt.Printf("%-10s %6d\n", "gap", t.gap)
}
