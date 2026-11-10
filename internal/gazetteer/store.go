package gazetteer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"

	_ "modernc.org/sqlite" // registers the "sqlite" driver, pure Go, no cgo
)

// schema is the gazetteer's tables. Geometry is GeoJSON text, which
// SQLite keeps as it is and anything can read; points are plain columns,
// so they can be queried without parsing.
const schema = `
CREATE TABLE places (
	id           TEXT PRIMARY KEY,  -- node/123, way/456, relation/789
	name         TEXT NOT NULL,
	kind         TEXT NOT NULL,     -- city, town, village, hamlet, ...
	lon          REAL NOT NULL,
	lat          REAL NOT NULL,
	population   INTEGER,           -- NULL if not tagged
	polygon      TEXT,              -- GeoJSON, when the outline is the place's own
	polygon_from TEXT               -- the place or area the outline came from
);
CREATE TABLE areas (
	id          TEXT PRIMARY KEY,   -- relation/789
	name        TEXT NOT NULL,
	kind        TEXT NOT NULL,      -- administrative or ceremonial
	admin_level INTEGER,            -- 5 (region) to 10 (parish); NULL if ceremonial
	min_lon REAL, min_lat REAL, max_lon REAL, max_lat REAL,
	polygon     TEXT NOT NULL       -- GeoJSON
);
-- Every name a place or area goes by, with the key Match compares.
CREATE TABLE names (
	id   TEXT NOT NULL,             -- a place's or an area's
	name TEXT NOT NULL,
	key  TEXT NOT NULL,
	PRIMARY KEY (id, name)
);
CREATE INDEX names_key ON names (key);
-- The areas each place's point lies in.
CREATE TABLE place_areas (
	place_id TEXT NOT NULL,
	area_id  TEXT NOT NULL,
	PRIMARY KEY (place_id, area_id)
);
CREATE INDEX place_areas_area ON place_areas (area_id);
`

// Save writes g to a new SQLite database at path, replacing any there.
func (g *Gazetteer) Save(ctx context.Context, path string) (err error) {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("gazetteer: creating %s: %w", path, err)
	}

	// One transaction for the lot: SQLite commits each statement on its
	// own otherwise, which is thousands of times slower.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // does nothing once committed

	insert := func(query string, rows func(stmt *sql.Stmt) error) error {
		stmt, err := tx.PrepareContext(ctx, query)
		if err != nil {
			return err
		}
		defer stmt.Close()
		return rows(stmt)
	}
	err = insert(`INSERT INTO areas VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, func(stmt *sql.Stmt) error {
		for _, a := range g.Areas {
			polygon, err := geoJSON(a.Polygon)
			if err != nil {
				return err
			}
			if _, err := stmt.ExecContext(ctx, a.ID, a.Name, a.Kind, nullZero(a.Level),
				a.Bound.Min[0], a.Bound.Min[1], a.Bound.Max[0], a.Bound.Max[1], polygon); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	areaIDs := make(map[string]bool, len(g.Areas))
	for _, a := range g.Areas {
		areaIDs[a.ID] = true
	}
	err = insert(`INSERT INTO places VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, func(stmt *sql.Stmt) error {
		for _, p := range g.Places {
			// An outline borrowed from an area is kept once, in areas.
			var polygon any
			if p.Polygon != nil && p.PolygonFrom != "" && !areaIDs[p.PolygonFrom] {
				s, err := geoJSON(p.Polygon)
				if err != nil {
					return err
				}
				polygon = s
			}
			if _, err := stmt.ExecContext(ctx, p.ID, p.Name, p.Kind, p.Point[0], p.Point[1],
				nullZero(p.Population), polygon, nullEmpty(p.PolygonFrom)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = insert(`INSERT OR IGNORE INTO names VALUES (?, ?, ?)`, func(stmt *sql.Stmt) error {
		write := func(id string, names []string) error {
			for _, n := range names {
				if _, err := stmt.ExecContext(ctx, id, n, Key(n)); err != nil {
					return err
				}
			}
			return nil
		}
		for _, p := range g.Places {
			if err := write(p.ID, p.Names); err != nil {
				return err
			}
		}
		for _, a := range g.Areas {
			if err := write(a.ID, a.Names); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = insert(`INSERT INTO place_areas VALUES (?, ?)`, func(stmt *sql.Stmt) error {
		for _, p := range g.Places {
			for _, a := range p.Areas {
				if _, err := stmt.ExecContext(ctx, p.ID, a.ID); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return tx.Commit()
}

// Load reads a gazetteer Save wrote.
func Load(ctx context.Context, path string) (_ *Gazetteer, err error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err // sql.Open would make an empty database instead
	}
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, db.Close()) }()

	g := &Gazetteer{}
	areas := make(map[string]*Area)
	err = each(ctx, db, `SELECT id, name, kind, coalesce(admin_level, 0), polygon FROM areas ORDER BY id`, func(rows *sql.Rows) error {
		a := &Area{}
		var polygon string
		if err := rows.Scan(&a.ID, &a.Name, &a.Kind, &a.Level, &polygon); err != nil {
			return err
		}
		mp, err := parseGeoJSON(polygon)
		if err != nil {
			return fmt.Errorf("gazetteer: %s: %w", a.ID, err)
		}
		a.Polygon, a.Bound = mp, mp.Bound()
		g.Areas = append(g.Areas, a)
		areas[a.ID] = a
		return nil
	})
	if err != nil {
		return nil, err
	}

	places := make(map[string]*Place)
	err = each(ctx, db, `SELECT id, name, kind, lon, lat, coalesce(population, 0), polygon, coalesce(polygon_from, '') FROM places ORDER BY id`, func(rows *sql.Rows) error {
		p := &Place{}
		var polygon sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &p.Kind, &p.Point[0], &p.Point[1], &p.Population, &polygon, &p.PolygonFrom); err != nil {
			return err
		}
		switch {
		case polygon.Valid:
			mp, err := parseGeoJSON(polygon.String)
			if err != nil {
				return fmt.Errorf("gazetteer: %s: %w", p.ID, err)
			}
			p.Polygon = mp
		case areas[p.PolygonFrom] != nil:
			p.Polygon = areas[p.PolygonFrom].Polygon
		}
		g.Places = append(g.Places, p)
		places[p.ID] = p
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = each(ctx, db, `SELECT id, name FROM names ORDER BY rowid`, func(rows *sql.Rows) error {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return err
		}
		if p := places[id]; p != nil {
			p.Names = append(p.Names, name)
		} else if a := areas[id]; a != nil {
			a.Names = append(a.Names, name)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// Smallest bound first, near enough the order locate leaves them in.
	err = each(ctx, db, `
		SELECT pa.place_id, pa.area_id FROM place_areas pa JOIN areas a ON a.id = pa.area_id
		ORDER BY pa.place_id, (a.max_lon - a.min_lon) * (a.max_lat - a.min_lat)`, func(rows *sql.Rows) error {
		var pid, aid string
		if err := rows.Scan(&pid, &aid); err != nil {
			return err
		}
		if p, a := places[pid], areas[aid]; p != nil && a != nil {
			p.Areas = append(p.Areas, a)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	g.index()
	return g, nil
}

// each runs query and calls f for each row.
func each(ctx context.Context, db *sql.DB, query string, f func(*sql.Rows) error) error {
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := f(rows); err != nil {
			return err
		}
	}
	return rows.Err()
}

func geoJSON(mp orb.MultiPolygon) (string, error) {
	var g orb.Geometry = mp
	if len(mp) == 1 {
		g = mp[0]
	}
	b, err := geojson.NewGeometry(g).MarshalJSON()
	return string(b), err
}

func parseGeoJSON(s string) (orb.MultiPolygon, error) {
	g, err := geojson.UnmarshalGeometry([]byte(s))
	if err != nil {
		return nil, err
	}
	mp := multi(g.Geometry())
	if mp == nil {
		return nil, fmt.Errorf("not a polygon: %s", g.Type)
	}
	return mp, nil
}

// nullZero and nullEmpty store a missing value as NULL, not 0 or "".
func nullZero(n int) any {
	if n == 0 {
		return nil
	}
	return n
}

func nullEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
