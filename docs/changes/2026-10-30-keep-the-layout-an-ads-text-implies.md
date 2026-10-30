# Keep the layout an ad's text implies

**Status:** in progress

**Origin:** roadmap 24

**Started:** 2026-10-30, in 727275d1ac905c0199fcb5726802ab51ae5afb0a

## The plan

**Serves:** structure

**From:** jdc-lab's session, 2026-10-30 (jdc-lab 727275d1ac905c0199fcb5726802ab51ae5afb0a)
**Waiting on:** nothing

**What:** keep the layout an ad's author meant, where blocks now drops
it: newlines in ads that also have tags, breaks inside list items, blank
lines, and tags left open.

**Why:** blocks reads an ad with any tag that lays out blocks exactly as
a browser shows it, and a browser shows a newline as a space. Many ads
are written as if their newlines would show: bare lines of text between
tags, or a plain-text ad with one tagged footer, which makes the whole
ad read as HTML. Their headings and fields are folded into the text
around them, so the headings stage never sees them ("Key
Responsibilities HR Administration & Employee Lifecycle" is one block),
and they make many of roadmap 5's walls. Rare shapes still matter: each
is an ad misread wherever it's posted.

**What's known:** across the three samples' 28,000 ads, all but one
with tags that lay out blocks, counted with the same HTML tokenizer
blocks uses:

| What blocks drops | Ads |
|---|---|
| Newlines in a tagged ad that look like layout: after a stop or colon, before a bullet, or around a short line starting with a capital | about 869 (the rule counts some wrapped sentences too) |
| Newlines in a tagged ad, all mid-sentence: wrapping only | 12 |
| A `<br>` between text inside a `<li>` | 235 |
| A `<p>` or `<div>` after text inside a `<li>` | 45 |
| A blank line from `<br><br>`: the gap, not the lines | 10,814 |
| A `<b>` or `<strong>` left open when a `<p>` opens: bold leaks into the next block | 4 |

The samples close their tags almost always, and use only `<p>`, `<br>`,
lists and bold, but ads from elsewhere may not. Made-up HTML shows
three more ways blocks misreads tags left open: a `<p>` or heading tag
after an item of a list never closed is swallowed by the item; a heading
tag never closed loses its heading cue, and turns every later `<br>`
into a space; and block-level tags blocks doesn't know, such as
`<aside>` or `<center>`, join the text either side with no space.
Blocks already ends a block at an opening `<p>`, `<div>` or other tag
it knows, closed or not, and a list item at the next `<li>`.

**Steps:**

1. Save a run first (`pipeline run claude/before`), and agree any new
   names (a cue for a blank line, say) before they reach the API.
2. Newlines: in a tagged ad, a newline in text outside a list item ends
   a line, and a blank line a block, as in a plain-text ad. Check the
   wrapping first: if newlines mid-sentence turn out common, only those
   that look like layout end a line.
3. List items: a `<br>` or a `<p>` between text inside a `<li>` keeps
   the item's lines apart, rather than running them together.
4. Blank lines: a blank line from `<br><br>`, or a newline left blank,
   is recorded as a cue, for later stages to use; blocks decides
   nothing from it.
5. Tags left open: an opening block or heading tag ends a heading tag
   still open, and the item of a list never closed, unless the item
   holds it; bold ends at a block's end; and `breaks` covers HTML's other
   block-level tags.
6. Measure each step with `pipeline run claude/after` and
   `pipeline compare claude/before claude/after`: headings found,
   unmapped text, and walls.

**Done when** each shape above has a test with made-up HTML, the runs
before and after show what moved (headings found, unmapped text, the
walls left), the ads above read as their authors laid them out,
PIPELINE.md's blocks stage says how layout is read, and roadmap 5 is
re-measured on what's left.

## What happened

Nothing yet.
