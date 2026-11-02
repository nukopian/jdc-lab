package facts

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nukopian/grammar"

	"github.com/nukopian/jdc-lab/internal/grammars"
)

// show writes the facts found in a line, one a line, values sorted.
func show(fs []Fact) string {
	var out []string
	for _, f := range fs {
		out = append(out, fmt.Sprintf("%s %q %v", f.Kind, f.Text, f.Values))
	}
	return strings.Join(out, "\n")
}

// grammarOf is the grammar each kind of fact is found with.
var grammarOf = map[string]*grammar.Grammar{
	"pay":      grammars.Pay,
	"hours":    grammars.Hours,
	"pattern":  grammars.Pattern,
	"schedule": grammars.Schedule,
	"location": grammars.Postcode,
}

// whyMissing says, for each kind of fact in want (as show writes them)
// that got lacks, why Find found none in line: that MayHold skipped the
// line, or FindErr's report of where the kind's grammar gave up, or, when
// it matched, that Find turned the match down.
func whyMissing(line, got, want string) string {
	var why []string
	for w := range strings.Lines(want) {
		kind, _, _ := strings.Cut(w, " ")
		g := grammarOf[kind]
		if g == nil || strings.HasPrefix(got, kind+" ") || strings.Contains(got, "\n"+kind+" ") {
			continue
		}
		switch m, err := g.FindErr(line, grammar.Outermost()); {
		case !grammars.MayHold(g, line):
			why = append(why, fmt.Sprintf("no %s: MayHold skips the line", kind))
		case err != nil:
			why = append(why, fmt.Sprintf("no %s: %v", kind, err))
		default:
			why = append(why, fmt.Sprintf("no %s: %s matched %q, and Find turned it down", kind, g.Name(), strings.TrimSpace(m.At("~").Str())))
		}
	}
	if why == nil {
		return ""
	}
	return "\n" + strings.Join(why, "\n")
}

func TestFind(t *testing.T) {
	tests := []struct{ line, want string }{
		{"50,000pa + excellent bonus and benefits", `pay "50,000pa" map[min:50000 period:year]`},
		{"Up to £45k - £55,000 per annum", `pay "Up to £45k - £55,000 per annum" map[currency:GBP max:55000 min:45000 period:year]`},
		{"£12.71ph", `pay "£12.71ph" map[currency:GBP min:12.71 period:hour]`},
		{"Pay: £130 - £170 per day", `pay "£130 - £170 per day" map[currency:GBP max:170 min:130 period:day]`},
		{"16.75ph", `pay "16.75ph" map[min:16.75 period:hour]`},
		{"£12.50 p.h. plus", `pay "£12.50 p.h." map[currency:GBP min:12.5 period:hour]`},
		{"£30,000 p.a.", `pay "£30,000 p.a." map[currency:GBP min:30000 period:year]`},
		{"£500 pw", `pay "£500 pw" map[currency:GBP min:500 period:week]`},
		{"£500 p/w", `pay "£500 p/w" map[currency:GBP min:500 period:week]`},
		{"£500 p.w.", `pay "£500 p.w." map[currency:GBP min:500 period:week]`},
		{"£1,200 pcm", `pay "£1,200 pcm" map[currency:GBP min:1200 period:month]`},
		{"40hrs pw", `hours "40hrs pw" map[per_week:40]`},
		{"37.5 hours per week", `hours "37.5 hours per week" map[per_week:37.5]`},
		{"Full-Time | Permanent", "pattern \"Full-Time\" map[value:full time]\npattern \"Permanent\" map[value:permanent]"},
		{"Temp to perm, hybrid working", "pattern \"Temp to perm\" map[value:temp to perm]\npattern \"hybrid working\" map[value:hybrid]"},
		{"Monday to Friday, 8am - 5pm", `schedule "Monday to Friday, 8am - 5pm" map[days:Monday-Friday times:8am-5pm]`},
		{"Mon-Fri 07:00-15:30", `schedule "Mon-Fri 07:00-15:30" map[days:Mon-Fri times:07:00-15:30]`},
		{"M-F 8-4", `schedule "M-F 8-4" map[days:M-F times:8-4]`},
		{"4 on 4 off, days and nights", `schedule "4 on 4 off" map[rota:4 on 4 off]`},
		{"202 High Street, Gateshead NE8 1AQ.", `location "NE8 1AQ" map[postcode:NE8 1AQ]`},
		// Nothing: figures that aren't pay, a time with no day, words.
		{"Minimum 5 years' experience", ""},
		{"Projects from 2015 to 2020", ""},
		{"We support 300 people", ""},
		{"An exciting opportunity", ""},
	}
	for _, tt := range tests {
		if got := show(Find(tt.line)); got != tt.want {
			t.Errorf("Find(%q):\n got %s\nwant %s%s", tt.line, got, tt.want, whyMissing(tt.line, got, tt.want))
		}
	}
}

// Pay written with a pay label but no currency or period is pay too.
func TestFindLabelledPay(t *testing.T) {
	if got := show(Find("Pay Rate - 13.10")); got != `pay "13.10" map[min:13.1]` {
		t.Errorf("got %s", got)
	}
}

// Faults found in the wash-up.
func TestFindWashUp(t *testing.T) {
	tests := []struct{ line, want string }{
		// The k of a range's top figure is the bottom's too.
		{"£25-28k per annum", `pay "£25-28k per annum" map[currency:GBP max:28000 min:25000 period:year]`},
		{"£25 - 28k", `pay "£25 - 28k" map[currency:GBP max:28000 min:25000]`},
		// Hours a day aren't hours a week; hours with no period say so.
		{"12 hours per day", `hours "12 hours per day" map[per_day:12]`},
		{"Guaranteed 40 hours", `hours "40 hours" map[hours:40]`},
		// A bare h is hours only with a period: "24h support" isn't.
		{"24h support", ""},
		// More ways to write an hourly rate.
		{"£12.50/hr", `pay "£12.50/hr" map[currency:GBP min:12.5 period:hour]`},
		{"£14 per hr", `pay "£14 per hr" map[currency:GBP min:14 period:hour]`},
	}
	for _, tt := range tests {
		if got := show(Find(tt.line)); got != tt.want {
			t.Errorf("Find(%q):\n got %s\nwant %s%s", tt.line, got, tt.want, whyMissing(tt.line, got, tt.want))
		}
	}
}

// More hours than a period holds aren't hours of work: over 24 a day,
// over 168 a week, or, with no period, over 80, a long week's. Nor is
// what's left of such a figure, "24 hours" of "224 hours", or "026 Hours"
// of "2026 Hours".
func TestImplausibleHours(t *testing.T) {
	tests := []struct{ line, want string }{
		{"272 hours annual leave", ""},
		{"- 224 hours holiday", ""},
		{"Dec 2026 Hours : 37.5 hours per week", `hours "37.5 hours per week" map[per_week:37.5]`},
		{"1800-0600hrs", ""},
		{"168 hours per month", ""},
		{"84 Hours per Fortnight", ""},
		{"25 hours per day", ""},
		// The limits themselves are plausible.
		{"24 hours per day", `hours "24 hours per day" map[per_day:24]`},
		{"80 hours", `hours "80 hours" map[hours:80]`},
		{"60 hours per week", `hours "60 hours per week" map[per_week:60]`},
	}
	for _, tt := range tests {
		if got := show(Find(tt.line)); got != tt.want {
			t.Errorf("Find(%q):\n got %s\nwant %s%s", tt.line, got, tt.want, whyMissing(tt.line, got, tt.want))
		}
	}
}

// A day's stop, as in "Mon." or at a sentence's end, isn't part of its
// value, so the same days state one fact.
func TestDaysWithoutStops(t *testing.T) {
	for _, line := range []string{"Monday to Friday.", "Mon. to Fri. 9am-5pm", "Monday to Friday"} {
		fs := Find(line)
		if len(fs) == 0 || fs[0].Kind != "schedule" {
			t.Fatalf("Find(%q) = %+v, want a schedule", line, fs)
		}
		if d := fs[0].Values["days"]; d != "Monday-Friday" && d != "Mon-Fri" {
			t.Errorf("Find(%q) days = %q, want no stops", line, d)
		}
	}
}

// "Contract" as a field's label states nothing ("Contract type: Temp to
// Perm" is temp to perm, not a contract), though other patterns before a
// colon are values; and a bonus isn't pay.
func TestLabelsAndBonuses(t *testing.T) {
	tests := []struct {
		line string
		want []string // kind: text
	}{
		{"Contract Type: Temp to Perm", []string{"pattern: Temp to Perm"}},
		{"Contract: Permanent", []string{"pattern: Permanent"}},
		{"Contract length: 12 months, full time", []string{"pattern: full time"}},
		{"A 12-month contract, full time", []string{"pattern: contract", "pattern: full time"}},
		{"Part-Time: 25 Hours per Week", []string{"pattern: Part-Time"}},
		{"Permanent: 37.5 hours per week", []string{"pattern: Permanent"}},
		{"£1,500 joining bonus for engineers", nil},
		{"£30 bonus for short-notice cover", nil},
		{"£32,500 basic salary plus bonus", []string{"pay: £32,500"}},
		{"£28k + bonus", []string{"pay: £28k"}},
		{"£30,000 plus bonus", []string{"pay: £30,000"}},
		{"£30,000 per annum plus bonus", []string{"pay: £30,000 per annum"}},
		{"£30k with bonus", []string{"pay: £30k"}},
	}
	for _, tt := range tests {
		var got []string
		for _, f := range Find(tt.line) {
			if f.Kind == "pattern" || f.Kind == "pay" {
				got = append(got, f.Kind+": "+f.Text)
			}
		}
		if strings.Join(got, "|") != strings.Join(tt.want, "|") {
			t.Errorf("Find(%q) = %q, want %q", tt.line, got, tt.want)
		}
	}
}

// Found in the wash-up: a figure in digits other than ASCII's, which \d
// matches, has its value, not 0.
func TestFiguresInOtherDigits(t *testing.T) {
	for line, want := range map[string]string{
		"Up to ٣٠ per hour":         `pay "Up to ٣٠ per hour" map[max:30 period:hour]`,
		"Salary: £٣٠,٠٠٠ per annum": `pay "£٣٠,٠٠٠ per annum" map[currency:GBP min:30000 period:year]`,
		"٣٧.٥ hours per week":       `hours "٣٧.٥ hours per week" map[per_week:37.5]`,
	} {
		if got := show(Find(line)); got != want {
			t.Errorf("Find(%q)\n got: %s\nwant: %s%s", line, got, want, whyMissing(line, got, want))
		}
	}
}
