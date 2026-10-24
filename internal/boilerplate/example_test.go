package boilerplate_test

import (
	"fmt"

	"github.com/nukopian/jdc-lab/internal/boilerplate"
	"github.com/nukopian/jdc-lab/internal/sections"
)

// Add every ad's texts, by section, then ask about any text. Text repeated
// in ads for different jobs is boilerplate if it's mostly outside the
// sections that describe the job, and a template for the job if not.
func ExampleIndex() {
	const (
		agency = "We are a leading recruitment agency with offices across the UK and Ireland."
		duties = "Prepare learning materials and support pupils with their reading every day."
	)
	var ix boilerplate.Index
	ix.Add("Crane Operator", sections.Company, agency)
	ix.Add("Teaching Assistant", sections.Company, agency)
	ix.Add("Teaching Assistant", sections.Responsibilities, duties)
	ix.Add("Cover Supervisor", sections.Responsibilities, duties)
	ix.Add("SEN Teaching Assistant", sections.Responsibilities, duties)

	for _, text := range []string{agency, duties} {
		fmt.Printf("groups %d, boilerplate %t, template %t\n",
			ix.Groups(text), ix.Boilerplate(text), ix.Template(text))
	}
	// Output:
	// groups 2, boilerplate true, template false
	// groups 2, boilerplate false, template true
}

// Describes says which kinds of section describe the job itself.
func ExampleDescribes() {
	fmt.Println(boilerplate.Describes(sections.Requirements), boilerplate.Describes(sections.Company))
	// Output:
	// true false
}
