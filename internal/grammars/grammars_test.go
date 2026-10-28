package grammars

import (
	"testing"

	"github.com/nukopian/grammar"
)

// TestRules checks the rules internal/headings uses to tell a line that
// holds a time, a pattern of work or a long figure.
func TestRules(t *testing.T) {
	tests := []struct {
		rule, text, want string // want is the first match's text, or "" for none
	}{
		{"clock", "Friday7:00am-3:00pm", "7:00am"},
		{"clock", "starting 745am", ""}, // not "45am", from inside a figure
		{"clock", "Shift 22:00 - 06:00", "22:00"},
		{"clock", "37.50 hours", ""},
		{"points", "Hours 8.30-3.30", "8.30-3.30"},
		{"span", "Monday to Friday 8-5", "8-5"},
		{"span", "2-3 days in the office", "2-3 days"},
		{"span", "Ages 5-11", "5-11"},
		{"span", "Ref 1234-5678", ""},
		{"days", "Mon to Thurs", "Mon to Thurs"},
		{"days", "Monmouth to Sunderland", ""},
		{"days", "M - F", "M - F"},
		{"rota", "4 days on, 4 days off", "4 days on, 4 days off"},
		{"rota", "4ON4OFF", "4ON4OFF"},
		{"shift", "Monday - Friday (Days)", "Days"},
		{"shift", "2-3 days", "3 days"},
		{"to", "2pm til 10pm", " til "},
	}
	for _, tt := range tests {
		got := Schedule.Find(tt.text, grammar.Rule(tt.rule)).At("~").Str()
		if got != tt.want {
			t.Errorf("%s in %q: got %q, want %q", tt.rule, tt.text, got, tt.want)
		}
	}
	for text, want := range map[string]string{"Ref 225682916": "225682916", "£30,000": "000", "40 hours": ""} {
		if got := Figure.Find(text).At("~").Str(); got != want {
			t.Errorf("Figure in %q: got %q, want %q", text, got, want)
		}
	}
}
