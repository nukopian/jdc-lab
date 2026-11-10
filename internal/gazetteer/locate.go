package gazetteer

import (
	"cmp"
	"math"
	"slices"

	"github.com/paulmach/orb/planar"
)

// cellSize is the side of a grid cell, in degrees: a quarter-degree is
// about 28 km north to south, so a parish spans a cell or two and a
// county a few dozen.
const cellSize = 0.25

type cell struct{ x, y int }

func cellOf(lon, lat float64) cell {
	return cell{int(math.Floor(lon / cellSize)), int(math.Floor(lat / cellSize))}
}

// locate gives each place the areas its point lies in, smallest first,
// and a place with no outline of its own the outline of the smallest
// area of its name around it: "Birmingham" the city's district, say.
//
// Testing every place against every area would be some 10⁹ tests for
// Great Britain, so the areas are first put in a grid of the cells their
// bounds cover, and a place is tested only against its cell's areas.
func locate(g *Gazetteer) {
	grid := make(map[cell][]*Area)
	for _, a := range g.Areas {
		lo, hi := cellOf(a.Bound.Min[0], a.Bound.Min[1]), cellOf(a.Bound.Max[0], a.Bound.Max[1])
		for x := lo.x; x <= hi.x; x++ {
			for y := lo.y; y <= hi.y; y++ {
				grid[cell{x, y}] = append(grid[cell{x, y}], a)
			}
		}
	}
	size := make(map[*Area]float64, len(g.Areas))
	for _, a := range g.Areas {
		size[a] = planar.Area(a.Polygon)
	}

	for _, p := range g.Places {
		p.Areas = p.Areas[:0]
		for _, a := range grid[cellOf(p.Point[0], p.Point[1])] {
			if a.Bound.Contains(p.Point) && planar.MultiPolygonContains(a.Polygon, p.Point) {
				p.Areas = append(p.Areas, a)
			}
		}
		slices.SortFunc(p.Areas, func(a, b *Area) int { return cmp.Compare(size[a], size[b]) })

		if p.Polygon != nil {
			continue
		}
		k := Key(p.Name)
		for _, a := range p.Areas {
			if hasKey(a.Names, k) {
				p.Polygon, p.PolygonFrom = a.Polygon, a.ID
				break
			}
		}
	}
}
