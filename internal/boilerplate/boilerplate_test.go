package boilerplate

import (
	"testing"

	"github.com/nukopian/jdc-lab/internal/sections"
)

const (
	agency   = "A dedicated consultant will be by your side at every step of the process."
	duties   = "Prepare learning materials and assist with engaging classroom activities every day."
	location = "This is a fantastic opportunity to join a busy team based in central Leeds."
)

func TestSimilar(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"support worker", "residential support worker", true},                 // one in the other
		{"paramedic or nurse part time", "paramedic or nurse full time", true}, // most words shared
		{"crane operator", "forklift driver", false},
		{"teaching assistant", "cover supervisor", false},
		{"sen teacher", "teacher", true},
	}
	for _, tt := range tests {
		if got := similar(tt.a, tt.b); got != tt.want {
			t.Errorf("similar(%q, %q) = %t, want %t", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestIndex(t *testing.T) {
	var ix Index // the zero value is ready to use
	ix.Add("Crane Operator", sections.Company, agency, location, "Pension")
	ix.Add("Forklift Driver", sections.Company, agency, "Pension")
	ix.Add("Teaching Assistant", sections.Responsibilities, duties, location)
	ix.Add("SEN Teaching Assistant", sections.Responsibilities, duties, location)
	ix.Add("Teaching Assistant", sections.Summary, location) // the same title again adds nothing

	tests := []struct {
		text   string
		groups int
	}{
		{agency, 2},    // two different jobs
		{duties, 1},    // two titles, but one job
		{location, 2},  // three titles, two jobs
		{"Pension", 0}, // too short to count
		{"Never seen in any ad at all, not even once, anywhere.", 0},
	}
	for _, tt := range tests {
		if got := ix.Groups(tt.text); got != tt.groups {
			t.Errorf("Groups(%q) = %d, want %d", tt.text, got, tt.groups)
		}
	}
	if !ix.Repeated(agency) || ix.Repeated(duties) {
		t.Errorf("Repeated: agency %t, duties %t; want true, false", ix.Repeated(agency), ix.Repeated(duties))
	}
}

func TestIndexCountsAgainAfterAdd(t *testing.T) {
	var ix Index
	ix.Add("Crane Operator", sections.Company, agency)
	if ix.Repeated(agency) {
		t.Fatal("repeated after one ad")
	}
	ix.Add("Forklift Driver", sections.Company, agency)
	if !ix.Repeated(agency) {
		t.Error("not repeated after a second job's ad")
	}
}

func TestBoilerplate(t *testing.T) {
	// Each text is in three ads for different jobs, in these kinds of
	// section; the verdict is the majority's, a tie going to template.
	tests := []struct {
		kinds       []sections.Kind
		boilerplate bool
	}{
		{[]sections.Kind{sections.Company, sections.Benefits, sections.Unmapped}, true},
		{[]sections.Kind{sections.Company, sections.Company, sections.Requirements}, true}, // misfiled once
		{[]sections.Kind{sections.Responsibilities, sections.Requirements, sections.Summary}, false},
		{[]sections.Kind{sections.Responsibilities, sections.Responsibilities, sections.Desirable}, false},
	}
	jobs := []string{"Crane Operator", "Teaching Assistant", "Chef"}
	for _, tt := range tests {
		var ix Index
		for i, kind := range tt.kinds {
			ix.Add(jobs[i], kind, agency)
		}
		if got := ix.Boilerplate(agency); got != tt.boilerplate {
			t.Errorf("in %v: Boilerplate = %t, want %t", tt.kinds, got, tt.boilerplate)
		}
		if got := ix.Template(agency); got == tt.boilerplate {
			t.Errorf("in %v: Template = %t, want %t", tt.kinds, got, !tt.boilerplate)
		}
	}
	var ix Index
	ix.Add("Chef", sections.Company, agency)
	if ix.Boilerplate(agency) || ix.Template(agency) {
		t.Error("a text in one ad is neither")
	}
}

func TestWidespreadIsBoilerplate(t *testing.T) {
	const teamwork = "Able to work effectively as part of a team and on your own initiative."
	var ix Index
	jobs := []string{"Chef", "Crane Operator", "Teaching Assistant", "Nurse", "Accountant"}
	for i, job := range jobs {
		ix.Add(job, sections.Requirements, teamwork)
		if got, want := ix.Boilerplate(teamwork), i+1 >= Widespread; got != want {
			t.Errorf("in %d jobs' requirements: Boilerplate = %t, want %t", i+1, got, want)
		}
	}
}
