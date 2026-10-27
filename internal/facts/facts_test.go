package facts

import (
	"fmt"
	"strings"
	"testing"
)

// show writes the facts found in a line, one a line, values sorted.
func show(fs []Fact) string {
	var out []string
	for _, f := range fs {
		out = append(out, fmt.Sprintf("%s %q %v", f.Kind, f.Text, f.Values))
	}
	return strings.Join(out, "\n")
}

func TestFind(t *testing.T) {
	tests := []struct{ line, want string }{
		{"50,000pa + excellent bonus and benefits", `pay "50,000pa" map[min:50000 period:year]`},
		{"Up to £45k - £55,000 per annum", `pay "Up to £45k - £55,000 per annum" map[currency:GBP max:55000 min:45000 period:year]`},
		{"£12.71ph", `pay "£12.71ph" map[currency:GBP min:12.71 period:hour]`},
		{"Pay: £130 - £170 per day", `pay "£130 - £170 per day" map[currency:GBP max:170 min:130 period:day]`},
		{"16.75ph", `pay "16.75ph" map[min:16.75 period:hour]`},
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
			t.Errorf("Find(%q):\n got %s\nwant %s", tt.line, got, tt.want)
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
			t.Errorf("Find(%q):\n got %s\nwant %s", tt.line, got, tt.want)
		}
	}
}
