package ads_test

import (
	"fmt"
	"log"
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
