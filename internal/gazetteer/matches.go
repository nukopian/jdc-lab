package gazetteer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
)

// A Tally is a name's match and how many times the name was seen.
type Tally struct {
	Match
	Count int
}

const matchesSchema = `
CREATE TABLE matches (
	name         TEXT PRIMARY KEY,  -- as given
	count        INTEGER NOT NULL,  -- times it was seen
	status       TEXT NOT NULL,     -- matched, chosen, out-of-area, area or unmatched
	candidates   INTEGER NOT NULL,  -- places of the name
	in_area      INTEGER NOT NULL,  -- of those, the ones in the area named
	id           TEXT,              -- the place or area taken
	osm_name     TEXT,
	kind         TEXT,              -- the place's kind, or the area's boundary
	lon          REAL,              -- the place's point; NULL for an area
	lat          REAL,
	polygon_from TEXT               -- where its outline is, in the gazetteer: NULL if it has none
);
`

// SaveMatches writes tallies to a new SQLite database at path, replacing
// any there. Outlines aren't copied: polygon_from names the place or area
// in the gazetteer that has it.
func SaveMatches(ctx context.Context, path string, tallies []Tally) (err error) {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	if _, err := db.ExecContext(ctx, matchesSchema); err != nil {
		return fmt.Errorf("gazetteer: creating %s: %w", path, err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO matches VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, t := range tallies {
		var id, name, kind, lon, lat, from any
		switch {
		case t.Place != nil:
			id, name, kind, lon, lat = t.Place.ID, t.Place.Name, t.Place.Kind, t.Place.Point[0], t.Place.Point[1]
			from = nullEmpty(t.Place.PolygonFrom)
		case t.Area != nil:
			id, name, kind, from = t.Area.ID, t.Area.Name, t.Area.Kind, t.Area.ID
		}
		if _, err := stmt.ExecContext(ctx, t.Name, t.Count, t.Status, t.Candidates, t.InArea,
			id, name, kind, lon, lat, from); err != nil {
			return err
		}
	}
	return tx.Commit()
}
