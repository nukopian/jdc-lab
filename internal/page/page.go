// Package page writes an ad before and after chunking, as a web page to
// read: the ad as posted beside its chunks (see internal/chunk), each
// coloured by what's done with it, embedded, kept as metadata or left
// out, with its size and what placed it, and a bar of where the ad's
// tokens go. An index page lists the ads, with why each was chosen.
//
// The pages are written with html/template, which escapes every value by
// where it goes in the page, so no text of an ad can break it. The ad's
// own HTML is shown as posted, but first cut down to the tags that lay
// text out, with none of their attributes, and no scripts or styles.
package page

import (
	"embed"
	"html/template"
	"io"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/nukopian/jdc-lab/internal/ads"
	"github.com/nukopian/jdc-lab/internal/chunk"
)

//go:embed page.html index.html
var files embed.FS

var templates = template.Must(template.New("").Funcs(template.FuncMap{
	"lines": lines,
	"class": class,
}).ParseFS(files, "page.html", "index.html"))

// An Ad is one ad to show: the ad, its chunks, and why it was chosen.
type Ad struct {
	Ad     ads.Ad
	Chunks chunk.Ad
	Why    string
}

// Shares are how many of an ad's tokens each use takes.
type Shares struct {
	Embed, Metadata, LeaveOut, Boilerplate int
}

// shares adds up an ad's chunks' tokens by use, the boilerplate chunk on
// its own.
func shares(c chunk.Ad) Shares {
	var s Shares
	for _, ch := range c.Chunks {
		switch {
		case ch.Kind == chunk.Boilerplate:
			s.Boilerplate += ch.Tokens
		case ch.Use == chunk.Embed:
			s.Embed += ch.Tokens
		case ch.Use == chunk.Metadata:
			s.Metadata += ch.Tokens
		default:
			s.LeaveOut += ch.Tokens
		}
	}
	return s
}

// Write writes a's page to w; title names the set of pages it's in.
func Write(w io.Writer, a Ad, title string) error {
	s := shares(a.Chunks)
	return templates.ExecuteTemplate(w, "page.html", struct {
		Ad
		Title  string
		Posted template.HTML
		Shares Shares
		Pct    [4]int
		All    int
	}{a, title, Posted(a.Ad.HTML), s, percents(s), total(s)})
}

// An Entry is a line of the index: an ad, and the file of its page.
type Entry struct {
	Ad
	File string
}

// WriteIndex writes the index of the pages to w.
func WriteIndex(w io.Writer, title string, es []Entry) error {
	type row struct {
		Entry
		Pct [4]int
	}
	rows := make([]row, len(es))
	for i, e := range es {
		rows[i] = row{e, percents(shares(e.Chunks))}
	}
	return templates.ExecuteTemplate(w, "index.html", struct {
		Title string
		Rows  []row
	}{title, rows})
}

func total(s Shares) int { return s.Embed + s.Metadata + s.LeaveOut + s.Boilerplate }

// keep are the tags an ad keeps when shown as posted: those that lay its
// text out.
var keep = map[string]bool{
	"p": true, "br": true, "ul": true, "ol": true, "li": true, "strong": true, "b": true,
	"em": true, "i": true, "u": true, "h1": true, "h2": true, "h3": true, "h4": true,
	"h5": true, "h6": true, "table": true, "thead": true, "tbody": true, "tr": true,
	"td": true, "th": true, "div": true, "span": true, "hr": true, "blockquote": true,
}

// dropped are the tags an ad loses with everything in them: what runs,
// styles, or isn't shown as text.
var dropped = map[string]bool{
	"script": true, "style": true, "noscript": true, "iframe": true, "textarea": true, "xmp": true,
	"title": true, "template": true, "svg": true, "math": true, "noembed": true, "object": true,
}

// Posted returns an ad's HTML cut down to the tags in keep, with no
// attributes, and without scripts, styles and the like or anything in
// them: safe to put in a page as it is. It's parsed as a browser would,
// and written back from the parse, so every tag it keeps is closed, and
// none closes what's outside it.
func Posted(src string) template.HTML {
	body := &html.Node{Type: html.ElementNode, Data: "body", DataAtom: atom.Body}
	nodes, err := html.ParseFragment(strings.NewReader(src), body)
	if err != nil {
		return template.HTML(html.EscapeString(src))
	}
	var b strings.Builder
	var write func(n *html.Node)
	write = func(n *html.Node) {
		switch {
		case n.Type == html.TextNode:
			b.WriteString(html.EscapeString(n.Data))
			return
		case n.Type != html.ElementNode || dropped[n.Data]:
			return
		}
		if keep[n.Data] {
			b.WriteString("<" + n.Data + ">")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			write(c)
		}
		if keep[n.Data] && n.Data != "br" && n.Data != "hr" {
			b.WriteString("</" + n.Data + ">")
		}
	}
	for _, n := range nodes {
		write(n)
	}
	return template.HTML(b.String())
}

// A line is a line of a chunk's text: a list's item, or not, and for the
// boilerplate chunk, where it came from.
type line struct {
	Text, From string
	Item       bool
}

// lines returns a chunk's lines to show, its heading, shown above them,
// left out.
func lines(ch chunk.Chunk) []line {
	ls := strings.Split(ch.Text, "\n")
	if ch.Heading != "" && len(ls) > 0 && ls[0] == ch.Heading {
		ls = ls[1:]
	}
	var out []line
	for _, l := range ls {
		if l == "" {
			continue
		}
		if ch.Kind == chunk.Boilerplate {
			// "summary: text", or `unmapped "Heading": text`.
			from, text, ok := strings.Cut(l, ": ")
			if !ok {
				from, text = "", l
			}
			out = append(out, line{Text: text, From: from, Item: true})
			continue
		}
		text, item := strings.CutPrefix(l, "- ")
		out = append(out, line{Text: text, Item: item})
	}
	return out
}

// percents returns each of s's shares of their total, in whole percent,
// rounded so that they add up to 100: the largest remainders round up.
func percents(s Shares) [4]int {
	ns := [4]int{s.Embed, s.Metadata, s.LeaveOut, s.Boilerplate}
	all := total(s)
	var ps [4]int
	if all == 0 {
		return ps
	}
	left := 100
	rem := make([]int, 4)
	for i, n := range ns {
		ps[i] = 100 * n / all
		left -= ps[i]
		rem[i] = 100 * n % all
	}
	for ; left > 0; left-- {
		top := 0
		for i := range rem {
			if rem[i] > rem[top] {
				top = i
			}
		}
		ps[top]++
		rem[top] = -1
	}
	return ps
}

// class returns the CSS class for a chunk: its use, or boilerplate.
func class(ch chunk.Chunk) string {
	switch {
	case ch.Kind == chunk.Boilerplate:
		return "boilerplate"
	case ch.Use == chunk.Embed:
		return "embed"
	case ch.Use == chunk.Metadata:
		return "metadata"
	}
	return "leave-out"
}
