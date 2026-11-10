package gazetteer

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/planar"
	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmgeojson"
	"github.com/paulmach/osm/osmpbf"
)

// placeKinds are the place=* values kept as places. Others (farm, island,
// square and the like) aren't where a job would say it is.
var placeKinds = map[string]bool{
	"city": true, "town": true, "village": true, "hamlet": true,
	"suburb": true, "quarter": true, "neighbourhood": true, "borough": true,
	"locality": true, "isolated_dwelling": true,
}

// nameTags are the tags a place or area's names come from, the first
// being its name. alt_name may hold several, separated by ";".
var nameTags = []string{"name", "name:en", "official_name", "alt_name", "short_name"}

// Build reads the OpenStreetMap extract at path (a .osm.pbf) and returns
// its places and areas, each place with the areas it lies in.
//
// It reads the file three times, as a PBF holds its nodes, then ways,
// then relations, and each needs the one before: the relations of the
// areas and place outlines first, then the ways they are made of, and
// last the nodes, both the places and the points of those ways. Keeping
// only those keeps the memory under 2 GB for the United Kingdom, built in
// under three minutes on an M2.
func Build(ctx context.Context, path string) (*Gazetteer, error) {
	// Pass 1: the relations wanted, and the ways they're made of.
	var relations []*osm.Relation
	wayWanted := make(map[osm.WayID]bool)
	err := scan(ctx, path, func(s *osmpbf.Scanner) {
		s.SkipNodes, s.SkipWays = true, true
		s.FilterRelation = func(r *osm.Relation) bool { return areaTags(r.Tags) || placeArea(r.Tags) }
	}, func(o osm.Object) {
		r := o.(*osm.Relation)
		relations = append(relations, r)
		for _, m := range r.Members {
			if m.Type == osm.TypeWay && (m.Role == "outer" || m.Role == "inner") {
				wayWanted[osm.WayID(m.Ref)] = true
			}
		}
	})
	if err != nil {
		return nil, err
	}

	// Pass 2: those ways, and the closed ways that are places' outlines.
	// FilterWay runs on the scanner's decoding goroutines, which is safe
	// as wayWanted is only read from now on.
	ways := make(map[osm.WayID]*osm.Way)
	var nodeIDs []osm.NodeID
	err = scan(ctx, path, func(s *osmpbf.Scanner) {
		s.SkipNodes, s.SkipRelations = true, true
		s.FilterWay = func(w *osm.Way) bool {
			return wayWanted[w.ID] || (closed(w) && placeArea(w.Tags))
		}
	}, func(o osm.Object) {
		w := o.(*osm.Way)
		ways[w.ID] = w
		for _, n := range w.Nodes {
			nodeIDs = append(nodeIDs, n.ID)
		}
	})
	if err != nil {
		return nil, err
	}

	// Pass 3: the place nodes, and the points of the ways. A sorted slice
	// searched by halves takes a third of the memory a map would.
	slices.Sort(nodeIDs)
	nodeIDs = slices.Compact(nodeIDs)
	points := make([]orb.Point, len(nodeIDs))
	g := &Gazetteer{}
	err = scan(ctx, path, func(s *osmpbf.Scanner) {
		s.SkipWays, s.SkipRelations = true, true
		s.FilterNode = func(n *osm.Node) bool {
			_, wanted := slices.BinarySearch(nodeIDs, n.ID)
			return wanted || placeNode(n.Tags)
		}
	}, func(o osm.Object) {
		n := o.(*osm.Node)
		if i, ok := slices.BinarySearch(nodeIDs, n.ID); ok {
			points[i] = n.Point()
		}
		if placeNode(n.Tags) {
			g.Places = append(g.Places, newPlace(fmt.Sprintf("node/%d", n.ID), n.Tags, n.Point()))
		}
	})
	if err != nil {
		return nil, err
	}
	for _, w := range ways {
		for i := range w.Nodes {
			if j, ok := slices.BinarySearch(nodeIDs, w.Nodes[i].ID); ok {
				w.Nodes[i].Lon, w.Nodes[i].Lat = points[j][0], points[j][1]
			}
		}
	}

	// The areas, and the places drawn as areas.
	var outlines []*Place
	for _, r := range relations {
		mp := relationPolygon(r, ways)
		if mp == nil {
			continue
		}
		id := fmt.Sprintf("relation/%d", r.ID)
		if areaTags(r.Tags) {
			g.Areas = append(g.Areas, newArea(id, r.Tags, mp))
		}
		if placeArea(r.Tags) {
			outlines = append(outlines, outlinePlace(id, r.Tags, mp))
		}
	}
	for _, w := range ways {
		if closed(w) && placeArea(w.Tags) {
			if mp := wayPolygon(w); mp != nil {
				outlines = append(outlines, outlinePlace(fmt.Sprintf("way/%d", w.ID), w.Tags, mp))
			}
		}
	}
	g.Places = append(g.Places, mergeOutlines(g.Places, outlines)...)

	locate(g)
	g.index()
	return g, nil
}

// scan reads the extract at path once, configured by configure, calling
// each for every object the scanner keeps.
func scan(ctx context.Context, path string, configure func(*osmpbf.Scanner), each func(osm.Object)) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	s := osmpbf.New(ctx, f, runtime.GOMAXPROCS(0))
	defer s.Close()
	configure(s)
	for s.Scan() {
		each(s.Object())
	}
	if err := s.Err(); err != nil {
		return fmt.Errorf("gazetteer: %s: %w", path, err)
	}
	return nil
}

// areaTags says whether tags make an area worth keeping: a named
// administrative boundary from region to parish, or a ceremonial county.
func areaTags(t osm.Tags) bool {
	if t.Find("name") == "" || !polygonType(t) {
		return false
	}
	switch t.Find("boundary") {
	case "administrative":
		level, err := strconv.Atoi(t.Find("admin_level"))
		return err == nil && level >= 5 && level <= 10
	case "ceremonial":
		return true
	}
	return false
}

// polygonType says whether a relation's type makes it a polygon. A way's
// tags have no type, and pass.
func polygonType(t osm.Tags) bool {
	switch t.Find("type") {
	case "", "multipolygon", "boundary":
		return true
	}
	return false
}

// placeNode says whether a node's tags make it a place.
func placeNode(t osm.Tags) bool {
	return placeKinds[t.Find("place")] && t.Find("name") != ""
}

// placeArea says whether a way or relation's tags make it a place drawn
// as an area.
func placeArea(t osm.Tags) bool {
	return placeNode(t) && polygonType(t)
}

// closed says whether a way ends where it starts, as an outline does.
func closed(w *osm.Way) bool {
	return len(w.Nodes) >= 4 && w.Nodes[0].ID == w.Nodes[len(w.Nodes)-1].ID
}

// names gives the names in t's nameTags, each once, its name first.
func names(t osm.Tags) []string {
	var ns []string
	for _, tag := range nameTags {
		for n := range strings.SplitSeq(t.Find(tag), ";") {
			if n = strings.TrimSpace(n); n != "" && !slices.Contains(ns, n) {
				ns = append(ns, n)
			}
		}
	}
	return ns
}

func newPlace(id string, t osm.Tags, p orb.Point) *Place {
	population, _ := strconv.Atoi(strings.ReplaceAll(t.Find("population"), ",", ""))
	return &Place{
		ID:         id,
		Name:       t.Find("name"),
		Names:      names(t),
		Kind:       t.Find("place"),
		Point:      p,
		Population: population,
	}
}

func newArea(id string, t osm.Tags, mp orb.MultiPolygon) *Area {
	level, _ := strconv.Atoi(t.Find("admin_level"))
	if t.Find("boundary") == "ceremonial" {
		level = 0
	}
	return &Area{
		ID:      id,
		Name:    t.Find("name"),
		Names:   names(t),
		Kind:    t.Find("boundary"),
		Level:   level,
		Polygon: mp,
		Bound:   mp.Bound(),
	}
}

// outlinePlace is a place drawn as an area, its point the outline's
// centroid until mergeOutlines finds it a node.
func outlinePlace(id string, t osm.Tags, mp orb.MultiPolygon) *Place {
	centroid, _ := planar.CentroidArea(mp)
	p := newPlace(id, t, centroid)
	p.Polygon, p.PolygonFrom = mp, id
	return p
}

// relationPolygon assembles a multipolygon or boundary relation's outline
// from its ways, whose nodes carry their points, or returns nil if it
// can't: if the extract cut it off at its edge, say. osmgeojson does the
// work of joining the ways into rings.
func relationPolygon(r *osm.Relation, ways map[osm.WayID]*osm.Way) orb.MultiPolygon {
	o := &osm.OSM{Relations: osm.Relations{r}}
	for _, m := range r.Members {
		if w := ways[osm.WayID(m.Ref)]; m.Type == osm.TypeWay && w != nil {
			o.Ways = append(o.Ways, w)
		}
	}
	// A boundary's type is "boundary", which osmgeojson takes for a
	// polygon, as it does "multipolygon".
	fc, err := osmgeojson.Convert(o, osmgeojson.NoMeta(true), osmgeojson.NoRelationMembership(true))
	if err != nil || len(fc.Features) == 0 {
		return nil
	}
	// The relation's feature comes first, before its ways'.
	return multi(fc.Features[0].Geometry)
}

// wayPolygon is a closed way's outline, or nil if its points are missing.
func wayPolygon(w *osm.Way) orb.MultiPolygon {
	ring := make(orb.Ring, 0, len(w.Nodes))
	for _, n := range w.Nodes {
		if n.Lon == 0 && n.Lat == 0 {
			return nil
		}
		ring = append(ring, orb.Point{n.Lon, n.Lat})
	}
	return orb.MultiPolygon{{ring}}
}

// multi makes a polygon a multipolygon of one, and anything but a polygon
// nil.
func multi(g orb.Geometry) orb.MultiPolygon {
	switch g := g.(type) {
	case orb.Polygon:
		return orb.MultiPolygon{g}
	case orb.MultiPolygon:
		return g
	}
	return nil
}

// mergeOutlines gives each outline to the place node of the same name
// that lies inside it, as OpenStreetMap often maps a suburb or town both
// as a node and as an area. It returns the outlines that found no node,
// which are places in their own right.
func mergeOutlines(nodes, outlines []*Place) []*Place {
	byKey := make(map[string][]*Place)
	for _, n := range nodes {
		byKey[Key(n.Name)] = append(byKey[Key(n.Name)], n)
	}
	var left []*Place
	for _, o := range outlines {
		merged := false
		for _, n := range byKey[Key(o.Name)] {
			if n.Polygon == nil && contains(o.Polygon, n.Point) {
				n.Polygon, n.PolygonFrom = o.Polygon, o.PolygonFrom
				merged = true
				break
			}
		}
		if !merged {
			left = append(left, o)
		}
	}
	return left
}

// contains says whether mp holds p, checking its bound first, which is
// far cheaper.
func contains(mp orb.MultiPolygon, p orb.Point) bool {
	return mp.Bound().Contains(p) && planar.MultiPolygonContains(mp, p)
}
