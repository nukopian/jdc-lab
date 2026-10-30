// Package ads is the lab's record of a job ad, as each sample in
// data/ingestion holds it: one JSON object per line.
package ads

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
)

// An Ad is one job ad. Location and LocationCanonical are nil where the
// source had none; the rest are always there.
type Ad struct {
	ID                string  `json:"id"`
	Title             string  `json:"title"`
	Location          *string `json:"location"`
	LocationCanonical *string `json:"location_canonical"`
	HTML              string  `json:"html"`
}

// Locations returns the ad's locations that it has: Location, then
// LocationCanonical.
func (a Ad) Locations() []string {
	var ls []string
	for _, l := range []*string{a.Location, a.LocationCanonical} {
		if l != nil {
			ls = append(ls, *l)
		}
	}
	return ls
}

// All reads ads from r, one JSON object per line, for a range loop. It stops
// at the end of r, or after yielding the first error, which says which ad
// it's in.
func All(r io.Reader) iter.Seq2[Ad, error] {
	return func(yield func(Ad, error) bool) {
		// A json.Decoder reads one value after another, so JSON Lines needs
		// nothing more, however long a line is.
		d := json.NewDecoder(r)
		for n := 1; ; n++ {
			var ad Ad
			err := d.Decode(&ad)
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				yield(Ad{}, fmt.Errorf("ad %d: %w", n, err))
				return
			}
			if !yield(ad, nil) {
				return
			}
		}
	}
}

// File reads the ads in the file at path, as All does, and closes it at the
// end. Its errors name the file.
func File(path string) iter.Seq2[Ad, error] {
	return func(yield func(Ad, error) bool) {
		f, err := os.Open(path)
		if err != nil {
			yield(Ad{}, err) // the error names the file already
			return
		}
		defer f.Close()
		for ad, err := range All(f) {
			if err != nil {
				err = fmt.Errorf("%s: %w", path, err)
			}
			if !yield(ad, err) {
				return
			}
		}
	}
}

// Distinct tells an ad from the ones before it, by its HTML: an ad posted
// twice is the same HTML under a second id. Its zero value is ready to use.
type Distinct struct {
	seen map[string]bool
	// Duplicates counts the ads First said weren't.
	Duplicates int
}

// First reports whether no ad before a, of those given to First, had the
// same HTML.
func (d *Distinct) First(a Ad) bool {
	if d.seen[a.HTML] {
		d.Duplicates++
		return false
	}
	if d.seen == nil {
		d.seen = map[string]bool{}
	}
	d.seen[a.HTML] = true
	return true
}

// ErrNotFound is what Find's error wraps when the file has no ad of the ID.
var ErrNotFound = errors.New("no ad of that ID")

// Find returns the first ad in the file at path with the ID id, reading as
// File does, or an error wrapping ErrNotFound if there's none.
func Find(path, id string) (Ad, error) {
	for ad, err := range File(path) {
		if err != nil {
			return Ad{}, err
		}
		if ad.ID == id {
			return ad, nil
		}
	}
	return Ad{}, fmt.Errorf("%s: %w: %s", path, ErrNotFound, id)
}
