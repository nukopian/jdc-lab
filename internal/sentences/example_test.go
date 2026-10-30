package sentences_test

import (
	"fmt"

	"github.com/nukopian/jdc-lab/internal/sentences"
)

// Split ends a sentence at a stop and a capital, but not after an
// abbreviation, nor before a lower-case word.
func ExampleSplit() {
	wall := "We're hiring. Use tools, e.g. Sage or Xero, and report to Mr. Patel. " +
		"Earn £45k+. Visit example. com to apply."
	for _, s := range sentences.Split(wall) {
		fmt.Println(s)
	}
	// Output:
	// We're hiring.
	// Use tools, e.g. Sage or Xero, and report to Mr. Patel.
	// Earn £45k+.
	// Visit example. com to apply.
}
