package gazetteer_test

import (
	"fmt"

	"github.com/nukopian/jdc-lab/internal/gazetteer"
)

func ExampleKey() {
	fmt.Println(gazetteer.Key("Billinge, St. Helens"))
	fmt.Println(gazetteer.Key("Bath & North East Somerset"))
	// Output:
	// billinge st helens
	// bath and north east somerset
}
