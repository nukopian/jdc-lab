package clean

import "testing"

func TestText(t *testing.T) {
	for in, want := range map[string]string{
		"✏️ Enjoy the flexibility":              "Enjoy the flexibility",
		"📢 Calling All Teachers 📢":              "Calling All Teachers",
		"“Quoted” and ‘single’":                 `"Quoted" and 'single'`,
		"Monday–Friday — 9am–5pm":               "Monday-Friday -- 9am-5pm",
		"a rest—then more":                      "a rest -- then more",
		"‹single› «double» „low“ ‚low‘ ′prime″": `'single' "double" "low" 'low' 'prime"`,
		"non‑breaking‐hyphen ‒figure":           "non-breaking-hyphen -figure",
		"full－width！ ？":                         "full-width! ?",
		"And there’s more…":                     "And there's more...",
		"Pay:\u00a0£30,000\u202fa year":         "Pay: £30,000 a year",
		"zero\u200bwidth\ufeff and \u0002ctl":   "zerowidth and ctl",
		"• point ✔ done ▪ more":                 "point done more",
		"Café Zoë, Señor":                       "Café Zoë, Señor",
		"broken \ufffd text":                    "broken text",
		"\uf0b7 Word's own bullet":              "Word's own bullet",
	} {
		if got := Text(in); got != want {
			t.Errorf("Text(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLine(t *testing.T) {
	tests := []struct{ in, text, added string }{
		// A leading "*" or "- " is a bullet, and goes; one later stays.
		{"*Servicing, Installation and Maintenance", "Servicing, Installation and Maintenance", ""},
		{"- Full-time teaching position.", "Full-time teaching position.", ""},
		{"£30,000* (Ts & Cs apply)", "£30,000* (Ts & Cs apply)", ""},
		{"📍 Leeds, West Yorkshire", "Location: Leeds, West Yorkshire", "Location: "},
		{"💷 £30,000 per annum", "Pay: £30,000 per annum", "Pay: "},
		{"💰Competitive salary", "Pay: Competitive salary", "Pay: "},
		{"⏰ Monday to Friday, 8am-5pm", "Hours: Monday to Friday, 8am-5pm", "Hours: "},
		{"📞 Call Vicki on 0113", "Contact: Call Vicki on 0113", "Contact: "},
		{"📧 jobs@example.com", "Email: jobs@example.com", "Email: "},
		// A line with a label of its own keeps it, and takes no other.
		{"📍 Location: Leeds", "Location: Leeds", ""},
		{"💷 Salary - £30,000", "Salary - £30,000", ""},
		// Symbols with no one meaning just go.
		{"📅 Start date: ASAP", "Start date: ASAP", ""},
		{"✅ Long term work", "Long term work", ""},
		// A label symbol later in a line isn't a label.
		{"Based in 📍 Leeds", "Based in Leeds", ""},
		{"📍", "", ""},
		// A line ending as a heading does takes no label: it's a heading.
		{"💰 WHY APPLY?", "WHY APPLY?", ""},
		{"📞 Apply Today!", "Apply Today!", ""},
		{"💷 Benefits:", "Benefits:", ""},
		// A label only for a value it could be: a heading stays one.
		{"💷 What We Offer", "What We Offer", ""},
		{"💷 Salary", "Salary", ""},
		{"💰 Competitive salary + bonus", "Pay: Competitive salary + bonus", "Pay: "},
		{"⏰ Full time, 40 hours", "Hours: Full time, 40 hours", "Hours: "},
		{"⏰ Working with us", "Working with us", ""},
		{"📞 (phone number removed)", "Contact: (phone number removed)", "Contact: "},
		{"📧 Apply today", "Apply today", ""},
	}
	for _, tt := range tests {
		text, added := Line(tt.in)
		if text != tt.text || added != tt.added {
			t.Errorf("Line(%q) = %q, %q; want %q, %q", tt.in, text, added, tt.text, tt.added)
		}
	}
}

func TestBullet(t *testing.T) {
	for in, want := range map[string]string{
		"*Servicing and maintenance":    "*",
		"* Home based":                  "*",
		"- Full-time teaching position": "-",
		"-5 degrees at night":           "", // a minus, not a bullet
		"£30,000*":                      "",
		"✅ Long term work":              "✅",
		"  ✔️ Weekly pay":               "✔",
		"• A point":                     "•",
		"📢 Calling all teachers 📢":      "", // the same symbol at its end: decoration
		"Plain text":                    "",
		"📍 Leeds":                       "📍",
		"✅":                             "", // nothing after it
	} {
		if got := Bullet(in); got != want {
			t.Errorf("Bullet(%q) = %q, want %q", in, got, want)
		}
	}
}
