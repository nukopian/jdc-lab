package chunk

import (
	"strings"
	"testing"
)

func TestWriteText(t *testing.T) {
	a := Ad{ID: "1", Title: "Driver", Tokens: 20, Embedded: 15, Unclassified: 0, Chunks: []Chunk{
		{Kind: "summary", Use: Embed, Tokens: 15, PlacedBy: "nothing before it", Text: "We're hiring.\nDrive."},
		{Kind: "details", Use: Metadata, Tokens: 5, PlacedBy: "lines of facts", Text: "Pay: £30,000",
			Fields: []Field{{Label: "Pay", Value: "£30,000"}}},
	}}
	var b strings.Builder
	if err := a.WriteText(&b); err != nil {
		t.Fatal(err)
	}
	want := `Driver · 1
about 20 tokens in all; 15 embedded (75%), 0 unclassified (0%)

[1] summary · embed · about 15 tokens
    placed by nothing before it
    | We're hiring.
    | Drive.

[2] details · metadata · about 5 tokens
    placed by lines of facts
    | Pay: £30,000
    field Pay: £30,000

`
	if got := b.String(); got != want {
		t.Errorf("WriteText:\n%s\nwant:\n%s", got, want)
	}
}
