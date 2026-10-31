# Italic as a cue, like bold

**Status:** in progress

**Origin:** roadmap 26

**Started:** 2026-10-31, in 8167591a685e4cb58368d727f16f52b0169b0ccc

## The plan

**Serves:** structure

**Size:** XS (1)

**What:** record a paragraph all in `<em>` or `<i>`, as blocks records
one all in bold, so headings read italic as a cue too.

**Why:** ads set headings ("You'll also need to:") and notices in
italic, and without a cue those are read as text: 470 ads in the
samples have one, so it reaches many ads for little work.

**What's known:** blocks counts the `<strong>` and `<b>` it's inside,
and a piece of text whose visible text was all inside one gets
`Cues.Bold`. A bold `<strong>` opened inside a `<p>` ends with it, as
ads leave one open. walls gives each line split from a wall the wall's
`Bold`. headings takes `Bold`, as it takes a heading tag, in two
places: to let a heading ending in ":" be longer (`MaxIntroWords`), and
as the cue that makes a short, heading-shaped line a heading. shape and
`cmd/blocks` report the cue. A bold lead ("**Salary:** £30,000") has
no italic counterpart in this plan.

**Steps:**

1. Save a run first: `go run ./cmd/pipeline run claude/before`.
2. blocks: count the `<em>` and `<i>` it's inside as it counts bold,
   ended with a `<p>` as bold is, and set a new `Cues.Italic` on a
   piece whose visible text was all italic.
3. walls: give split lines the wall's `Italic`, as its `Bold`.
4. headings: take `Italic` wherever it takes `Bold`.
5. shape and `cmd/blocks`: report the cue.
6. Tests with made-up text: `<em>` and `<i>`, all italic and part, a
   `<p>` left with its `<em>` open, and headings reading an italic
   line as it reads a bold one.
7. `go run ./cmd/pipeline run claude/after`, then
   `go run ./cmd/pipeline compare claude/before claude/after`, and look
   at the headings gained.
8. PIPELINE.md's blocks and headings stages say so, its Known gaps row
   goes, and the README's Where it stands and the CHANGELOG are updated.

**Done when** blocks records italic beside bold; headings takes it as a
cue where it takes bold; tests cover both tags with made-up text; a run
against the one before shows which headings were gained, and nothing
else moved; and PIPELINE.md says so, its Known gaps row gone.

## What happened
