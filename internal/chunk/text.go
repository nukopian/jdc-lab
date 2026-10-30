package chunk

import (
	"fmt"
	"io"
	"strings"
)

// WriteText writes the ad's chunks as text to read: its title and totals,
// then each chunk with its kind, what's done with it, its size and what
// placed it, its lines, and its fields.
func (a Ad) WriteText(w io.Writer) error {
	share := func(n int) int {
		if a.Tokens == 0 {
			return 0
		}
		return 100 * n / a.Tokens
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s · %s\n", a.Title, a.ID)
	fmt.Fprintf(&b, "about %d tokens in all; %d embedded (%d%%), %d unclassified (%d%%)\n",
		a.Tokens, a.Embedded, share(a.Embedded), a.Unclassified, share(a.Unclassified))
	for i, ch := range a.Chunks {
		fmt.Fprintf(&b, "\n[%d] %s · %s · about %d tokens\n    placed by %s\n", i+1, ch.Kind, ch.Use, ch.Tokens, ch.PlacedBy)
		if ch.Text != "" {
			for _, line := range strings.Split(ch.Text, "\n") {
				fmt.Fprintf(&b, "    | %s\n", line)
			}
		}
		for _, f := range ch.Fields {
			fmt.Fprintf(&b, "    field %s: %s\n", f.Label, f.Value)
		}
	}
	fmt.Fprintln(&b)
	_, err := io.WriteString(w, b.String())
	return err
}
