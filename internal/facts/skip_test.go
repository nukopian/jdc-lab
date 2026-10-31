package facts

import (
	"reflect"
	"testing"

	"github.com/nukopian/jdc-lab/internal/grammars"
)

// tricky are lines on the edges of grammars.MayHold: digits that aren't
// ASCII, letters that fold to ASCII, and each grammar's least.
var tricky = []string{
	"Salary: £30,000 per annum", "Pay Rate - 13.10", "Up to ٣٠ per hour",
	"40hrs pw", "37.5 hours per week", "NE8 1AQ",
	"Mon-Fri 07:00-15:30", "M - F", "Monday to Friday", "MON - FRI", "Wednesday to Sunday",
	"4 on 4 off", "8.30-3.30",
	"Permanent", "PERM", "Temp-to-perm", "Fixed-term contract", "FTC", "WFH", "Full-time",
	"Fully remote", "Hybrid working", "Office-based", "Site based",
	"WOR\u212A FROM HOME", "ſite based", "Work from home", "HYBRİD", "SİTE BASED",
	"We are a friendly team.", "",
}

// Skipping lines that can't hold a fact changes no fact.
func TestSkipLinesChangeNothing(t *testing.T) {
	lines := tricky
	if testing.Short() {
		t.Log("short: the made-up lines only")
	} else if b := sampleLinesOrNil(2000); b != nil {
		lines = append(lines, b...)
	}
	skipped := 0
	for _, l := range lines {
		on := Find(l)
		skipLines = false
		off := Find(l)
		skipLines = true
		if !reflect.DeepEqual(on, off) {
			t.Errorf("Find(%q)\nskipping: %v\n     not: %v", l, on, off)
		}
		if !grammars.MayHold(grammars.Pay, l) && !grammars.MayHold(grammars.Schedule, l) && !grammars.MayHold(grammars.Pattern, l) {
			skipped++
		}
	}
	if skipped == 0 {
		t.Error("no line skipped: the checks say yes to everything")
	}
	t.Logf("%d lines, %d skipped by every grammar", len(lines), skipped)
}
