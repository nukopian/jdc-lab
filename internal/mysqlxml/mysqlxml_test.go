package mysqlxml

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/nukopian/jdc-lab/internal/ads"
)

// export is two ads as MySQL's --xml writes them: HTML escaped as entities,
// a NULL location, a CRLF inside a value, characters beyond ASCII, and a
// column the lab doesn't use.
const export = "<?xml version=\"1.0\"?>\n\n" +
	`<resultset statement="select ..." xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <row>
	<field name="ref">2</field>
	<field name="location">Leeds, West Yorkshire</field>
	<field name="location_canonical">Leeds</field>
	<field name="jobtitle">Café Manager</field>
	<field name="description_formatted">&lt;p&gt;&lt;strong&gt;The role&lt;/strong&gt;&lt;/p&gt;&lt;p&gt;£30,000 ✨&amp;nbsp;` + "\r\n" + `next&lt;/p&gt;</field>
	<field name="extra">ignored</field>
  </row>
  <row>
	<field name="ref">1</field>
	<field name="location" xsi:nil="true" />
	<field name="location_canonical"></field>
	<field name="jobtitle">Driver</field>
	<field name="description_formatted">&lt;p&gt;Drive.&lt;/p&gt;</field>
	<field name="extra">ignored</field>
  </row>
</resultset>
`

// readAll reads every row of text, failing the test on an error.
func readAll(t *testing.T, text string) []Row {
	t.Helper()
	r := NewReader(strings.NewReader(text))
	var rows []Row
	for {
		row, err := r.Read()
		if err == io.EOF {
			return rows
		}
		if err != nil {
			t.Fatal(err)
		}
		rows = append(rows, row)
	}
}

func TestReadEveryRowInOrder(t *testing.T) {
	rows := readAll(t, export)
	if len(rows) != 2 || *rows[0]["ref"] != "2" || *rows[1]["ref"] != "1" {
		t.Fatalf("got %d rows, want refs 2 then 1", len(rows))
	}
}

func TestReadValues(t *testing.T) {
	rows := readAll(t, export)
	first, second := rows[0], rows[1]
	html := *first["description_formatted"]
	tests := []struct {
		name string
		ok   bool
	}{
		{"unescapes the HTML", strings.HasPrefix(html, "<p><strong>The role</strong></p>")},
		{"keeps characters and HTML entities", strings.Contains(html, "£30,000 ✨&nbsp;")},
		// XML parsers must (XML 1.0, section 2.11), unless the export had
		// escaped the \r, which MySQL doesn't.
		{"turns CRLF into LF", strings.Contains(html, "\nnext") && !strings.Contains(html, "\r")},
		{"reads NULL as nil", second["location"] == nil},
		{"tells empty from NULL", second["location_canonical"] != nil && *second["location_canonical"] == ""},
	}
	for _, tt := range tests {
		if !tt.ok {
			t.Errorf("Read doesn't: %s", tt.name)
		}
	}
}

func TestReadRefusesAFieldWithoutAName(t *testing.T) {
	_, err := NewReader(strings.NewReader("<resultset><row><field>1</field></row></resultset>")).Read()
	if err == nil || !strings.Contains(err.Error(), "without a name") {
		t.Fatalf("got %v, want an error about a field without a name", err)
	}
}

func TestReadRefusesBrokenXML(t *testing.T) {
	_, err := NewReader(strings.NewReader("<resultset><row><field name=\"ref\">1</row>")).Read()
	if err == nil || errors.Is(err, io.EOF) {
		t.Fatalf("got %v, want a syntax error", err)
	}
}

func TestAd(t *testing.T) {
	ad, err := readAll(t, export)[0].Ad()
	if err != nil {
		t.Fatal(err)
	}
	if ad.ID != "2" || ad.Title != "Café Manager" || *ad.Location != "Leeds, West Yorkshire" {
		t.Errorf("got %+v", ad)
	}
}

func TestAdRefuses(t *testing.T) {
	s := func(v string) *string { return &v }
	full := func() Row {
		return Row{"ref": s("1"), "jobtitle": s("Driver"), "location": nil,
			"location_canonical": nil, "description_formatted": s("<p>Drive.</p>")}
	}
	tests := []struct {
		name   string
		change func(Row)
		want   string
	}{
		{"a missing column", func(r Row) { delete(r, "jobtitle") }, "row lacks jobtitle"},
		{"a NULL title", func(r Row) { r["jobtitle"] = nil }, "jobtitle is NULL"},
		{"a missing location", func(r Row) { delete(r, "location") }, "row lacks location"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := full()
			tt.change(row)
			if _, err := row.Ad(); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("got %v, want %q", err, tt.want)
			}
		})
	}
	if _, err := full().Ad(); err != nil {
		t.Errorf("a complete row: %v", err)
	}
}

func TestConvert(t *testing.T) {
	var out strings.Builder
	n, err := Convert(strings.NewReader(export), &out)
	if err != nil || n != 2 {
		t.Fatalf("got %d, %v; want 2 ads", n, err)
	}
	text := out.String()
	for _, want := range []string{"<p><strong>", "Café", "£", "✨", `"location":null`} {
		if !strings.Contains(text, want) {
			t.Errorf("output lacks %s", want)
		}
	}
	// One JSON object per line, each reading back as the ad it came from.
	lines := bufio.NewScanner(strings.NewReader(text))
	lines.Buffer(nil, 1<<20)
	rows := readAll(t, export)
	i := 0
	for ; lines.Scan(); i++ {
		var got ads.Ad
		if err := json.Unmarshal(lines.Bytes(), &got); err != nil {
			t.Fatalf("line %d: %v", i+1, err)
		}
		want, _ := rows[i].Ad()
		if got.ID != want.ID || got.HTML != want.HTML || (got.Location == nil) != (want.Location == nil) {
			t.Errorf("line %d: got %+v, want %+v", i+1, got, want)
		}
	}
	if i != 2 {
		t.Errorf("got %d lines, want 2", i)
	}
}

func TestConvertSaysWhichRowFailed(t *testing.T) {
	broken := strings.Replace(export, `<field name="jobtitle">Driver</field>`, "", 1)
	n, err := Convert(strings.NewReader(broken), io.Discard)
	if n != 1 || err == nil || !strings.Contains(err.Error(), "row 2: row lacks jobtitle") {
		t.Fatalf("got %d, %v; want 1 ad, then an error for row 2", n, err)
	}
}
