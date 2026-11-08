# TODO

## Drop MayHold's checks for a digit, now grammar rejects such lines itself

**From:** grammar's session, 2026-11-08 (grammar 178b67175959fd6179df36af3b9e71bbb16a7bf2)
**Waiting on:** a grammar release with roadmap 23 (the release after v0.5.4)

**What:** `grammars.MayHold` checks a line for a digit before
`facts.Find` runs the Pay, Hours, Postcode and Figure grammars. grammar's
roadmap 23 does that check itself: `Find` and `FindAll` work out the
characters every match needs, a digit for each of these four, and reject
a line with none before reading it. Drop those four checks from
`MayHold`; keep the ones for Schedule and Pattern, which it can't yet do
as well (a word list needs only its phrases' first letters).

**What grammar's session measured,** with `BenchmarkFindSamples` over
every line of 2,000 ads, 64,244 lines, with grammar's working copy through
`go.work`:

- 497 ms with grammar v0.5.3 and `MayHold` as it is;
- 466 ms with roadmap 23 and `MayHold` as it is;
- 477 ms with roadmap 23 and `MayHold` only for Schedule and Pattern;
- 673 ms with roadmap 23 and no `MayHold`.

Facts were the same, byte for byte, as JSON, for every line, with v0.5.3,
with roadmap 23, and with roadmap 23 and no `MayHold`.

So dropping the four checks costs about 2% against keeping them, and
still leaves jdc-lab about 4% faster than with v0.5.3. Whether that's
worth the simpler code is Iain's call; keeping them is fine too.

**Steps:**

1. Take up the grammar release in `go.mod`, and check with `GOWORK=off`.
2. In `internal/grammars/grammars.go`, drop `Pay, Hours, Postcode, Figure`
   from `MayHold`'s switch, so it returns true for them, and update its
   doc comment; `hasDigit` may still be needed by Schedule's case.
3. Update `TestMayHold` in `internal/grammars/grammars_test.go`, whose
   cases expect false for those grammars on lines with no digit.
4. Run `TestSkipLinesChangeNothing` (not `-short`, so it reads the run's
   2,000 ads) and `BenchmarkFindSamples`, and compare with the figures
   above.

**Done when** `MayHold` checks only Schedule and Pattern, the tests
pass, and the facts are unchanged.
