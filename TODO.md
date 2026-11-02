# To do

## Name the captures Pay's and Hours' assertions read

**From:** grammar's session, 2026-11-02 (grammar f827a363542da63d9712a8ab24b85614092d99e8)
**Waiting on:** a grammar release after v0.5.0 with named reads (its roadmap 24)

**What:** pass the captures each code assertion reads to
`grammar.Assertion`, after the function, so grammar builds only those
for the match so far.

**Why:** each assertion builds every capture of the match so far at each
figure. Naming the reads made the chunker about 2% faster over the
28,000 ads of the latest run (37.05–37.69 s against 38.01–38.36 s, three
runs each, interleaved), with identical output, byte for byte. Reading a
capture that isn't named then panics, saying which assertion and what it
names, so a function changed to read more is caught at once.

**Files:** `internal/grammars/grammars.go`. The change:

```diff
-})
+}, "cur", "period")
```

at the end of `pay`, and

```diff
-})
+}, "num", "per")
```

at the end of `implausible`.

**Steps:**

1. `go get github.com/nukopian/grammar@` the release, and make the change.
2. Rerun the chunker over every ad, against the output before, and run
   the tests.

**Done when** the output is unchanged and the tests pass.
