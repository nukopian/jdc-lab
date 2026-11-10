package gazetteer

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/paulmach/orb"
	"github.com/paulmach/osm"
)

// square is a square outline with its south-west corner at lon, lat.
func square(lon, lat, side float64) orb.MultiPolygon {
	return orb.MultiPolygon{{{
		{lon, lat}, {lon + side, lat}, {lon + side, lat + side}, {lon, lat + side}, {lon, lat},
	}}}
}

func area(id, name string, level int, mp orb.MultiPolygon) *Area {
	return &Area{ID: id, Name: name, Names: []string{name}, Kind: "administrative", Level: level, Polygon: mp, Bound: mp.Bound()}
}

func place(id, name, kind string, lon, lat float64) *Place {
	return &Place{ID: id, Name: name, Names: []string{name}, Kind: kind, Point: orb.Point{lon, lat}}
}

// testGazetteer is two counties side by side, each with an Oakley in it,
// one with a district named for its town, a hamlet called Leeds besides
// the city, and areas and places named in the ways the ads name them
// differently.
func testGazetteer() *Gazetteer {
	g := &Gazetteer{
		Areas: []*Area{
			area("relation/1", "Hampshire", 6, square(0, 0, 1)),
			area("relation/2", "Kent", 6, square(1, 0, 1)),
			area("relation/3", "Basingstoke and Deane", 8, square(0, 0, 0.5)),
			area("relation/4", "Leeds", 8, square(5, 5, 0.5)),
			area("relation/5", "Borough of Wigan", 8, square(3, 3, 0.5)),
			area("relation/6", "Tameside", 8, square(7, 7, 0.5)),
			area("relation/7", "London Borough of Brent", 8, square(11, 11, 0.5)),
			area("relation/8", "York", 6, square(13, 13, 0.5)),
		},
		Places: []*Place{
			place("node/1", "Oakley", "village", 0.2, 0.2),
			place("node/2", "Oakley", "hamlet", 1.5, 0.5),
			place("node/3", "Leeds", "city", 5.2, 5.2),
			place("node/4", "Leeds", "hamlet", 1.6, 0.6),
			place("node/5", "St Helens", "town", 9, 9),
			place("node/6", "Crooke", "hamlet", 3.1, 3.1),
			place("node/7", "Stalybridge", "town", 7.1, 7.1),
			place("node/8", "Bishop's Stortford", "town", 20, 20),
			place("node/9", "Enfield", "town", 21, 21),
			place("node/10", "City of London", "city", 22, 22),
			place("node/11", "York", "city", 13.1, 13.1),
			place("node/12", "Regents Park", "suburb", 30, 30),
			place("node/13", "Regent's Park", "suburb", 31, 31),
		},
	}
	locate(g)
	g.index()
	return g
}

func TestKey(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"St. Helens", "st helens"},
		{"St Helens", "st helens"},
		{"Bath & North East Somerset", "bath and north east somerset"},
		{"Newcastle-upon-Tyne", "newcastle upon tyne"},
		{"Bishop's Stortford", "bishops stortford"},
		{"Bishop’s Stortford", "bishops stortford"},
		{"  Ynys Môn ", "ynys môn"},
		{"", ""},
	} {
		if got := Key(tc.in); got != tc.want {
			t.Errorf("Key(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestLocate(t *testing.T) {
	g := testGazetteer()
	oakley := g.Places[0]
	if len(oakley.Areas) != 2 || oakley.Areas[0].Name != "Basingstoke and Deane" || oakley.Areas[1].Name != "Hampshire" {
		t.Errorf("Oakley's areas = %v, want Basingstoke and Deane then Hampshire", names2(oakley.Areas))
	}
	leeds := g.Places[2]
	if leeds.PolygonFrom != "relation/4" {
		t.Errorf("Leeds's outline is from %q, want its district, relation/4", leeds.PolygonFrom)
	}
	if hamlet := g.Places[3]; hamlet.Polygon != nil {
		t.Errorf("the hamlet Leeds has an outline, from %q, but lies in no area of its name", hamlet.PolygonFrom)
	}
}

func names2(as []*Area) []string {
	var ns []string
	for _, a := range as {
		ns = append(ns, a.Name)
	}
	return ns
}

func TestMatch(t *testing.T) {
	g := testGazetteer()
	for _, tc := range []struct {
		name, status, id string
	}{
		{"Oakley, Hampshire", Matched, "node/1"},
		{"Oakley, Kent", Matched, "node/2"},
		{"Oakley, Basingstoke and Deane", Matched, "node/1"},
		{"Oakley", Chosen, "node/1"},           // the village before the hamlet
		{"Leeds", Chosen, "node/3"},            // the city before the hamlet
		{"Oakley, Devon", OutOfArea, "node/1"}, // no Devon: the biggest Oakley
		{"St. Helens", Matched, "node/5"},
		{"Crooke, Borough of Wigan", Matched, "node/6"},
		{"Stalybridge, Borough of Tameside", Matched, "node/7"}, // "Tameside" in OSM
		{"York, City of York", Matched, "node/11"},              // "York" in OSM
		{"Bishops Stortford", Matched, "node/8"},                // "Bishop's Stortford"
		{"Enfield Town", Matched, "node/9"},                     // "Enfield"
		{"Regent's Park", Chosen, "node/13"},                    // as written, before "Regents Park"
		{"Regents Park", Chosen, "node/12"},
		{"St Helens", Matched, "node/5"},
		{"City of London", Matched, "node/10"}, // not London
		{"Brent", AreaOnly, ""},                // "London Borough of Brent"
		{"Hampshire", AreaOnly, ""},
		{"Atlantis", Unmatched, ""},
	} {
		m := g.Match(tc.name)
		var id string
		if m.Place != nil {
			id = m.Place.ID
		}
		if m.Status != tc.status || id != tc.id {
			t.Errorf("Match(%q) = %s %q, want %s %q", tc.name, m.Status, id, tc.status, tc.id)
		}
	}
	if m := g.Match("Hampshire"); m.Area == nil || m.Area.ID != "relation/1" {
		t.Errorf("Match(%q).Area = %v, want relation/1", "Hampshire", m.Area)
	}
}

func TestRelationPolygon(t *testing.T) {
	// A square boundary drawn as two ways, each half of it, the second
	// running the opposite way round, as real boundaries often do. Its
	// corners are away from (0, 0), which osmgeojson takes for a node
	// with no point.
	nodes := func(pts ...orb.Point) osm.WayNodes {
		var wn osm.WayNodes
		for i, p := range pts {
			wn = append(wn, osm.WayNode{ID: osm.NodeID(100 + i), Lon: p[0], Lat: p[1]})
		}
		return wn
	}
	w1 := &osm.Way{ID: 1, Nodes: nodes(orb.Point{-1, 51}, orb.Point{0, 51}, orb.Point{0, 52})}
	w2 := &osm.Way{ID: 2, Nodes: nodes(orb.Point{-1, 51}, orb.Point{-1, 52}, orb.Point{0, 52})}
	r := &osm.Relation{
		ID:   9,
		Tags: osm.Tags{{Key: "type", Value: "boundary"}, {Key: "boundary", Value: "administrative"}, {Key: "admin_level", Value: "8"}, {Key: "name", Value: "Squareshire"}},
		Members: osm.Members{
			{Type: osm.TypeWay, Ref: 1, Role: "outer"},
			{Type: osm.TypeWay, Ref: 2, Role: "outer"},
		},
	}
	mp := relationPolygon(r, map[osm.WayID]*osm.Way{1: w1, 2: w2})
	if mp == nil {
		t.Fatal("relationPolygon gave no outline")
	}
	if !contains(mp, orb.Point{-0.5, 51.5}) || contains(mp, orb.Point{0.5, 51.5}) {
		t.Errorf("outline %v doesn't hold (-0.5, 51.5) alone", mp)
	}
	if !areaTags(r.Tags) {
		t.Error("areaTags refused an admin_level 8 boundary")
	}

	// With a way missing, as at the edge of an extract, there's no ring.
	if mp := relationPolygon(r, map[osm.WayID]*osm.Way{1: w1}); mp != nil {
		t.Errorf("relationPolygon with half a boundary gave %v, want nil", mp)
	}
}

func TestMergeOutlines(t *testing.T) {
	node := place("node/1", "Hulme", "suburb", 0.5, 0.5)
	outline := place("way/1", "Hulme", "suburb", 0.4, 0.4)
	outline.Polygon, outline.PolygonFrom = square(0, 0, 1), "way/1"
	lone := place("way/2", "Elsewhere", "suburb", 3.5, 3.5)
	lone.Polygon, lone.PolygonFrom = square(3, 3, 1), "way/2"

	left := mergeOutlines([]*Place{node}, []*Place{outline, lone})
	if node.PolygonFrom != "way/1" {
		t.Errorf("the node's outline is from %q, want way/1", node.PolygonFrom)
	}
	if len(left) != 1 || left[0] != lone {
		t.Errorf("mergeOutlines left %d outlines, want only Elsewhere's", len(left))
	}
}

func TestSaveLoad(t *testing.T) {
	ctx := context.Background()
	g := testGazetteer()
	path := filepath.Join(t.TempDir(), "gazetteer.sqlite")
	if err := g.Save(ctx, path); err != nil {
		t.Fatal(err)
	}
	h, err := Load(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if len(h.Places) != len(g.Places) || len(h.Areas) != len(g.Areas) {
		t.Fatalf("loaded %d places and %d areas, want %d and %d", len(h.Places), len(h.Areas), len(g.Places), len(g.Areas))
	}
	for _, name := range []string{"Oakley, Kent", "Leeds", "Crooke, Borough of Wigan", "Hampshire"} {
		want, got := g.Match(name), h.Match(name)
		if got.Status != want.Status || (want.Place != nil && (got.Place == nil || got.Place.ID != want.Place.ID)) {
			t.Errorf("after loading, Match(%q) = %s, want %s", name, got.Status, want.Status)
		}
	}
	for _, p := range h.Places {
		if p.ID == "node/3" && (p.Polygon == nil || p.PolygonFrom != "relation/4") {
			t.Errorf("after loading, Leeds's outline is %v from %q, want relation/4's", p.Polygon != nil, p.PolygonFrom)
		}
	}
	if _, err := Load(ctx, filepath.Join(t.TempDir(), "missing.sqlite")); err == nil {
		t.Error("Load of a missing file gave no error")
	}
}
