package ads_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nukopian/jdc-lab/internal/ads"
)

// All reads ads.jsonl one ad at a time, for a range loop.
func ExampleAll() {
	jsonl := strings.NewReader(`{"id":"1","title":"Driver","location":null,"location_canonical":"Leeds","html":"<p>Drive.</p>"}
{"id":"2","title":"Nurse","location":"York","location_canonical":"York","html":"<p>Care.</p>"}
`)
	for ad, err := range ads.All(jsonl) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(ad.ID, ad.Title)
	}
	// Output:
	// 1 Driver
	// 2 Nurse
}

// File reads a file of ads, such as data/ingestion/ads.jsonl, and closes it
// at the end.
func ExampleFile() {
	path := filepath.Join(os.TempDir(), "jdc-lab-example-ads.jsonl")
	defer os.Remove(path)
	os.WriteFile(path, []byte(`{"id":"1","title":"Driver","location":null,"location_canonical":null,"html":""}`+"\n"), 0o644)
	for ad, err := range ads.File(path) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(ad.ID, ad.Title)
	}
	// Output:
	// 1 Driver
}

// Distinct leaves out an ad posted again, with the same HTML.
func ExampleDistinct() {
	var d ads.Distinct
	for _, a := range []ads.Ad{
		{ID: "1", HTML: "<p>Drive.</p>"},
		{ID: "2", HTML: "<p>Care.</p>"},
		{ID: "3", HTML: "<p>Drive.</p>"},
	} {
		if d.First(a) {
			fmt.Println(a.ID)
		}
	}
	fmt.Println(d.Duplicates, "duplicate")
	// Output:
	// 1
	// 2
	// 1 duplicate
}
