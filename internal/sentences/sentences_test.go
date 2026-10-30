package sentences

import (
	"slices"
	"strings"
	"testing"
)

func TestSplit(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []string
	}{
		{"a stop and a capital", "We're hiring. You'll drive.", []string{"We're hiring.", "You'll drive."}},
		{"question and exclamation marks", "Is it you? Apply now! We reply fast.",
			[]string{"Is it you?", "Apply now!", "We reply fast."}},
		{"no stop at the end", "We're hiring. No stop at the end", []string{"We're hiring.", "No stop at the end"}},
		{"one sentence", "Drive safely", []string{"Drive safely"}},
		{"nothing", "", nil},
		{"a lower-case word after a stop goes on", "Visit example. com to apply.", []string{"Visit example. com to apply."}},
		{"a figure after a stop goes on", "Grade 3. 2 posts are open.", []string{"Grade 3. 2 posts are open."}},
		{"abbreviations", "Use tools, e.g. Sage or Xero. Report to Mr. Patel, i.e. The Head. Start soon.",
			[]string{"Use tools, e.g. Sage or Xero.", "Report to Mr. Patel, i.e. The Head.", "Start soon."}},
		{"initials", "Contact J. Smith or A. N. Other today. We reply fast.",
			[]string{"Contact J. Smith or A. N. Other today.", "We reply fast."}},
		{"pay before a stop", "Earn £45k+. Plus overtime.", []string{"Earn £45k+.", "Plus overtime."}},
		{"closing quotes and brackets stay with their sentence",
			`They said "Great place." (Full training given.) Apply now.`,
			[]string{`They said "Great place."`, `(Full training given.)`, "Apply now."}},
		{"an opening quote before the capital", `Apply now. "Best job ever," says Sam.`,
			[]string{"Apply now.", `"Best job ever," says Sam.`}},
		{"an ellipsis ends one", "And more... Apply today.", []string{"And more...", "Apply today."}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Split(tt.text)
			if !slices.Equal(got, tt.want) {
				t.Errorf("Split(%q)\n got: %q\nwant: %q", tt.text, got, tt.want)
			}
			if joined := strings.Join(got, " "); joined != tt.text {
				t.Errorf("Split(%q) joined is %q: not the text", tt.text, joined)
			}
		})
	}
}
