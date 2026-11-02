# To do

## Check pay in the Pay grammar, with a code assertion

**From:** grammar's session, 2026-11-02 (grammar 24b190ab4ab18da57055eb91a43627f59d8a2d23)
**Waiting on:** a grammar release with code assertions (after v0.4.4)

**What:** Pay's rule that a figure is pay only with a currency sign or a
period is checked after the search, in two places kept in step by hand:
`facts.payOf` and the `cut` of Pay in `headings.onlyFacts`. grammar's
code assertions (roadmap 3) put it in the grammar, once:
`<?{ pay($/) }>` before `<!before <.bonus>>`, calling a Go function
passed to `MustCompile` as `grammar.Assertion("pay", fn)`.

**Why:** one copy of the rule, in the grammar, beside the rest of what
makes a figure pay; and a match that isn't pay is never made, rather
than made and thrown away, so it can't hide one inside it under
`Outermost`.

**What's known:** tried against grammar's checkout on 2026-11-02 with
the patch below. Chunking all 28,000 ads of the latest run
(`xargs chunk -json < ids.txt`) gave output identical, byte for byte,
to grammar v0.4.4's, and `go test ./...` passed. It was 4–6% slower
(38.4–39.0 s against 36.7–36.9 s), as each figure's attempt builds the
match so far for the check; grammar's reference has it among its
Engineering gaps. Whether one copy of the rule is worth that is the
lab's call.

**Steps,** once the release is out:

1. `go get github.com/nukopian/grammar@` that release, and apply:

```diff
diff --git a/internal/facts/facts.go b/internal/facts/facts.go
index 380a834..744b08a 100644
--- a/internal/facts/facts.go
+++ b/internal/facts/facts.go
@@ -97,15 +97,13 @@ func fact(kind string, m *scalar.SV, values map[string]string) Fact {
 	return Fact{Kind: kind, Text: strings.TrimSpace(m.At("~").Str()), Values: values, From: int(m.At("from").Int())}
 }
 
-// payOf reads a pay match, if it's pay: it needs a currency or a period.
+// payOf reads a pay match. The grammar has made sure it's pay, with a
+// currency or a period.
 func payOf(m *scalar.SV) (Fact, bool) {
 	// cur is captured twice in the rule, before each amount, so it's an
 	// array, empty when there's none; period is one match, or missing.
 	curs := m.At("cur").AV().Items()
 	cur, period := len(curs) > 0, m.At("period").Defined()
-	if !cur && !period {
-		return Fact{}, false
-	}
 	v := map[string]string{"min": amount(m.At("min"))}
 	if m.At("max").Defined() {
 		v["max"] = amount(m.At("max"))
diff --git a/internal/grammars/grammars.go b/internal/grammars/grammars.go
index 3887850..a9766f6 100644
--- a/internal/grammars/grammars.go
+++ b/internal/grammars/grammars.go
@@ -11,6 +11,7 @@ import (
 	"unicode"
 
 	"github.com/nukopian/grammar"
+	"github.com/nukopian/scalar"
 )
 
 // Pay is a figure of money: with a currency sign, or a period, or both,
@@ -21,7 +22,7 @@ import (
 var Pay = grammar.MustCompile(`
 grammar Pay {
     token TOP    { [ $<upto>=[ :i up \s+ to ] \s+ ]? <cur>? \s? <min=amount>
-                   [ <.sep> <cur>? \s? <max=amount> ]? <period>? <!before <.bonus>> }
+                   [ <.sep> <cur>? \s? <max=amount> ]? <period>? <?{ pay($/) }> <!before <.bonus>> }
     # A bonus isn't pay: "£30 bonus for cover", "£1,500 joining bonus";
     # but pay "plus bonus" or "with bonus" is pay, with a bonus besides.
     token bonus  { \s+ [ :i bonus >>
@@ -37,7 +38,14 @@ grammar Pay {
                             | per \s+ day | a \s+ day | daily
                             | pw | 'p/w' | per \s+ week | a \s+ week | weekly
                             | pcm | per \s+ month | a \s+ month | monthly ] >> ] }
-}`)
+}`, pay)
+
+// pay says whether a match of Pay so far is pay: it has a currency sign
+// or a period. cur is captured twice in the rule, before each amount, so
+// it's an array, empty when there's none; period is one match, or missing.
+var pay = grammar.Assertion("pay", func(m *scalar.SV) bool {
+	return len(m.At("cur").AV().Items()) > 0 || m.At("period").Defined()
+})
 
 // Hours is hours of work: "40hrs pw", "37.5 hours per week", "12 hours
 // per day", "40 hours". A bare "h" is hours only with a period after it,
diff --git a/internal/headings/headings.go b/internal/headings/headings.go
index 4688533..7940060 100644
--- a/internal/headings/headings.go
+++ b/internal/headings/headings.go
@@ -402,9 +402,7 @@ func onlyFacts(text string) bool {
 	cut(grammars.Pattern, nil)
 	cut(grammars.Schedule, nil)
 	cut(grammars.Hours, nil)
-	cut(grammars.Pay, func(m *scalar.SV) bool {
-		return len(m.At("cur").AV().Items()) > 0 || m.At("period").Defined()
-	})
+	cut(grammars.Pay, nil)
 	if string(left) == text {
 		return false // no facts at all
 	}
```

2. `payOf` no longer needs its bool; drop it, or keep it for the
   labelled-pay path, as reads best.
3. Update PIPELINE.md's Facts section if it says where the check is.

**Check:** `scripts/test-all.sh` passes, and chunking every ad gives
output identical to before the change.
