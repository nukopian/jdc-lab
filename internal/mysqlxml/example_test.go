package mysqlxml_test

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/nukopian/jdc-lab/internal/mysqlxml"
)

const export = `<resultset xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <row>
	<field name="ref">1</field>
	<field name="location" xsi:nil="true" />
	<field name="location_canonical">Leeds</field>
	<field name="jobtitle">Driver</field>
	<field name="description_formatted">&lt;p&gt;Drive.&lt;/p&gt;</field>
  </row>
</resultset>`

// Convert writes each row as a line of JSON, with the lab's field names.
func ExampleConvert() {
	if _, err := mysqlxml.Convert(strings.NewReader(export), os.Stdout); err != nil {
		log.Fatal(err)
	}
	// Output:
	// {"id":"1","title":"Driver","location":null,"location_canonical":"Leeds","html":"<p>Drive.</p>"}
}

// A Reader gives one row at a time, and io.EOF after the last.
func ExampleReader_Read() {
	r := mysqlxml.NewReader(strings.NewReader(export))
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(*row["jobtitle"], row["location"] == nil)
	}
	// Output:
	// Driver true
}

// Ad renames a row's columns to the lab's names.
func ExampleRow_Ad() {
	row, _ := mysqlxml.NewReader(strings.NewReader(export)).Read()
	ad, err := row.Ad()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ad.ID, ad.Title, ad.HTML)
	// Output:
	// 1 Driver <p>Drive.</p>
}
