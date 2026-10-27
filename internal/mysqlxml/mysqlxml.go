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
//
// That includes control characters such as U+0002. XML 1.0 forbids them,
// even escaped, but MySQL writes them as they are, so Go's decoder would
// refuse the export. The Reader stands each in for a character in Unicode's
// private use area while the XML is parsed, then swaps it back; an export
// that already holds one of those stand-ins is refused, as it couldn't be
// read back as it was.
package mysqlxml

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

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
	return &Reader{d: xml.NewDecoder(&controls{r: r})}
}

// standIn is where the stand-ins start: control character c becomes
// standIn+c, from U+E000 to U+E01F.
const standIn = 0xE000

// isControl reports whether XML 1.0 forbids byte b, which is then a whole
// character: in UTF-8, bytes below 0x80 are never part of a longer one.
func isControl(b byte) bool {
	return b < 0x20 && b != '\t' && b != '\n' && b != '\r'
}

// controls is an io.Reader that replaces each control character read from
// r with its stand-in, and fails on a stand-in already there.
type controls struct {
	r    io.Reader
	buf  [32 << 10]byte
	held int    // bytes at the start of buf that may begin a stand-in
	out  []byte // what's ready to be read
	err  error
}

func (c *controls) Read(p []byte) (int, error) {
	for len(c.out) == 0 {
		if c.err != nil {
			return 0, c.err
		}
		n, err := c.r.Read(c.buf[c.held:])
		c.err = err
		data := c.buf[:c.held+n]
		// A stand-in is EE 80 then 80 to 9F, so one that may be split
		// across reads waits for the next.
		keep := 0
		if err == nil {
			switch {
			case len(data) >= 2 && data[len(data)-2] == 0xEE && data[len(data)-1] == 0x80:
				keep = 2
			case len(data) >= 1 && data[len(data)-1] == 0xEE:
				keep = 1
			}
		}
		c.out = c.out[:0]
		for i, b := range data[:len(data)-keep] {
			if b == 0xEE && i+2 < len(data) && data[i+1] == 0x80 && data[i+2] >= 0x80 && data[i+2] <= 0x9F {
				c.out, c.err = nil, fmt.Errorf("the export holds U+%04X, which the reader uses for a control character", standIn+rune(data[i+2]-0x80))
				return 0, c.err
			}
			if isControl(b) {
				c.out = utf8.AppendRune(c.out, standIn+rune(b))
			} else {
				c.out = append(c.out, b)
			}
		}
		c.held = copy(c.buf[:], data[len(data)-keep:])
	}
	n := copy(p, c.out)
	c.out = c.out[n:]
	return n, nil
}

// restore swaps the stand-ins in s back to the control characters they
// stand for.
func restore(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= standIn && r < standIn+0x20 {
			return r - standIn
		}
		return r
	}, s)
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
				text := restore(f.Text)
				out[*f.Name] = &text
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
