// Package ads is the lab's record of a job ad, as data/ingestion/ads.jsonl
// holds it: one JSON object per line.
package ads

import (
	"encoding/json"
	"errors"
	"io"
	"iter"
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
// at the end of r, or after yielding the first error.
func All(r io.Reader) iter.Seq2[Ad, error] {
	return func(yield func(Ad, error) bool) {
		// A json.Decoder reads one value after another, so JSON Lines needs
		// nothing more, however long a line is.
		d := json.NewDecoder(r)
		for {
			var ad Ad
			err := d.Decode(&ad)
			if errors.Is(err, io.EOF) {
				return
			}
			if !yield(ad, err) || err != nil {
				return
			}
		}
	}
}
