// Package shape shows a job ad in each of the shapes the pipeline gives it,
// stage by stage, as text to read: the HTML as posted, its blocks with
// their cues, the blocks with its walls split into sentences, the role
// headings gives each, its sections, and its chunks.
// Reading one ad down the stages shows where its shape changes, and which
// stage to look at when it goes wrong.
//
// It runs the same code the pipeline does, so it shows what the code does
// now, not what a run saved.
package shape

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/nukopian/jdc-lab/internal/ads"
	"github.com/nukopian/jdc-lab/internal/blocks"
	"github.com/nukopian/jdc-lab/internal/chunk"
	"github.com/nukopian/jdc-lab/internal/headings"
	"github.com/nukopian/jdc-lab/internal/sections"
	"github.com/nukopian/jdc-lab/internal/walls"
)

// Stages are the stages Write shows, in the pipeline's order.
var Stages = []string{"posted", "blocks", "walls", "headings", "sections", "chunks"}

// Write writes the ad at each of the stages named, or at all of them if none
// is, in the pipeline's order, each under its name. boilerplate says
// whether a text is boilerplate, as a run's verdicts do, for the chunks.
func Write(w io.Writer, a ads.Ad, boilerplate func(string) bool, stages ...string) error {
	for _, s := range stages {
		if !slices.Contains(Stages, s) {
			return fmt.Errorf("shape: no stage %q: the stages are %s", s, strings.Join(Stages, ", "))
		}
	}
	show := func(s string) bool { return len(stages) == 0 || slices.Contains(stages, s) }

	bs := blocks.Split(a.HTML)
	ws := walls.Split(bs)
	parts := headings.Read(ws, a.Title, a.Locations()...)
	ss := sections.Divide(parts)

	var b strings.Builder
	fmt.Fprintf(&b, "%s · %s\n", a.Title, a.ID)
	if show("posted") {
		fmt.Fprintf(&b, "\n## posted: the HTML as stored\n\n%s\n", strings.TrimRight(a.HTML, "\n"))
	}
	if show("blocks") {
		fmt.Fprintf(&b, "\n## blocks: %d, as a reader sees them, with their cues\n\n", len(bs))
		writeBlocks(&b, bs)
	}
	if show("walls") {
		n := 0
		for _, bl := range bs {
			if walls.Is(bl) {
				n++
			}
		}
		fmt.Fprintf(&b, "\n## walls: %d over %d characters, split into sentences, a line each\n\n", n, walls.MinChars)
		if n > 0 {
			writeBlocks(&b, ws)
		}
	}
	if show("headings") {
		fmt.Fprintf(&b, "\n## headings: %d parts, each with its role\n\n", len(parts))
		for i, p := range parts {
			item := " "
			if p.Item {
				item = "-"
			}
			fmt.Fprintf(&b, "%3d %-8s %s %s%s\n", i+1, p.Role, item, p.Text, keyed(p))
		}
	}
	if show("sections") {
		fmt.Fprintf(&b, "\n## sections: %d, each a heading and what follows it\n", len(ss))
		for i, s := range ss {
			heading := "no heading: what comes before the first"
			if s.Heading != "" {
				heading = fmt.Sprintf("%q", s.Heading)
			}
			if s.ByName {
				heading += ", by name"
			}
			fmt.Fprintf(&b, "\n[%d] %s · %s\n", i+1, s.Kind, heading)
			for _, p := range s.Parts {
				switch {
				case p.Role == headings.Field:
					fmt.Fprintf(&b, "    field %s\n", p.Text)
				case p.Item:
					fmt.Fprintf(&b, "    | - %s\n", p.Text)
				default:
					fmt.Fprintf(&b, "    | %s\n", p.Text)
				}
			}
		}
	}
	if _, err := io.WriteString(w, b.String()); err != nil {
		return err
	}
	if show("chunks") {
		if _, err := io.WriteString(w, "\n## chunks: what's embedded, kept as metadata, or left out\n\n"); err != nil {
			return err
		}
		c := chunk.AdOf(a.ID, a.Title, ss, boilerplate)
		return c.WriteText(w)
	}
	_, err := io.WriteString(w, "\n")
	return err
}

// writeBlocks writes blocks numbered, a list's items under it.
func writeBlocks(b *strings.Builder, bs []blocks.Block) {
	for i, bl := range bs {
		if bl.Kind == blocks.List {
			fmt.Fprintf(b, "%3d list of %d\n", i+1, len(bl.Items))
			for _, it := range bl.Items {
				fmt.Fprintf(b, "      - %s%s\n", it.Text, cues(it.Cues))
			}
			continue
		}
		fmt.Fprintf(b, "%3d %-9s %s%s\n", i+1, bl.Kind, bl.Text, cues(bl.Cues))
	}
}

// cues writes a block's cues after its text, in brackets: "[bold, gap]".
func cues(c blocks.Cues) string {
	var cs []string
	if c.Heading {
		cs = append(cs, "heading tag")
	}
	if c.Bold {
		cs = append(cs, "bold")
	}
	if c.BoldLead != "" {
		cs = append(cs, fmt.Sprintf("bold lead %q", c.BoldLead))
	}
	if c.Added != "" {
		cs = append(cs, fmt.Sprintf("label added %q", c.Added))
	}
	if c.Bullet != "" {
		cs = append(cs, fmt.Sprintf("bullet %q", c.Bullet))
	}
	if c.Gap {
		cs = append(cs, "gap")
	}
	if c.RunIn {
		cs = append(cs, "run in")
	}
	if len(cs) == 0 {
		return ""
	}
	return "  [" + strings.Join(cs, ", ") + "]"
}

// keyed writes a part's key and value after its text, where it has them.
func keyed(p headings.Part) string {
	var s string
	if p.Key != "" {
		s += fmt.Sprintf("  key %q", p.Key)
	}
	if p.Value != "" {
		s += fmt.Sprintf("  value %q", p.Value)
	}
	return s
}
