package grammars_test

import (
	"fmt"

	"github.com/nukopian/grammar"

	"github.com/nukopian/jdc-lab/internal/grammars"
)

// A grammar reads a fact's values from its captures; one of its rules,
// named with grammar.Rule, tells whether a line holds a part of one.
func Example() {
	m := grammars.Schedule.Find("Mon-Fri 07:00-15:30, overtime available")
	fmt.Println(m.At("~").Str(), "| from", m.At("days.first.~").Str(), "to", m.At("days.last.~").Str())

	clock := grammars.Schedule.Find("Start at 7:30am", grammar.Rule("clock"))
	fmt.Println(clock.At("~").Str())
	// Output:
	// Mon-Fri 07:00-15:30 | from Mon to Fri
	// 7:30am
}
