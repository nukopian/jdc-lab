// Package blocks splits a job ad's HTML into blocks: the paragraphs, lines
// and lists a reader sees, in order, each with its text as written.
//
// It records evidence and decides nothing. Whether a block is a heading, a
// field such as "Salary: £30,000", or plain text is for later steps, which
// look at its text and its Cues: what the markup said that the text alone
// doesn't, such as that all of it was bold. Keeping the two apart means the
// decisions can be measured and changed without touching the HTML.
package blocks

import (
	"strings"
	"unicode"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/nukopian/jdc-lab/internal/clean"
)

// A Kind is what sort of block a Block is.
type Kind int

const (
	// Paragraph is a block of text on its own: a <p>, a heading tag, a
	// <div>, or text outside any of them.
	Paragraph Kind = iota
	// Line is one of several lines a <br> or a newline split a block into.
	Line
	// List is consecutive list items, from <ul>, <ol> or bare <li>s.
	List
)

var kindNames = [...]string{Paragraph: "paragraph", Line: "line", List: "list"}

// String returns the kind's name, as JSON writes it: "paragraph".
func (k Kind) String() string {
	if k < 0 || int(k) >= len(kindNames) {
		return "unknown"
	}
	return kindNames[k]
}

// MarshalText writes the kind as its name, so JSON shows "line", not 1.
func (k Kind) MarshalText() ([]byte, error) {
	return []byte(k.String()), nil
}

// Cues are what the markup says about a piece of text that the text alone
// doesn't.
type Cues struct {
	// Heading is true for text in an <h1> to <h6>.
	Heading bool `json:"heading,omitempty"`
	// Bold is true when all the text is in <strong> or <b>.
	Bold bool `json:"bold,omitempty"`
	// BoldLead is the bold text it starts with, when not all of it is
	// bold: "Salary:" in "<strong>Salary:</strong> £30,000".
	BoldLead string `json:"bold_lead,omitempty"`
	// Added is the English label cleaning added at its start, for a
	// symbol that stood for it: "Location: " for "📍 Leeds".
	Added string `json:"added,omitempty"`
	// Bullet is the symbol it opened with, as a bullet (see
	// clean.Bullet): lines in a row with the same one are a list.
	Bullet string `json:"bullet,omitempty"`
	// Gap is true when a blank line came before it: two <br>s with
	// nothing between, or a blank line in text laid out by its newlines.
	// The space every paragraph has from the one before doesn't count.
	Gap bool `json:"gap,omitempty"`
	// RunIn is true for a heading internal/walls found run into the text
	// of a wall, at a sentence's start, and split off as a Line of its own:
	// "The Role" of "… completion. The Role Reporting to the MD …". Split
	// never sets it.
	RunIn bool `json:"run_in,omitempty"`
}

// An Item is a piece of text with its cues: a list item, or the text of a
// Paragraph or Line.
type Item struct {
	Text string `json:"text"`
	Cues Cues   `json:"cues,omitzero"`
}

// A Block is one block of an ad. A Paragraph or Line has Text and Cues; a
// List has Items instead.
type Block struct {
	Kind  Kind   `json:"kind"`
	Text  string `json:"text,omitempty"`
	Cues  Cues   `json:"cues,omitzero"`
	Items []Item `json:"items,omitempty"`
}

// Split returns the blocks of src, an HTML fragment, in order. Text is kept
// as written, with HTML's entities decoded and its runs of white space
// collapsed to one, as a browser shows them. Blocks with nothing visible,
// such as <p>&nbsp;</p>, are dropped. Broken HTML is read as far as it
// makes sense; nothing makes Split fail.
//
// Many ads are laid out by their newlines as much as by their tags, as if
// the newlines would show: some are plain text, with no tags that lay out
// blocks, and others have tags around only part of their text, such as a
// footer. So a newline ends a Line and a blank line ends a block, outside
// list items and heading tags, as the author laid them out, though a
// browser would show a space. In an ad with tags, only a newline with text
// either side, and no tag between, counts: one beside a tag is how the
// source is formatted. And there a newline before a lower-case letter, or
// after a comma, "&" or a word such as "to" that a line can't end with, is
// a line wrapped in the source, and stays a space. A wall of text with no
// newlines stays one Paragraph: splitting it is for later steps.
func Split(src string) []Block {
	b := builder{plain: !laidOut(src)}
	z := html.NewTokenizer(strings.NewReader(src))
	for {
		switch z.Next() {
		case html.ErrorToken: // the end of src: the tokenizer stops at nothing else
			b.finish()
			return bulleted(b.blocks)
		case html.TextToken:
			b.text(string(z.Text()))
		case html.StartTagToken, html.SelfClosingTagToken:
			name, _ := z.TagName()
			b.start(atom.Lookup(name))
		case html.EndTagToken:
			name, _ := z.TagName()
			b.end(atom.Lookup(name))
		}
	}
}

// builder turns the tokenizer's stream of tags and text into blocks. Its zero
// value is ready to use.
type builder struct {
	blocks []Block

	piece   piece  // the text being gathered
	lines   []Item // pieces of the current block that <br>s have ended
	items   []Item // the list being gathered
	bold    int    // how many <strong> or <b> we're inside
	skip    int    // how many <script> or <style> we're inside
	heading int    // how many <h1> to <h6> we're inside
	list    int    // how many <ul> or <ol> we're inside
	inItem  bool   // inside an <li>

	plain    bool // src is plain text, laid out by its newlines
	newlines int  // newlines since the last visible text
	gap      bool // a blank line came since the last line or block
	afterBr  bool // a <br> came since the last visible text, in this block

	inP   bool // inside a <p>, closed or not
	pBold int  // how many <strong> or <b> we were inside when it opened
}

func (b *builder) text(s string) {
	if b.skip > 0 {
		return
	}
	if visible(s) {
		b.afterBr = false
	}
	if !b.plain && (b.inItem || b.heading > 0) {
		b.piece.write(s, b.bold > 0) // an item or heading stays one piece
		return
	}
	// Newlines end lines and blocks, but only once more visible text comes:
	// spaces and further newlines in between change nothing, and trailing
	// ones end nothing. In HTML, a newline before a lower-case letter, or
	// after a word a line can't end with, is a line wrapped in the source,
	// not laid out, so it stays a space: ads with tags lay out as many lines
	// by their newlines as plain text does, but wrap many more. And there,
	// only a newline with text either side, in one run of text, counts: one
	// beside a tag is how the source is formatted, often a tag to a line.
	s = strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\r", "\n")
	if !b.plain {
		b.newlines = 0
	}
	start, seen := 0, false
	for i, r := range s {
		switch {
		case r == '\n' && (b.plain || seen):
			b.piece.write(s[start:i], b.bold > 0)
			start = i + 1
			b.newlines++
		case b.newlines > 0 && visible(string(r)):
			b.piece.write(s[start:i], b.bold > 0)
			start = i
			switch {
			case !b.plain && (unicode.IsLower(r) || b.piece.continues()):
				b.piece.write(" ", b.bold > 0)
			case b.newlines > 1:
				b.endBlock()
				b.gap = b.started() // a blank line before all of it is no gap
			default:
				b.endLine()
			}
			b.newlines = 0
		}
		if !seen && visible(string(r)) {
			seen = true
		}
	}
	b.piece.write(s[start:], b.bold > 0)
}

func (b *builder) start(tag atom.Atom) {
	switch {
	case tag == atom.Script || tag == atom.Style:
		b.skip++
	case tag == atom.Strong || tag == atom.B:
		b.bold++
	case tag == atom.Br:
		switch {
		case b.heading > 0:
			b.piece.write(" ", false) // a heading stays one piece
		case b.inItem:
			b.leaveItem()
		default:
			if b.afterBr && !visible(b.piece.text.String()) && b.started() {
				b.gap = true // a <br> with nothing since the last one
			}
			b.endLine()
			b.afterBr = true
		}
	case isHeading(tag):
		b.closeP()
		b.endHeading()
		if b.inItem {
			b.leaveItem()
		} else {
			b.endBlock()
		}
		b.heading++
	case tag == atom.Li:
		b.closeP()
		b.endHeading()
		if b.inItem {
			b.endItem() // the item before, if its </li> was left out
		} else {
			b.endBlock() // the text before a bare <li>, or between items
		}
		b.inItem = true
	case tag == atom.Ul || tag == atom.Ol:
		b.closeP()
		b.endHeading()
		if b.inItem {
			b.endItem() // the outer item's own text, before its sublist
		} else {
			b.endBlock() // the text before the list, in an item's own or not
		}
		b.list++
	case breaks(tag):
		b.closeP()
		b.endHeading()
		b.boundary()
		if tag == atom.P {
			b.inP, b.pBold = true, b.bold
		}
	}
}

func (b *builder) end(tag atom.Atom) {
	switch {
	case tag == atom.Script || tag == atom.Style:
		b.skip = max(b.skip-1, 0)
	case tag == atom.Strong || tag == atom.B:
		b.bold = max(b.bold-1, 0)
	case isHeading(tag):
		if b.inItem {
			b.piece.write(" ", false)
			b.heading = max(b.heading-1, 0)
		} else {
			b.endHeading()
		}
	case tag == atom.Li:
		switch {
		case b.inItem:
			b.endItem()
		case b.list > 0: // the outer item's own text, after its sublist
			b.endBlock()
		} // else a stray </li>, which ends nothing
	case tag == atom.Ul || tag == atom.Ol:
		if b.list == 0 {
			break // a stray </ul>, which ends nothing
		}
		if b.inItem {
			b.endItem()
		} else {
			b.endBlock() // text in the list, in no item
		}
		b.list--
		if b.list == 0 {
			b.endList()
		}
	case breaks(tag):
		if tag == atom.P {
			b.closeP()
		}
		b.endHeading() // a heading tag left open ends with the block around it
		b.boundary()
	}
}

// endHeading ends a heading tag, if one is open, as a Paragraph with the
// heading cue: at its end tag, or at a block, list or heading that opens
// before it, as when the end tag was left out ("<h3>Heading<p>Text").
// Inside a list item, the heading's text stays in the item.
func (b *builder) endHeading() {
	if b.heading == 0 {
		return
	}
	b.heading = 0
	if b.inItem {
		return
	}
	if it, ok := b.piece.take(); ok {
		it.Cues.Heading = true
		it.Cues.Gap, b.gap = b.gap, false
		b.add(Block{Kind: Paragraph, Text: it.Text, Cues: it.Cues})
	}
}

// closeP ends a <p>, at its end tag or at a block or list that opens
// before it, closing it as a browser does. Bold opened in the paragraph
// ends with it, closed or not, though a browser carries it on: where ads
// leave a <strong> open, the paragraph after is a footer, not meant bold.
// Bold around the paragraph goes on.
func (b *builder) closeP() {
	if b.inP {
		b.bold = min(b.bold, b.pBold)
		b.inP = false
	}
}

// boundary handles a tag that ends a block, such as <p> or </div>.
func (b *builder) boundary() {
	if b.inItem {
		b.leaveItem()
	} else {
		b.endBlock()
	}
}

// leaveItem handles a <br>, or a tag that ends a block, inside a list item.
// Before the item's text it's only a space, so "<li><p>Drive</p></li>"
// stays one item. After it, the item ends there, and the text after it in
// the <li> is blocks of its own, in the ad's order: ads run headings,
// fields and footers into a list's items, its last most of all ("Drive
// safely<br>Skills and Experience"), and as part of an item they'd go
// unseen.
func (b *builder) leaveItem() {
	if !visible(b.piece.text.String()) {
		b.piece.write(" ", false)
		return
	}
	b.endItem()
}

// endLine ends the current piece at a <br>, keeping it as one of the
// block's lines.
func (b *builder) endLine() {
	if it, ok := b.piece.take(); ok {
		it.Cues.Gap, b.gap = b.gap, false
		b.lines = append(b.lines, it)
	}
}

// started reports whether any text has been gathered yet, in a line, a
// block or a list: a blank line before all of it is no gap.
func (b *builder) started() bool {
	return len(b.lines) > 0 || len(b.blocks) > 0 || len(b.items) > 0
}

// endBlock ends the current block: one Paragraph, or a Line for each piece
// if <br>s split it.
func (b *builder) endBlock() {
	b.afterBr = false
	b.endLine()
	if len(b.lines) == 0 {
		return
	}
	kind := Paragraph
	if len(b.lines) > 1 {
		kind = Line
	}
	for _, it := range b.lines {
		b.add(Block{Kind: kind, Text: it.Text, Cues: it.Cues})
	}
	b.lines = b.lines[:0]
}

// endItem ends the current list item, if there's one with any text.
func (b *builder) endItem() {
	if it, ok := b.piece.take(); ok {
		b.items = append(b.items, it)
		b.gap = false // a list's items have no gap: it would be the block's after it
	}
	b.inItem = false
}

// endList adds the list gathered so far, if any, as a block.
func (b *builder) endList() {
	if len(b.items) > 0 {
		b.blocks = append(b.blocks, Block{Kind: List, Items: b.items})
		b.items = nil
	}
}

// add adds a block, after the items gathered before it, so that the
// ad's order is kept: inside a list too, as when an item holds
// paragraphs or headings of its own, and the items after it start a
// new list.
func (b *builder) add(block Block) {
	b.endList()
	b.blocks = append(b.blocks, block)
}

// bulleted makes each run of two or more paragraphs or lines that open
// with the same bullet a list, each an item: ads lay lists out so as
// often as with list tags.
func bulleted(bs []Block) []Block {
	var out []Block
	for i := 0; i < len(bs); {
		j := i
		for j < len(bs) && bs[j].Kind != List && bs[j].Cues.Bullet != "" && bs[j].Cues.Bullet == bs[i].Cues.Bullet {
			j++
		}
		if j-i < 2 {
			out = append(out, bs[i])
			i++
			continue
		}
		list := Block{Kind: List}
		for _, b := range bs[i:j] {
			list.Items = append(list.Items, Item{Text: b.Text, Cues: b.Cues})
		}
		out = append(out, list)
		i = j
	}
	return out
}

// finish ends whatever is open at the end of the HTML.
func (b *builder) finish() {
	if b.inItem {
		b.endItem()
	}
	b.endBlock()
	b.endList()
}

// laidOut reports whether src has any tag that lays out blocks, as HTML
// does, rather than leaving that to its newlines, as plain text does.
func laidOut(src string) bool {
	z := html.NewTokenizer(strings.NewReader(src))
	for {
		switch z.Next() {
		case html.ErrorToken:
			return false
		case html.StartTagToken, html.SelfClosingTagToken, html.EndTagToken:
			name, _ := z.TagName()
			tag := atom.Lookup(name)
			if breaks(tag) || isHeading(tag) || tag == atom.Br || tag == atom.Li ||
				tag == atom.Ul || tag == atom.Ol {
				return true
			}
		}
	}
}

func isHeading(tag atom.Atom) bool {
	switch tag {
	case atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6:
		return true
	}
	return false
}

// breaks reports whether tag starts or ends a block, as a browser lays it
// out. Lists, list items, headings and <br> are handled on their own.
func breaks(tag atom.Atom) bool {
	switch tag {
	case atom.P, atom.Div, atom.Section, atom.Article, atom.Header, atom.Footer,
		atom.Blockquote, atom.Pre, atom.Table, atom.Tr, atom.Td, atom.Th,
		atom.Dl, atom.Dt, atom.Dd, atom.Hr,
		atom.Main, atom.Nav, atom.Aside, atom.Address, atom.Figure, atom.Figcaption,
		atom.Details, atom.Summary, atom.Form, atom.Fieldset, atom.Legend,
		atom.Hgroup, atom.Menu, atom.Center, atom.Caption:
		return true
	}
	return false
}

// piece gathers the text of one Item, noting which of it was bold. Its zero
// value is ready to use.
type piece struct {
	text     strings.Builder
	lead     strings.Builder // the bold text at the start
	leadDone bool            // plain text has ended the lead
	bold     bool            // some visible text was bold
	plain    bool            // some visible text wasn't
}

// write adds s, which was in bold or not.
func (p *piece) write(s string, bold bool) {
	seen := visible(s)
	switch {
	case seen && bold:
		p.bold = true
	case seen:
		p.plain = true
	}
	if !p.leadDone {
		if seen && !bold {
			p.leadDone = true
		} else {
			collapse(&p.lead, s)
		}
	}
	collapse(&p.text, s)
}

// continues reports whether the text so far stops where a line can't end,
// so that a newline after it is a line wrapped, whatever comes next: after
// a comma or "&", or a word such as "to" or "and" that needs more after it
// ("Welcome to" before a name).
func (p *piece) continues() bool {
	t := strings.TrimRight(p.text.String(), " ")
	if strings.HasSuffix(t, ",") || strings.HasSuffix(t, "&") {
		return true
	}
	return joining[strings.ToLower(t[strings.LastIndexByte(t, ' ')+1:])]
}

// joining are the words a line doesn't end with.
var joining = map[string]bool{
	"a": true, "an": true, "the": true, "and": true, "or": true, "of": true, "to": true, "in": true,
	"on": true, "at": true, "for": true, "with": true, "by": true, "from": true, "as": true,
	"our": true, "your": true, "their": true, "its": true,
}

// take returns what's been gathered as an Item, and whether it has anything
// visible, and starts afresh.
func (p *piece) take() (Item, bool) {
	// The text is cleaned for the stages after: see internal/clean.
	text, added := clean.Line(p.text.String())
	if p.bold && !p.plain { // all bold, a heading's, so no label
		text, added = clean.Text(p.text.String()), ""
	}
	it := Item{Text: text, Cues: Cues{Added: added, Bullet: clean.Bullet(p.text.String())}}
	it.Cues.Bold = p.bold && !p.plain
	if p.bold && p.plain && p.leadDone {
		it.Cues.BoldLead, it.Text = clean.Lead(p.lead.String(), text)
	}
	ok := visible(text)
	*p = piece{}
	return it, ok
}

// collapse appends s to w with each run of HTML white space (space, tab,
// newline, carriage return, form feed) made one space, as a browser shows
// it, and none at the start. Other spaces, such as the no-break space, are
// kept as they are.
func collapse(w *strings.Builder, s string) {
	for _, r := range s {
		if !isHTMLSpace(r) {
			w.WriteRune(r)
			continue
		}
		if t := w.String(); t != "" && t[len(t)-1] != ' ' {
			w.WriteByte(' ')
		}
	}
}

func isHTMLSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\f'
}

// visible reports whether s has anything a reader would see: not only
// spaces of any kind, and invisible characters such as the zero-width space.
func visible(s string) bool {
	for _, r := range s {
		if !unicode.IsSpace(r) && !invisible(r) {
			return true
		}
	}
	return false
}

func invisible(r rune) bool {
	switch r {
	case '\u00ad', // soft hyphen
		'\u200b', '\u200c', '\u200d', // zero-width space, non-joiner, joiner
		'\u2060', '\ufeff': // word joiner, byte order mark
		return true
	}
	return false
}
