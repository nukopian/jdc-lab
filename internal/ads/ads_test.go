package ads

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
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

func TestAllSaysWhichAd(t *testing.T) {
	const lines = `{"id":"1"}
{"id":
`
	for _, err := range All(strings.NewReader(lines)) {
		if err != nil && !strings.HasPrefix(err.Error(), "ad 2: ") {
			t.Errorf("got %q, want it to start with \"ad 2: \"", err)
		}
	}
}

func TestFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ads.jsonl")
	if err := os.WriteFile(path, []byte(`{"id":"1"}`+"\n"+`{"id":`), 0o644); err != nil {
		t.Fatal(err)
	}
	var ids []string
	var errs []error
	for ad, err := range File(path) {
		if err != nil {
			errs = append(errs, err)
			continue
		}
		ids = append(ids, ad.ID)
	}
	if len(ids) != 1 || len(errs) != 1 || !strings.HasPrefix(errs[0].Error(), path+": ad 2: ") {
		t.Errorf("got ids %v and errors %v", ids, errs)
	}
	for _, err := range File(filepath.Join(t.TempDir(), "none.jsonl")) {
		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("a missing file: got %v", err)
		}
	}
}

func TestDistinct(t *testing.T) {
	var d Distinct // the zero value is ready to use
	got := []bool{
		d.First(Ad{ID: "1", HTML: "<p>A</p>"}),
		d.First(Ad{ID: "2", HTML: "<p>B</p>"}),
		d.First(Ad{ID: "3", HTML: "<p>A</p>"}),
	}
	if !slices.Equal(got, []bool{true, true, false}) || d.Duplicates != 1 {
		t.Errorf("got %v and %d duplicates; want [true true false] and 1", got, d.Duplicates)
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

func TestFindAll(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ads.jsonl")
	// The last line is broken: FindAll stops reading once it has every ad,
	// so only a search that needs it reaches it.
	lines := `{"id":"1","title":"Driver","location":null,"location_canonical":null,"html":"<p>Drive.</p>"}
{"id":"2","title":"Chef","location":null,"location_canonical":null,"html":"<p>Cook.</p>"}
{"id":"1","title":"Driver again","location":null,"location_canonical":null,"html":"<p>Drive.</p>"}
{"id":
`
	if err := os.WriteFile(path, []byte(lines), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := FindAll(path, []string{"2", "1", "2"})
	if err != nil || len(got) != 2 || got["1"].Title != "Driver" || got["2"].Title != "Chef" {
		t.Errorf(`FindAll(path, 2, 1, 2) = %v, %v; want the first Driver and the Chef`, got, err)
	}
	// Found in the wash-up: asking for no ads reads nothing, so the broken
	// line isn't reached.
	if got, err := FindAll(path, nil); err != nil || len(got) != 0 {
		t.Errorf("FindAll(path, nil) = %v, %v; want nothing, read", got, err)
	}
	_, err = FindAll(path, []string{"1", "4", "3"})
	if err == nil || errors.Is(err, ErrNotFound) {
		t.Errorf("FindAll reading the broken line: error = %v, want the line's", err)
	}
	path = filepath.Join(t.TempDir(), "ads.jsonl")
	if err := os.WriteFile(path, []byte(strings.Join(strings.Split(lines, "\n")[:3], "\n")), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := FindAll(filepath.Join(t.TempDir(), "none.jsonl"), []string{"1"}); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("FindAll in a missing file: error = %v, want one that it doesn't exist", err)
	}
	got, err = FindAll(path, []string{"1", "4", "3", "4"})
	var nf *NotFoundError
	if !errors.Is(err, ErrNotFound) || !errors.As(err, &nf) || !slices.Equal(nf.IDs, []string{"4", "3"}) ||
		len(got) != 1 || got["1"].Title != "Driver" {
		t.Errorf("FindAll(path, 1, 4, 3, 4) = %v, %v; want the Driver, and ErrNotFound naming 4, 3", got, err)
	}
}
