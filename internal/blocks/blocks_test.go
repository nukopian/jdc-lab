package blocks

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"
)

// show writes blocks as one line each, for comparing: the kind, the text,
// and any cues; a list's items follow it, indented.
func show(blocks []Block) []string {
	var out []string
	item := func(prefix string, it Item) string {
		s := fmt.Sprintf("%s%q", prefix, it.Text)
		if it.Cues.Heading {
			s += " heading"
		}
		if it.Cues.Bold {
			s += " bold"
		}
		if it.Cues.BoldLead != "" {
			s += fmt.Sprintf(" lead=%q", it.Cues.BoldLead)
		}
		if it.Cues.Added != "" {
			s += fmt.Sprintf(" added=%q", it.Cues.Added)
		}
		return s
	}
	for _, b := range blocks {
		if b.Kind == List {
			out = append(out, "list")
			for _, it := range b.Items {
				out = append(out, item("  ", it))
			}
			continue
		}
		out = append(out, item(b.Kind.String()+" ", Item{Text: b.Text, Cues: b.Cues}))
	}
	return out
}

func TestSplit(t *testing.T) {
	tests := []struct {
		name string
		html string
		want []string
	}{
		{
			"a bold heading, then text",
			`<p><strong>The role</strong></p><p>We build things.</p>`,
			[]string{`paragraph "The role" bold`, `paragraph "We build things."`},
		},
		{
			"a field with a bold label",
			`<p><strong>Salary:</strong> £30,000</p>`,
			[]string{`paragraph "Salary: £30,000" lead="Salary:"`},
		},
		{
			"fields on lines split by <br>",
			`<p><strong>Hours:</strong> 40<br><strong>Contract:</strong> Permanent</p>`,
			[]string{`line "Hours: 40" lead="Hours:"`, `line "Contract: Permanent" lead="Contract:"`},
		},
		{
			"a heading and its text in one paragraph",
			`<p><strong>The role</strong><br>We build things.</p>`,
			[]string{`line "The role" bold`, `line "We build things."`},
		},
		{
			"a trailing <br> leaves one paragraph",
			`<p>Text<br></p><p><br>More</p>`,
			[]string{`paragraph "Text"`, `paragraph "More"`},
		},
		{
			"spacers are dropped",
			"<p> </p><p>&nbsp;</p><p><strong> </strong></p><p>\u200b</p><p><br></p>",
			nil,
		},
		{
			"a list, with a paragraph inside an item",
			`<ul><li>Drive</li><li><p>Deliver</p></li></ul>`,
			[]string{"list", `  "Drive"`, `  "Deliver"`},
		},
		{
			"text before bare items, and after them at the end",
			`Intro<li>One</li>Outro`,
			[]string{`paragraph "Intro"`, "list", `  "One"`, `paragraph "Outro"`},
		},
		{
			"bare items make a list, ended by what follows",
			`<li>One</li><li>Two</li><p>After</p>`,
			[]string{"list", `  "One"`, `  "Two"`, `paragraph "After"`},
		},
		{
			"a sublist joins its list",
			`<ul><li>Shifts<ul><li>Days</li><li>Nights</li></ul></li><li>Pay</li></ul>`,
			[]string{"list", `  "Shifts"`, `  "Days"`, `  "Nights"`, `  "Pay"`},
		},
		{
			"items whose </li> is left out",
			`<ul><li>One<li>Two</ul>`,
			[]string{"list", `  "One"`, `  "Two"`},
		},
		{
			"a <br> in an item is a space",
			`<ul><li>Days<br>Nights</li></ul>`,
			[]string{"list", `  "Days Nights"`},
		},
		{
			"a bold item",
			`<ul><li><strong>Benefits</strong></li><li>Pension</li></ul>`,
			[]string{"list", `  "Benefits" bold`, `  "Pension"`},
		},
		{
			"text outside any tag, between blocks",
			`<p>Intro</p>Loose text<ul><li>Item</li></ul>`,
			[]string{`paragraph "Intro"`, `paragraph "Loose text"`, "list", `  "Item"`},
		},
		{
			"a heading tag",
			`<h3>About <em>us</em></h3><p>We care.</p>`,
			[]string{`paragraph "About us" heading`, `paragraph "We care."`},
		},
		{
			"entities decoded, white space collapsed, the no-break space made plain",
			"<p>Fish &amp; chips\n\t  for &pound;5&nbsp;each &lt;p&gt;</p>",
			[]string{`paragraph "Fish & chips for £5 each <p>"`},
		},
		{
			"script and style dropped",
			`<style>p{}</style><p>Text<script>x()</script></p>`,
			[]string{`paragraph "Text"`},
		},
		{
			// Text is cleaned (see internal/clean): an emoji bullet goes,
			// and with it the bold lead it was.
			"an emoji bullet in bold",
			`<p><strong>✏️ </strong>Enjoy flexibility.</p>`,
			[]string{`paragraph "Enjoy flexibility."`},
		},
		{
			"a label symbol becomes its English label",
			`<p>📍 Leeds</p><p><b>💷 </b>£30,000 a year</p><ul><li>✅ Long term work</li></ul>`,
			[]string{`paragraph "Location: Leeds" added="Location: "`, `paragraph "Pay: £30,000 a year" added="Pay: "`,
				"list", `  "Long term work"`},
		},
		{
			// Lines that open with the same symbol, not closing with it,
			// are a list; a lone one, or ones that differ, aren't.
			"lines opening with the same symbol are a list",
			`<p>✅ Weekly pay<br>✅ Free parking<br>✔️ Training</p><p>📢 Hiring now 📢</p><p>📢 Apply today</p>`,
			[]string{"list", `  "Weekly pay"`, `  "Free parking"`, `line "Training"`,
				`paragraph "Hiring now"`, `paragraph "Apply today"`},
		},
		{
			"paragraphs opening with the same symbol are a list too",
			`<p>✨ One</p><p>✨ Two</p><p>Then text.</p>`,
			[]string{"list", `  "One"`, `  "Two"`, `paragraph "Then text."`},
		},
		{
			// All in bold, it's a heading's, and takes no label.
			"a bold line opening with a label symbol keeps no label",
			`<p><strong>💷 What We Offer</strong></p><p><b>📍 The Role</b></p>`,
			[]string{`paragraph "What We Offer" bold`, `paragraph "The Role" bold`},
		},
		{
			"typographic punctuation made plain",
			`<p>“Quoted” – text…</p>`,
			[]string{`paragraph "\"Quoted\" - text..."`},
		},
		{
			"bold later in the text is no cue",
			`<p>Pay is <b>great</b></p>`,
			[]string{`paragraph "Pay is great"`},
		},
		{
			"bold split across tags is still all bold",
			`<p><strong>What you'll</strong> <b>be doing</b></p>`,
			[]string{`paragraph "What you'll be doing" bold`},
		},
		{
			"plain text: a newline ends a line, a blank line a block",
			`The role
We build things.

Benefits:
- Pension

`,
			[]string{`line "The role"`, `line "We build things."`, `line "Benefits:"`, `line "- Pension"`},
		},
		{
			"plain text: one line on its own is a paragraph",
			`Driver needed.

Apply now.`,
			[]string{`paragraph "Driver needed."`, `paragraph "Apply now."`},
		},
		{
			"plain text keeps inline tags' cues, across newlines",
			`<b>Salary:</b> £30,000
<b>The role</b>
Drive.`,
			[]string{`line "Salary: £30,000" lead="Salary:"`, `line "The role" bold`, `line "Drive."`},
		},
		{
			"plain text with CRLF line ends",
			"One\r\nTwo\r\n\r\nThree",
			[]string{`line "One"`, `line "Two"`, `paragraph "Three"`},
		},
		{
			"a wall of text stays one paragraph",
			`We're hiring.What you'll be doingYou'll lead a team.How to applyEmail us.`,
			[]string{`paragraph "We're hiring.What you'll be doingYou'll lead a team.How to applyEmail us."`},
		},
		{
			"in HTML, newlines are only spacing",
			`<ul>
 <li>
 <p>Drive</p>
 </li>
</ul>
<p>Apply
now.</p>`,
			[]string{"list", `  "Drive"`, `paragraph "Apply now."`},
		},
		{
			"stray end tags change nothing",
			`</strong></ul></li><p>Text</p>`,
			[]string{`paragraph "Text"`},
		},
		{
			"a stray </li> or </ul> after text makes no list",
			`Intro</li><p>More</p>Outro</ul>`,
			[]string{`paragraph "Intro"`, `paragraph "More"`, `paragraph "Outro"`},
		},
		{
			"a heading tag inside an item stays in the item",
			`<ul><li>Intro <h3>Head</h3> more</li></ul>`,
			[]string{"list", `  "Intro Head more"`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := show(Split(tt.html))
			if !slices.Equal(got, tt.want) {
				t.Errorf("Split(%q)\n got: %q\nwant: %q", tt.html, got, tt.want)
			}
		})
	}
}

func TestJSON(t *testing.T) {
	blocks := Split(`<p><strong>Hours:</strong> 40<br>Mon-Fri</p><ul><li>Pension</li></ul>`)
	data, err := json.Marshal(blocks)
	if err != nil {
		t.Fatal(err)
	}
	want := `[{"kind":"line","text":"Hours: 40","cues":{"bold_lead":"Hours:"}},` +
		`{"kind":"line","text":"Mon-Fri"},` +
		`{"kind":"list","items":[{"text":"Pension"}]}]`
	if string(data) != want {
		t.Errorf("\n got: %s\nwant: %s", data, want)
	}
}

func TestKindString(t *testing.T) {
	for k, want := range map[Kind]string{Paragraph: "paragraph", Line: "line", List: "list", 7: "unknown", -1: "unknown"} {
		if got := k.String(); got != want {
			t.Errorf("Kind(%d).String() = %q, want %q", int(k), got, want)
		}
	}
}

// FuzzSplit checks that no HTML makes Split panic, and that every block it
// returns is well formed.
func FuzzSplit(f *testing.F) {
	for _, seed := range []string{
		`<p><strong>Salary:</strong> £30,000<br>Hours</p>`,
		`<ul><li>One<li>Two<ul><li>Three</ul></ul>`,
		`<h2>About</h2><p>&nbsp;</p><script>x</script>`,
		`</b></b><b><li></p>text`,
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, src string) {
		for _, b := range Split(src) {
			switch b.Kind {
			case Paragraph, Line:
				if !visible(b.Text) || b.Items != nil {
					t.Fatalf("%s with text %q and %d items", b.Kind, b.Text, len(b.Items))
				}
			case List:
				if len(b.Items) == 0 || b.Text != "" {
					t.Fatalf("list with text %q and no items", b.Text)
				}
				for _, it := range b.Items {
					if !visible(it.Text) {
						t.Fatalf("list item %q", it.Text)
					}
				}
			default:
				t.Fatalf("kind %d", b.Kind)
			}
			if b.Text != strings.TrimSpace(b.Text) {
				t.Fatalf("untrimmed %q", b.Text)
			}
		}
	})
}
