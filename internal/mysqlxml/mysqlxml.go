// Package mysqlxml reads job ads exported by MySQL's --xml option, and
// writes them as the lab's records, in JSON Lines.
//
// The export is a <resultset> of <row>s, each holding a <field name="..."> per
// column, with NULL written as xsi:nil="true". XML is awkward to work with, so
// the lab converts it once to JSON Lines: one ad per line, as a JSON object
// with the lab's own field names (see ads.Ad), which jq, Go and Python all read
// without help.
//
// The conversion keeps every row and every value as it was: duplicates, HTML
// and odd characters are for later steps to deal with. The one change is
// XML's own: parsing turns each CRLF in a value into LF.
package mysqlxml

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"

	"github.com/nukopian/jdc-lab/internal/ads"
)

// A Row is one row of the export: each column's value, or nil for NULL.
type Row map[string]*string

// A Reader reads the rows of an export, one at a time.
type Reader struct {
	d *xml.Decoder
}

// NewReader returns a Reader that reads an export from r.
func NewReader(r io.Reader) *Reader {
	return &Reader{d: xml.NewDecoder(r)}
}

// row is a <row> as it's written, for the decoder to fill.
type row struct {
	Fields []struct {
		Name *string `xml:"name,attr"`
		Nil  string  `xml:"http://www.w3.org/2001/XMLSchema-instance nil,attr"`
		Text string  `xml:",chardata"`
	} `xml:"field"`
}

// Read returns the next row, or io.EOF when there are no more.
func (r *Reader) Read() (Row, error) {
	for {
		token, err := r.d.Token()
		if err != nil {
			return nil, err // io.EOF at the end, as Read promises
		}
		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "row" {
			continue
		}
		var x row
		if err := r.d.DecodeElement(&x, &start); err != nil {
			return nil, err
		}
		out := make(Row, len(x.Fields))
		for _, f := range x.Fields {
			if f.Name == nil {
				return nil, errors.New("a <field> without a name: not a MySQL export")
			}
			if f.Nil == "true" {
				out[*f.Name] = nil
			} else {
				out[*f.Name] = &f.Text
			}
		}
		return out, nil
	}
}

// Ad renames the row's columns to the lab's names, dropping the rest. A
// missing column is an error, not a silent nil: it means the export isn't
// the one this was written for.
func (row Row) Ad() (ads.Ad, error) {
	var err error
	column := func(name string, nullable bool) *string {
		v, ok := row[name]
		switch {
		case !ok:
			err = errors.Join(err, fmt.Errorf("row lacks %s", name))
		case v == nil && !nullable:
			err = errors.Join(err, fmt.Errorf("%s is NULL", name))
		}
		return v
	}
	text := func(name string) string {
		if v := column(name, false); v != nil {
			return *v
		}
		return ""
	}
	ad := ads.Ad{
		ID:                text("ref"),
		Title:             text("jobtitle"),
		Location:          column("location", true),
		LocationCanonical: column("location_canonical", true),
		HTML:              text("description_formatted"),
	}
	return ad, err
}

// Convert reads an export from r and writes its ads to w as JSON Lines,
// returning how many it wrote.
func Convert(r io.Reader, w io.Writer) (int, error) {
	buf := bufio.NewWriter(w)
	enc := json.NewEncoder(buf)
	// The HTML stays readable: <p>, not \u003cp\u003e.
	enc.SetEscapeHTML(false)
	rows := NewReader(r)
	count := 0
	for {
		row, err := rows.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, err
		}
		ad, err := row.Ad()
		if err != nil {
			return count, fmt.Errorf("row %d: %w", count+1, err)
		}
		if err := enc.Encode(ad); err != nil { // Encode ends each with a newline
			return count, err
		}
		count++
	}
	return count, buf.Flush()
}
