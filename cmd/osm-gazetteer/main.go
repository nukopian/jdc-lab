// Command osm-gazetteer builds a gazetteer of named places and the areas
// that hold them from an OpenStreetMap extract (see internal/gazetteer),
// and saves it as SQLite.
//
//	osm-gazetteer [-out FILE] EXTRACT.osm.pbf
//
// From the lab's folder, with Geofabrik's extract for the United Kingdom
// in local/osm:
//
//	go run ./cmd/osm-gazetteer local/osm/united-kingdom-latest.osm.pbf
//
// It writes local/osm/gazetteer.sqlite unless told otherwise.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/nukopian/jdc-lab/internal/gazetteer"
)

func main() {
	log.SetFlags(0)
	out := flag.String("out", "local/osm/gazetteer.sqlite", "the SQLite file to write")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("usage: osm-gazetteer [-out FILE] EXTRACT.osm.pbf")
	}
	// Ctrl-C cancels the context, which stops the scanner between blocks.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	start := time.Now()
	g, err := gazetteer.Build(ctx, flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	withOutline := 0
	for _, p := range g.Places {
		if p.Polygon != nil {
			withOutline++
		}
	}
	log.Printf("osm-gazetteer: %d places (%d with an outline) and %d areas in %s",
		len(g.Places), withOutline, len(g.Areas), time.Since(start).Round(time.Second))

	if err := g.Save(ctx, *out); err != nil {
		log.Fatal(err)
	}
	log.Printf("osm-gazetteer: wrote %s in %s", *out, time.Since(start).Round(time.Second))
}
