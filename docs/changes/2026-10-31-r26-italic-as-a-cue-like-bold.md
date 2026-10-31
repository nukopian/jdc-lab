# Italic as a cue, like bold

**Status:** shipped, unreleased

**Origin:** roadmap 26

**Started:** 2026-10-31, in 8167591a685e4cb58368d727f16f52b0169b0ccc

**Finished:** 2026-10-31

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

- **Outcome:** done against Done when. blocks records `italic` beside
  `bold`, for `<em>` and `<i>`, ended with a `<p>` as bold is; walls
  gives it to each line of a wall; headings takes it as a cue where it
  takes bold; shape and `cmd/blocks` report it. Tests cover both tags,
  all italic and part, a `<p>` left open, bold and italic together, and
  the headings and sections rules below, with made-up text.
- **Changes from the plan,** both found by the compare:
  - *A bold lead wins over italic.* An ad set each of its lines in
    italic, with a bold label: "***Directorate:** Transformation*".
    Read as a cue, the italic made the whole line one heading, where the
    bold lead had made "Directorate:" a heading and the rest text. So an
    italic line with a bold lead isn't set apart: 3 headings were lost
    this way, and none is now.
  - *Italic alone is a weaker cue than bold.* As planned, headings
    gained 265 heading uses, but in sections only 131 were of a known
    kind. The rest were mostly not headings: job-board codes
    ("INDTPSUR", "#LI-HYBRID"), recruiters' signatures, slogans, notes
    ("Hybrid/Flexible working options available") and straplines, each
    starting a section of no known kind and taking the text after it.
    So headings marks a heading that only italic set apart (`Part`'s
    `Italic`), and sections reads one of no known kind as text, unless
    a list follows, as it reads the ad's subtitle (`sections.Demoted`,
    which `cmd/sections` counts by too). That keeps the 131, and 18 of
    no known kind over a list, sub-headings such as "Cost Analysis and
    Commercial Support"; 116 are text. Some real headings go with them
    ("Want to know more?", "Week 1"), as headings errs towards text.
- **Measurements:** runs `claude/before` (at 76f7e24) and
  `claude/after`, compared over the three samples. blocks: 4,723
  paragraphs, lines and items in italic. heading-counts: headings
  92,445 to 92,710, text 682,177 to 681,914, fields 38,570 to 38,569,
  ads with no heading 2,996 to 2,985. sections: heading uses 91,691 to
  91,840, of a known kind 74,646 to 74,777 (81%); lines starting a
  section by name 19,121 to 19,070, as some are now headings; ads with
  every heading known 15,965 to 15,961 (61.5% both, but rounding to 61%
  now, where it was 62%); 14 headings newly unmapped, each over a list.
  boilerplate and view unchanged. The field lost is "Hybrid working -",
  a heading after the title joined with the line after it, which is now
  an italic strapline read as a heading of kind `apply`, as "Application
  Support | SQL Databases | …" has a kind's word in it. Chunked
  (`chunk.AdOf` over the 25,954 distinct ads, each run's boilerplate,
  by a throwaway program, as `cmd/chunk` finds each ad by reading the
  whole file), embedded tokens go from 67.74% to 67.73%, and metadata
  stays 3.06%: PIPELINE.md's chunk figures stand.
- **What's left:** codes and straplines are headings in bold too, about
  450 and 400 ads in the samples, and a strapline with a kind's word in
  it starts that kind's section: idea
  [42](../../ROADMAP.md#42-codes-and-straplines-arent-headings), with a
  row in PIPELINE.md's Known gaps. An italic lead, as a bold lead is
  read, isn't recorded; nothing in the compare asked for one.
