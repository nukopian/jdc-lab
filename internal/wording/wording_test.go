package wording

import (
	"testing"

	"github.com/nukopian/jdc-lab/internal/sections"
)

func TestOf(t *testing.T) {
	tests := []struct {
		line   string
		kind   sections.Kind
		phrase string
	}{
		{"Driving License essential", sections.Requirements, "Driving License"},
		{"- Minimum 5 years' experience in a Project Management role.", sections.Requirements, "Minimum"},
		{"Experience of working in a care setting", sections.Requirements, "Experience of"},
		{"You'll have a full UK licence", sections.Requirements, "You'll have"},
		{"Lead lifecycle and M&E projects from concept to completion.", sections.Responsibilities, "Lead"},
		{"Conduct site surveys and develop scopes of work.", sections.Responsibilities, "Conduct"},
		{"Responsible for a team of five", sections.Responsibilities, "Responsible for"},
		{"Free parking", sections.Benefits, "Free parking"},
		{"33 days holiday including bank holidays", sections.Benefits, "33 days holiday"},
		{"You'll get a company van", sections.Benefits, "You'll get"},
		// The longer phrase wins: an opportunity is a benefit.
		{"You'll have the opportunity to develop your parts knowledge", sections.Benefits, "You'll have the opportunity"},
		// Nothing: a pitch, a question, the words later in the line.
		{"Join our team as a Project Manager", "", ""},
		{"Are you an experienced Project Manager?", "", ""},
		{"We offer free parking", "", ""},
		{"Leading the way in care since 1990", "", ""},
	}
	for _, tt := range tests {
		kind, phrase, ok := Of(tt.line)
		if kind != tt.kind || phrase != tt.phrase || ok != (tt.kind != "") {
			t.Errorf("Of(%q) = %q, %q, %t; want %q, %q", tt.line, kind, phrase, ok, tt.kind, tt.phrase)
		}
	}
}
