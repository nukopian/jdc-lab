// Package gazetteer is a list of named places and the areas that hold
// them, built from an OpenStreetMap extract, for finding where a place
// name is: its point, and its polygon where OpenStreetMap has one.
//
// A place is anything OpenStreetMap tags place=*: a city, town, village,
// hamlet, suburb and so on, most of them a single node. An area is an
// administrative or ceremonial boundary, from region (admin_level 5) down
// to parish (10), always a polygon. Each place knows the areas its point
// lies in, so "Oakley, Hampshire" can be told from the Oakleys elsewhere.
//
// Build reads the extract, Save and Load keep the result in SQLite, and
// Match finds a name of the form "Place" or "Place, Area".
package gazetteer

import (
	"strings"
	"unicode"

	"github.com/paulmach/orb"
)

// A Place is a named place: a node, or a way or relation drawn as an
// area, tagged place=*.
type Place struct {
	ID         string    // "node/123", "way/456" or "relation/789"
	Name       string    // its name tag
	Names      []string  // every name it goes by: name, name:en, official_name, alt_name, short_name
	Kind       string    // the place tag's value: city, town, village, …
	Point      orb.Point // where it is, longitude first, as orb has it
	Population int       // from its population tag, 0 if none

	// Polygon is the place's outline, if it has one: its own, or that of
	// the smallest area of the same name around its point. PolygonFrom
	// is the ID of whichever it came from.
	Polygon     orb.MultiPolygon
	PolygonFrom string

	Areas []*Area // the areas its point lies in
}

// An Area is a named administrative or ceremonial boundary.
type Area struct {
	ID      string   // "relation/789"
	Name    string   // its name tag
	Names   []string // every name it goes by, as Place's
	Kind    string   // the boundary tag's value: administrative or ceremonial
	Level   int      // its admin_level, 0 for a ceremonial area
	Polygon orb.MultiPolygon
	Bound   orb.Bound
}

// A Gazetteer is the places and areas of one extract.
type Gazetteer struct {
	Places []*Place
	Areas  []*Area

	placesByKey map[string][]*Place
	areasByKey  map[string][]*Area
}

// index builds the lookups by name. Build and Load call it once their
// places and areas are complete.
func (g *Gazetteer) index() {
	g.placesByKey = make(map[string][]*Place)
	for _, p := range g.Places {
		for _, k := range keys(p.Names) {
			g.placesByKey[k] = append(g.placesByKey[k], p)
		}
	}
	g.areasByKey = make(map[string][]*Area)
	for _, a := range g.Areas {
		for _, k := range keys(a.Names) {
			g.areasByKey[k] = append(g.areasByKey[k], a)
			if bk := bareKey(k); bk != k {
				g.areasByKey[bk] = append(g.areasByKey[bk], a)
			}
		}
	}
}

// Key is the form of a name that lookups compare: lower case, "&" as
// "and", apostrophes gone, and every run of anything else but letters and
// digits as one space, so "St. Helens", "St Helens" and "st helens" are
// the same key, as are "Bishop's Stortford" and "Bishops Stortford".
func Key(name string) string {
	return wordKey(strings.NewReplacer("'", "", "’", "").Replace(name))
}

// wordKey is Key with apostrophes left in, so they part words: "Regent's
// Park" is "regent s park", unlike "Regents Park".
func wordKey(name string) string {
	name = strings.ReplaceAll(strings.ToLower(name), "&", " and ")
	words := strings.FieldsFunc(name, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	return strings.Join(words, " ")
}

// areaPrefixes are the forms an area's name takes in full, which some
// sources use and others don't: "Borough of Tameside" is "Tameside" in
// OpenStreetMap, and OpenStreetMap's "London Borough of Brent" is
// "Brent" elsewhere.
var areaPrefixes = []string{
	"city and county of ", "city and borough of ", "metropolitan borough of ",
	"london borough of ", "royal borough of ", "county borough of ",
	"borough of ", "county of ", "city of ", "district of ",
}

// bareKey is a key without its area prefix, if it has one: "city of york"
// is "york". Only areas are compared bare, as "City of London" is a
// place in its own right, not London.
func bareKey(k string) string {
	for _, p := range areaPrefixes {
		if rest, ok := strings.CutPrefix(k, p); ok && rest != "" {
			return rest
		}
	}
	return k
}

// areaNamed says whether a goes by the area name whose key is k, in full
// or bare.
func areaNamed(a *Area, k string) bool {
	bk := bareKey(k)
	for _, n := range a.Names {
		if nk := Key(n); nk == k || bareKey(nk) == bk {
			return true
		}
	}
	return false
}

// keys gives the distinct keys of names, in their order.
func keys(names []string) []string {
	var ks []string
	seen := make(map[string]bool, len(names))
	for _, n := range names {
		k := Key(n)
		if k != "" && !seen[k] {
			seen[k] = true
			ks = append(ks, k)
		}
	}
	return ks
}

// hasKey says whether any of names has the key k.
func hasKey(names []string, k string) bool {
	for _, n := range names {
		if Key(n) == k {
			return true
		}
	}
	return false
}
