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
	"strings"
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

// ErrNotFound is what Find's and FindAll's errors wrap when the file has
// no ad of an ID.
var ErrNotFound = errors.New("no ad of that ID")

// A NotFoundError is FindAll's error when the file has no ad of some of
// the IDs. It wraps ErrNotFound, so errors.Is finds that, and errors.As
// gives the IDs.
type NotFoundError struct {
	Path string
	IDs  []string // the IDs missing, in the order asked for
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s: %v: %s", e.Path, ErrNotFound, strings.Join(e.IDs, ", "))
}

// Unwrap returns ErrNotFound.
func (e *NotFoundError) Unwrap() error { return ErrNotFound }

// Find returns the first ad in the file at path with the ID id, reading as
// File does, or an error wrapping ErrNotFound if there's none.
func Find(path, id string) (Ad, error) {
	found, err := FindAll(path, []string{id})
	return found[id], err
}

// FindAll returns the first ad in the file at path with each of ids, by
// ID, reading the file once, as File does, and no further than the last
// of them. If any is missing, it returns those it found, and a
// *NotFoundError naming the rest.
func FindAll(path string, ids []string) (map[string]Ad, error) {
	want := make(map[string]bool, len(ids))
	for _, id := range ids {
		want[id] = true
	}
	found := make(map[string]Ad, len(want))
	for ad, err := range File(path) {
		if err != nil {
			return nil, err
		}
		if want[ad.ID] {
			found[ad.ID] = ad
			delete(want, ad.ID)
			if len(want) == 0 {
				return found, nil
			}
		}
	}
	if len(want) == 0 {
		return found, nil // no IDs were asked for
	}
	var missing []string
	for _, id := range ids {
		if want[id] {
			missing = append(missing, id)
			delete(want, id) // named once
		}
	}
	return found, &NotFoundError{Path: path, IDs: missing}
}
