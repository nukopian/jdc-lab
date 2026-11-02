# Checks in the fact grammars, with code assertions

**Status:** shipped, unreleased

**Origin:** two TODO tasks from grammar's session (grammar
24b190ab4ab18da57055eb91a43627f59d8a2d23 and
450a8507cab32d1f2fe2ad9c530120bbed479fe1)

**Started:** 2026-11-02, in e94d4963779aed4a1943437a1c114a94078efe61

**Finished:** 2026-11-02

## The plan

Two tasks, taken up together, as both wait on the same grammar release
and change the same file.

### Check pay in the Pay grammar, with a code assertion

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

### Drop implausible hours in the Hours grammar, with a code assertion

**From:** grammar's session, 2026-11-02 (grammar 450a8507cab32d1f2fe2ad9c530120bbed479fe1)
**Waiting on:** a grammar release with code assertions (after v0.4.4)

**What:** Hours reads any figure before "hours" as hours, so the latest
run's facts include "272 hours" (annual leave), "Dec 2026 Hours : 37.5
hours per week" (a year), "1800-0600hrs" (clock times), "168 hours per
month" and "84 Hours per Fortnight" (periods Hours doesn't know, read as
bare hours). A code assertion, `<!{ implausible($/) }>`, drops a match
with more hours than its period holds: over 24 a day, over 168 a week,
or, with no period, over 80.

**Why:** these are misreadings, and a line holding one counts as a line
of facts: "- 224 hours holiday" in four care-home ads was taken out of
the summary into a details chunk.

**What's known:** tried against grammar's checkout on 2026-11-02 with
the patch below. Over all 28,000 ads, 76 hours facts go (4,729 to
4,653), every one a bare figure over 80, 33 readings in all, and
nothing is added; the only other change is those four ads' holiday
lines staying in their summaries. `go test ./internal/...` passed.
The left boundary, `<!after <[\d.]>>`, is needed: without it, a
rejected "224 hours" leaves the search to find "24 hours" inside it,
as "2026 Hours" gave "026 Hours". The threshold of 80 for a bare figure
is a guess at a long week; the lab may prefer another, or periods for
months and fortnights instead.

**Steps,** once the release is out (with the Pay task, the scalar
import is added once):

```diff
diff --git a/internal/grammars/grammars.go b/internal/grammars/grammars.go
index 3887850..60b92d3 100644
--- a/internal/grammars/grammars.go
+++ b/internal/grammars/grammars.go
@@ -11,6 +11,7 @@ import (
 	"unicode"
 
 	"github.com/nukopian/grammar"
+	"github.com/nukopian/scalar"
 )
 
 // Pay is a figure of money: with a currency sign, or a period, or both,
@@ -44,12 +45,27 @@ grammar Pay {
 // as "24h support" isn't.
 var Hours = grammar.MustCompile(`
 grammar Hours {
-    token TOP  { <num> \s* :i [ [ hours | hrs | hr ] >> [ \s* <per> ]? | h >> \s* <per> ] }
+    token TOP  { <!after <[\d.]>> <num> \s* :i [ [ hours | hrs | hr ] >> [ \s* <per> ]? | h >> \s* <per> ] <!{ implausible($/) }> }
     token num  { \d+ [ '.' \d+ ]? }
     token per  { <week> | <day> }
     token week { :i [ per \s+ week | pw | 'p/w' | a \s+ week | weekly ] >> }
     token day  { :i [ per \s+ day | a \s+ day | daily | per \s+ shift ] >> }
-}`)
+}`, implausible)
+
+// implausible says whether a match of Hours is more hours than its period
+// holds: over 24 a day, over 168 a week, or, with no period, over 80, a
+// long week's, as "272 hours" of annual leave, "2026 Hours" after a date
+// or "0600hrs" are.
+var implausible = grammar.Assertion("implausible", func(m *scalar.SV) bool {
+	n := m.At("num.~").Num()
+	switch {
+	case m.At("per.day").Defined():
+		return n > 24
+	case m.At("per.week").Defined():
+		return n > 168
+	}
+	return n > 80
+})
 
 // patternRules are the work patterns and arrangements, each rule its
 // canonical value first, then what ads write for it.
```

**Check:** `scripts/test-all.sh` passes; chunking every ad drops only
hours facts over the limits, and adds none.

## What happened

- **Outcome:** both done, with grammar v0.5.0 and scalar v0.6.3, which
  it needs. `scripts/test-all.sh` passes. Chunking all 28,000 ads of
  the latest run with Pay's check in its grammar gave output identical,
  byte for byte, to grammar v0.4.4's. Adding Hours' check, 76 hours
  facts went (4,729 to 4,653), 33 readings, every one a bare figure
  over 80, and no fact was added or changed; the only other change was
  four ads' holiday lines staying in their summaries, as the task
  found. Every stage's output is unchanged (`pipeline compare
  claude/before claude/after`).
- **Changes from the plan:** `payOf` lost its bool, as nothing calls it
  for labelled pay; `onlyFacts`'s `cut` lost its `keep` argument, which
  no grammar needed any more, and headings no longer imports scalar. A
  test of the misreadings the Hours task names, and of its limits,
  fails before the change and passes after.
- **Measurements,** on 2026-11-02, with six runs of each benchmark:
  `BenchmarkFindSamples` 1.14 s to 1.22 s (+7.6%), and
  `BenchmarkAdOfSamples` 1.46 s to 1.58 s (+8.5%), with 8% more
  allocations. Chunking every ad took 37.7 s, 39.0 s with Pay's check,
  and 39.2 s with both: 4% in all, about 0.06 ms an ad. Iain took the
  cost for one copy of each rule. Profiling grammar's `BenchmarkAssert`
  showed why: each assertion builds the match so far, a hash for every
  capture, 39% of a search's allocations, though Pay's function reads
  two captures.
- **Found in the release's wash-up:** `implausible` read the figure
  with scalar's `Num`, which reads ASCII digits only, so "٢٧٢ hours"
  passed as plausible. `parse` moved from facts to
  `grammars.Number`, which both now use; a test of such figures failed
  before and passes after, and chunks and every stage are unchanged.
- **What's left:** building only what an assertion reads, now
  grammar's roadmap 24 and scalar's 5, both at the top of Next. When
  it ships, take it up and measure again. The threshold of 80 for a
  bare figure is still a guess; "168 hours per month" and "84 Hours per
  Fortnight" are dropped, not read, as Hours has no month or
  fortnight.
