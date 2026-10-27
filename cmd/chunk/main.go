// Command chunk prints the chunks of job ads, as an embedding job would
// ask for them (see internal/chunk): each section with its kind, what to
// do with it (embed it, keep it as metadata, or leave it out), about how
// many tokens it is, and what placed it in its section.
//
//	chunk [-json] [-run DIR] [-title TITLE] AD...
//
// An AD is an ad's ID in the run's ads, or a file of an ad's HTML, or "-"
// for HTML on standard input; -title gives a file's title, which the
// chunker uses to find the ad repeating it. Boilerplate is what the run's
// boilerplate/repeated.tsv says it is, so run the pipeline first. It
// prints each ad's chunks as text to read, or with -json as one JSON
// object a line.
//
// From the lab's folder:
//
//	go run ./cmd/chunk 225682916
//	go run ./cmd/chunk -json -title "Driver" ad.html
//
// or, once installed with `go install ./cmd/...`, just chunk.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nukopian/jdc-lab/internal/ads"
	"github.com/nukopian/jdc-lab/internal/blocks"
	"github.com/nukopian/jdc-lab/internal/chunk"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
)

// chunked is an ad's chunks, as -json writes them.
type chunked struct {
	ID     string        `json:"id"`
	Title  string        `json:"title"`
	Chunks []chunk.Chunk `json:"chunks"`
	// Tokens is about how many tokens the whole ad is, and Embedded how
	// many of them its embedded chunks are.
	Tokens   int `json:"tokens"`
	Embedded int `json:"embedded_tokens"`
}

func main() {
	log.SetFlags(0)
	asJSON := flag.Bool("json", false, "write each ad as a line of JSON")
	run := flag.String("run", "data/runs/latest", "the run whose ads and boilerplate to use")
	title := flag.String("title", "", "the title of an ad given as HTML")
	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatal("usage: chunk [-json] [-run DIR] [-title TITLE] AD...")
	}
	boiler, err := readBoilerplate(filepath.Join(*run, "boilerplate", "repeated.tsv"))
	if err != nil {
		log.Fatal(err)
	}
	isBoilerplate := func(s string) bool { return boiler[headings.Key(s)] }

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	for _, arg := range flag.Args() {
		a, err := find(arg, *run, *title)
		if err != nil {
			log.Fatal(err)
		}
		ss := sections.Divide(headings.Read(blocks.Split(a.HTML), a.Title, a.Locations()...))
		c := chunked{ID: a.ID, Title: a.Title, Chunks: chunk.Of(ss, isBoilerplate)}
		for _, s := range ss {
			c.Tokens += chunk.Tokens(s.Heading)
			for _, p := range s.Parts {
				c.Tokens += chunk.Tokens(p.Text)
			}
		}
		for _, ch := range c.Chunks {
			if ch.Use == chunk.Embed {
				c.Embedded += ch.Tokens
			}
		}
		if *asJSON {
			enc := json.NewEncoder(w)
			enc.SetEscapeHTML(false)
			if err := enc.Encode(c); err != nil {
				log.Fatal(err)
			}
			continue
		}
		write(w, c)
	}
}

// find returns the ad arg names: a file of HTML, "-" for standard input,
// or an ID in the run's ads.
func find(arg, run, title string) (ads.Ad, error) {
	if arg == "-" {
		b, err := io.ReadAll(os.Stdin)
		return ads.Ad{ID: "-", Title: title, HTML: string(b)}, err
	}
	if b, err := os.ReadFile(arg); err == nil {
		return ads.Ad{ID: filepath.Base(arg), Title: title, HTML: string(b)}, nil
	}
	for a, err := range ads.File(filepath.Join(run, "ads.jsonl")) {
		if err != nil {
			return ads.Ad{}, err
		}
		if a.ID == arg {
			return a, nil
		}
	}
	return ads.Ad{}, fmt.Errorf("chunk: %s: no such file, nor an ad of that ID in %s", arg, run)
}

// readBoilerplate reads a repeated.tsv, and returns the keys of the texts
// it found to be boilerplate.
func readBoilerplate(path string) (map[string]bool, error) {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("chunk: %s: no such file; run the pipeline first (go run ./cmd/pipeline run latest)", path)
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	keys := map[string]bool{}
	sc := bufio.NewScanner(f)
	sc.Buffer(nil, 1<<20)
	for sc.Scan() {
		f := strings.SplitN(sc.Text(), "\t", 6)
		if len(f) == 6 && f[4] == "boilerplate" {
			keys[headings.Key(f[5])] = true
		}
	}
	return keys, sc.Err()
}

// write writes an ad's chunks as text to read.
func write(w io.Writer, c chunked) {
	share := 0
	if c.Tokens > 0 {
		share = 100 * c.Embedded / c.Tokens
	}
	fmt.Fprintf(w, "%s · %s\n", c.Title, c.ID)
	fmt.Fprintf(w, "about %d tokens in all; %d embedded (%d%%)\n", c.Tokens, c.Embedded, share)
	for i, ch := range c.Chunks {
		fmt.Fprintf(w, "\n[%d] %s · %s · about %d tokens\n    placed by %s\n", i+1, ch.Kind, ch.Use, ch.Tokens, ch.PlacedBy)
		for _, line := range strings.Split(ch.Text, "\n") {
			fmt.Fprintf(w, "    | %s\n", line)
		}
		for _, f := range ch.Fields {
			fmt.Fprintf(w, "    field %s: %s\n", f.Label, f.Value)
		}
		for _, b := range ch.Boilerplate {
			fmt.Fprintf(w, "    boilerplate, left out: %s\n", b)
		}
	}
	fmt.Fprintln(w)
}
