// Command chunk prints the chunks of job ads, as an embedding job would
// ask for them (see internal/chunk): each section with its kind, what to
// do with it (embed it, keep it as metadata, or leave it out), about how
// many tokens it is, and what placed it in its section; boilerplate and
// unmapped text are gathered in a chunk of their own.
//
//	chunk [-json | -yaml] [-run DIR] [-title TITLE] AD...
//	chunk -pages FILE [-run DIR]
//
// An AD is an ad's ID in the run's ads, or a file of an ad's HTML, or "-"
// for HTML on standard input; -title gives a file's title, which the
// chunker uses to find the ad repeating it. Boilerplate is what the run's
// boilerplate/repeated.tsv says it is, so run the pipeline first. That's
// quicker than reading every ad again, as the boilerplate stage does, but
// it misses what that stage finds with the ad's own title and place
// taken out, or as a near-copy: the view, which reads every ad, finds
// those too, so its chunks can gather more boilerplate than these. It
// prints each ad's chunks as text to read; with -json as one JSON object
// a line, for programs; or with -yaml as a YAML document each, for
// people, each chunk's text a block of its lines.
//
// With -pages, it writes the ads FILE lists as web pages, each ad as
// posted beside its chunks (see internal/page), into FILE's folder: a
// page for each, ID.html, and index.html, listing them. Each line of
// FILE is an ad's ID, a tab, and why it's there; a line starting "#" is
// a comment.
//
// From the lab's folder:
//
//	go run ./cmd/chunk 225682916
//	go run ./cmd/chunk -yaml 225682916
//	go run ./cmd/chunk -json -title "Driver" ad.html
//	go run ./cmd/chunk -pages data/pages/ads.tsv
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
	"github.com/nukopian/jdc-lab/internal/boilerplate"
	"github.com/nukopian/jdc-lab/internal/chunk"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/page"
	"github.com/nukopian/jdc-lab/internal/sections"
)

func main() {
	log.SetFlags(0)
	asJSON := flag.Bool("json", false, "write each ad as a line of JSON")
	asYAML := flag.Bool("yaml", false, "write each ad as a YAML document")
	run := flag.String("run", "data/runs/latest", "the run whose ads and boilerplate to use")
	title := flag.String("title", "", "the title of an ad given as HTML")
	pages := flag.String("pages", "", "write the ads this file lists as pages, beside it")
	flag.Parse()
	if *pages == "" && flag.NArg() == 0 {
		log.Fatal("usage: chunk [-json | -yaml] [-run DIR] [-title TITLE] AD...\n       chunk -pages FILE [-run DIR]")
	}
	if *pages != "" && (flag.NArg() > 0 || *asJSON || *asYAML || *title != "") {
		log.Fatal("chunk: -pages takes the ads from its file, with -run only")
	}
	if *asJSON && *asYAML {
		log.Fatal("chunk: -json or -yaml, not both")
	}
	boiler, err := readBoilerplate(filepath.Join(*run, "boilerplate", "repeated.tsv"))
	if err != nil {
		log.Fatal(err)
	}
	isBoilerplate := func(s string) bool { return boiler[headings.Key(s)] }
	chunked := func(a ads.Ad) chunk.Ad {
		ss := sections.Divide(headings.Read(blocks.Split(a.HTML), a.Title, a.Locations()...))
		return chunk.Totalled(chunk.Ad{ID: a.ID, Title: a.Title, Chunks: chunk.Of(ss, isBoilerplate)}, chunk.Whole(ss))
	}
	if *pages != "" {
		if err := writePages(*pages, *run, chunked); err != nil {
			log.Fatal(err)
		}
		return
	}

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	for _, arg := range flag.Args() {
		a, err := find(arg, *run, *title)
		if err != nil {
			log.Fatal(err)
		}
		c := chunked(a)
		if *asYAML {
			fmt.Fprintln(w, "---")
			if err := c.WriteYAML(w); err != nil {
				log.Fatal(err)
			}
			continue
		}
		if *asJSON {
			enc := json.NewEncoder(w)
			enc.SetEscapeHTML(false)
			if err := enc.Encode(c); err != nil {
				log.Fatal(err)
			}
			continue
		}
		if err := c.WriteText(w); err != nil {
			log.Fatal(err)
		}
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
	a, err := ads.Find(filepath.Join(run, "ads.jsonl"), arg)
	if errors.Is(err, ads.ErrNotFound) {
		return a, fmt.Errorf("chunk: %s: no such file, nor an ad of that ID in %s", arg, run)
	}
	return a, err
}

// writePages writes the ads the file at list names as pages, beside it,
// with an index.
func writePages(list, run string, chunked func(ads.Ad) chunk.Ad) error {
	b, err := os.ReadFile(list)
	if err != nil {
		return err
	}
	why := map[string]string{}
	var ids []string
	for _, l := range strings.Split(string(b), "\n") {
		if l = strings.TrimSpace(l); l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		id, reason, _ := strings.Cut(l, "\t")
		id = strings.TrimSpace(id)
		if _, dup := why[id]; dup {
			return fmt.Errorf("chunk: %s: %s is listed twice", list, id)
		}
		if id != filepath.Base(id) || strings.HasPrefix(id, ".") || strings.Contains(id, " ") {
			return fmt.Errorf("chunk: %s: %q isn't an ad's ID; is a tab missing after it?", list, id)
		}
		ids = append(ids, id)
		why[id] = strings.TrimSpace(reason)
	}
	found := map[string]ads.Ad{}
	for a, err := range ads.File(filepath.Join(run, "ads.jsonl")) {
		if err != nil {
			return err
		}
		if _, ok := why[a.ID]; ok {
			if _, seen := found[a.ID]; !seen {
				found[a.ID] = a
			}
		}
	}
	const title = "Job ads before and after chunking"
	dir := filepath.Dir(list)
	var es []page.Entry
	for _, id := range ids {
		a, ok := found[id]
		if !ok {
			return fmt.Errorf("chunk: %s: no ad of that ID in %s", id, run)
		}
		e := page.Entry{Ad: page.Ad{Ad: a, Chunks: chunked(a), Why: why[id]}, File: id + ".html"}
		if err := writeFile(filepath.Join(dir, e.File), func(w io.Writer) error { return page.Write(w, e.Ad, title) }); err != nil {
			return err
		}
		es = append(es, e)
	}
	if err := writeFile(filepath.Join(dir, "index.html"), func(w io.Writer) error { return page.WriteIndex(w, title, es) }); err != nil {
		return err
	}
	fmt.Printf("%d pages and their index written to %s\n", len(es), dir)
	return nil
}

// writeFile creates the file at path, and has write fill it.
func writeFile(path string, write func(io.Writer) error) (err error) {
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
	if err := write(w); err != nil {
		return err
	}
	return w.Flush()
}

// readBoilerplate reads a repeated.tsv, and returns the keys of the texts
// it found to be boilerplate.
func readBoilerplate(path string) (map[string]bool, error) {
	keys, err := boilerplate.Verdicts(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("chunk: %s: no such file; run the pipeline first (go run ./cmd/pipeline run latest)", path)
	}
	return keys, err
}

// write writes an ad's chunks as text to read.
