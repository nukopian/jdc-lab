package gazetteer

import (
	"cmp"
	"slices"
	"strings"
)

// The statuses a Match can have, best first.
const (
	Matched   = "matched"     // one place of the name, in the area named if any
	Chosen    = "chosen"      // several, and the biggest kind was taken
	OutOfArea = "out-of-area" // places of the name, none in the area named; the biggest taken
	AreaOnly  = "area"        // no place of the name, but an area of it
	Unmatched = "unmatched"   // nothing of the name
)

// A Match is what a name was found to be.
type Match struct {
	Name   string // as given
	Status string // one of the statuses above
	Place  *Place // the place taken, nil unless one was
	Area   *Area  // the area taken, when Status is AreaOnly

	Candidates int // the places of the name
	InArea     int // of those, the ones in the area named, or all if none was
}

// kindRank orders places of the same name, biggest first, so that a bare
// "Leeds" is the city, not a hamlet somewhere that shares its name.
var kindRank = map[string]int{
	"city": 0, "town": 1, "borough": 2, "suburb": 3, "quarter": 4, "village": 5,
	"neighbourhood": 6, "hamlet": 7, "locality": 8, "isolated_dwelling": 9,
}

// Match finds name, of the form "Place" or "Place, Area". The area, if
// given, is any area around the place's point that has that name, at
// any level: "Kent", "Borough of Wigan" or "Greater London", with or
// without a prefix such as "Borough of". A place that isn't found is
// looked for again without a trailing "Town". A name of
// more than two parts takes its last as the area and the rest as the
// place.
func (g *Gazetteer) Match(name string) Match {
	m := Match{Name: name}
	place, area := name, ""
	if i := strings.LastIndex(name, ", "); i >= 0 {
		place, area = name[:i], name[i+2:]
	}
	pk, ak := Key(place), Key(area)

	candidates := g.placesByKey[pk]
	// "Enfield Town" is OpenStreetMap's "Enfield".
	if rest, ok := strings.CutSuffix(pk, " town"); ok && len(candidates) == 0 && rest != "" {
		candidates = g.placesByKey[rest]
	}
	m.Candidates = len(candidates)
	inArea := candidates
	if ak != "" {
		inArea = nil
		for _, p := range candidates {
			if slices.ContainsFunc(p.Areas, func(a *Area) bool { return areaNamed(a, ak) }) {
				inArea = append(inArea, p)
			}
		}
	}
	m.InArea = len(inArea)

	switch {
	case len(inArea) == 1:
		m.Status, m.Place = Matched, inArea[0]
	case len(inArea) > 1:
		m.Status, m.Place = Chosen, biggest(inArea, place)
	case len(candidates) > 0:
		m.Status, m.Place = OutOfArea, biggest(candidates, place)
	default:
		as := g.areasByKey[pk]
		if len(as) == 0 {
			as = g.areasByKey[bareKey(pk)]
		}
		if len(as) > 0 {
			m.Status, m.Area = AreaOnly, smallestLevel(as)
		} else {
			m.Status = Unmatched
		}
	}
	return m
}

// biggest is the place named as written, apostrophes and all, if any is,
// as Key joins "Regent's Park" with "Regents Park"; then
// of the biggest kind; then with the most people; then the first ID, so
// the choice is the same every run.
func biggest(ps []*Place, written string) *Place {
	wk := wordKey(written)
	asWritten := func(p *Place) int {
		if slices.ContainsFunc(p.Names, func(n string) bool { return wordKey(n) == wk }) {
			return 0
		}
		return 1
	}
	return slices.MinFunc(ps, func(a, b *Place) int {
		return cmp.Or(
			cmp.Compare(asWritten(a), asWritten(b)),
			cmp.Compare(rank(a.Kind), rank(b.Kind)),
			cmp.Compare(b.Population, a.Population),
			cmp.Compare(a.ID, b.ID),
		)
	})
}

func rank(kind string) int {
	if r, ok := kindRank[kind]; ok {
		return r
	}
	return len(kindRank)
}

// smallestLevel is the area of the highest admin level, the most local,
// then the first ID.
func smallestLevel(as []*Area) *Area {
	return slices.MinFunc(as, func(a, b *Area) int {
		return cmp.Or(cmp.Compare(b.Level, a.Level), cmp.Compare(a.ID, b.ID))
	})
}
