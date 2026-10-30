# Keep the layout an ad's text implies

**Status:** shipped, unreleased

**Origin:** roadmap 24

**Started:** 2026-10-30, in 727275d1ac905c0199fcb5726802ab51ae5afb0a

**Finished:** 2026-10-30

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

- **Outcome.** Done when is met: each shape in What's known has a test
  with made-up HTML; the runs before and after are compared below; the
  ads looked at when planning read as their authors laid them out;
  PIPELINE.md's blocks stage says how newlines, list items, blank lines
  and tags left open are read; and roadmap 5's walls are counted again.
  The gain is mostly in blocks, not yet in headings: over the three
  samples, from `claude/before` to `claude/step5`, lines 146,756 to
  149,392, items 458,079 to 459,083, and 69,126 lines and paragraphs
  with a gap; but heading uses only 91,497 to 91,625, of a known kind
  74,467 to 74,585 (81% both), lines starting a section by name 19,026
  to 19,105, and ads with no heading 3,014 to 3,006. Walls hardly moved
  (ads with a paragraph over 800 characters 381 to 375, over 1,500 17
  to 16): what's left of them is mostly prose, for roadmap 5.
- **Names agreed (step 1).** Iain's choices: text after a `<br>` or
  `<p>` inside a list item leaves the item, each line a block of its
  own after the list, as blocks already does for a paragraph inside an
  item; and the cue for a blank line before a block is `gap`. The
  baseline is `claude/before`, the pipeline at
  6e589dbf47910ddb758821ac283050fb221b04a4.
- **Newlines (step 2), and wrapping checked first.** Outside list
  items, tagged ads have about 12,000 newlines before a lower-case
  letter, in 453 ads, nearly all a sentence wrapped in the source, and
  about 2,400 after a stop or before a capital or a symbol, in about
  1,000 ads, mostly layout. So a newline before a lower-case letter
  stays a space. Two refinements came from the runs:
  - **Wrapping before a capital.** The first run cut a heading in 28
    ads at a wrap before a capital ("Welcome to" / "Northern Repairs", a team named in a heading), so a newline
    after a comma, "&" or a word a line can't end with ("to", "and",
    "of", "the" and the like) stays a space too.
  - **A tag to a line.** The second run split "**Location:** Hampshire"
    in two, and sentences at each bold word, in ads whose HTML is
    formatted with a tag on each line. So in a tagged ad only a newline
    with text either side, in one run of text with no tag between,
    counts; plain text still counts every newline, across inline tags
    too.
- **Measurements, step 2** (`pipeline compare claude/before
  claude/after`): lines 146,756 to 149,236, paragraphs 259,921 to
  259,359; heading uses 91,497 to 91,539, of a known kind 74,467 to
  74,509 (81% both); lines starting a section by name 19,026 to 19,089.
  Two headings are newly unmapped, both real ones once run into the
  line before ("Process Development & Continuous Improvement:"). The six
  ads looked at when planning all read as laid out: one, a
  plain-text ad with a tagged footer, was one summary chunk and now has
  its summary, role, responsibilities and requirements. Walls: ads with
  a paragraph over 800 characters 381 to 375, over 1,500 17 to 16.
- **List items (step 3).** An item ends at a `<br>`, or a tag that ends
  a block, after its text, and the rest of the `<li>` is blocks of its
  own where it stands, as Iain chose; before the item's text, as in
  `<li><p>Drive</p></li>`, the tag is still only a space. This also
  fixes the first of the tags left open (step 5): a `<p>` after the
  last item of a list never closed no longer joins the item. The old
  test that a `<br>` in an item is a space now says the opposite,
  `Days<br>Nights` being two blocks: the price of the choice.
- **Measurements, step 3** (`pipeline compare claude/step2
  claude/after`, step 2's run kept as `claude/step2`): items 458,112
  to 459,083, lists 74,642 to 74,970 (an item's trailing text now ends
  its list); heading uses 91,539 to 91,625, of a known kind 74,509 to
  74,585; lines starting a section by name 19,089 to 19,105. Six
  headings are newly unmapped, all from benefits lists set as lines in
  one item, whose bold labels ("Retail discounts - Save at…") headings
  takes for headings: a question for headings, left for now. In one ad,
  the application text, footer and four fields written into the last
  benefit are lines again, and chunk files them as apply, boilerplate
  and details.
- **Blank lines (step 4).** The cue `gap` marks the line or paragraph
  after two `<br>`s with nothing visible between them, in one block, or
  after a blank line in text laid out by its newlines. A first version
  also marked a single `<br>` at a block's start, or after a list, which
  two old tests showed isn't a blank line in the sense meant, so it
  needs two in a row. The paragraph break every `<p>` has doesn't
  count, nor does a gap before a list's items. Over the three samples,
  69,126 lines and paragraphs have it (`pipeline compare claude/step3
  claude/after`), and nothing else moved, as nothing reads it yet.
- **Tags left open (step 5).** A `<p>`, list or heading tag ends a
  heading tag still open, keeping its cue; a heading tag after an
  item's text leaves the item, as a `<p>` does since step 3 (the old
  test that it stays in the item now says so); `breaks` covers HTML's
  other block-level tags. For bold, the plan's "bold ends at a block's
  end" was narrowed: a browser carries an unclosed `<b>` into the next
  paragraph, and one ad wraps whole paragraphs in a `<strong>`, so only
  bold opened inside a `<p>` ends with it. On the samples, which close
  heading tags and use no other block tags, only that moved: 5 fewer
  bold paragraphs, all footers after a forgotten `</strong>`, and one
  more bold lead (`pipeline compare claude/step4 claude/after`).
- **What's left.** Bold labels on benefit lines set in one item, now
  lines, are read as headings of no known kind (6 in the samples):
  [idea 28](../../ROADMAP.md#28-bold-labels-on-benefit-lines-arent-headings).
  The wall paragraphs left are
  [roadmap 5](2026-10-30-r5-split-walls-of-text-into-sentences.md)'s.
  Inline lists and italic stay ideas 25 and 26. Nothing reads `gap` yet.
