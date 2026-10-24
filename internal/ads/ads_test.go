package ads

import (
	"slices"
	"strings"
	"testing"
)

func TestAll(t *testing.T) {
	const lines = `{"id":"2","title":"Café Manager","location":"Leeds","location_canonical":null,"html":"<p>£30k</p>"}
{"id":"1","title":"Driver","location":null,"location_canonical":"","html":"<p>Drive.</p>"}
`
	var got []Ad
	for ad, err := range All(strings.NewReader(lines)) {
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, ad)
	}
	if len(got) != 2 {
		t.Fatalf("got %d ads, want 2", len(got))
	}
	first, second := got[0], got[1]
	if first.Title != "Café Manager" || *first.Location != "Leeds" || first.LocationCanonical != nil {
		t.Errorf("first: got %+v", first)
	}
	if second.Location != nil || *second.LocationCanonical != "" || second.HTML != "<p>Drive.</p>" {
		t.Errorf("second: got %+v", second)
	}
}

func TestLocations(t *testing.T) {
	leeds, empty := "Leeds", ""
	tests := []struct {
		ad   Ad
		want []string
	}{
		{Ad{Location: &leeds, LocationCanonical: &leeds}, []string{"Leeds", "Leeds"}},
		{Ad{LocationCanonical: &empty}, []string{""}},
		{Ad{}, nil},
	}
	for _, tt := range tests {
		if got := tt.ad.Locations(); !slices.Equal(got, tt.want) {
			t.Errorf("Locations() = %q, want %q", got, tt.want)
		}
	}
}

func TestAllStopsAtTheFirstError(t *testing.T) {
	const lines = `{"id":"1"}
{"id":
{"id":"3"}
`
	var ids []string
	var errs int
	for ad, err := range All(strings.NewReader(lines)) {
		if err != nil {
			errs++
			continue
		}
		ids = append(ids, ad.ID)
	}
	if len(ids) != 1 || ids[0] != "1" || errs != 1 {
		t.Errorf("got ids %v and %d errors; want [1] and 1 error", ids, errs)
	}
}

func TestAllStopsWhenTheLoopDoes(t *testing.T) {
	const lines = `{"id":"1"}
{"id":"2"}
`
	n := 0
	for range All(strings.NewReader(lines)) {
		n++
		break
	}
	if n != 1 {
		t.Errorf("the loop body ran %d times after break, want 1", n)
	}
}
