# Split walls of text into sentences

**Status:** in progress

**Origin:** roadmap 5

**Started:** 2026-10-30, in 051e5fee0a24f3d03609757e21679944f75a0ede

## The plan

**Serves:** structure

**From:** jdc-lab's session, 2026-10-25 (jdc-lab 167f3607190f0a42fe04c3131c4b96545d76c36c)
**Waiting on:** nothing

**What:** split long paragraphs into sentences, and find the headings
and bullets run into them.

**Why:** some ads are a single wall of text, with headings run into the
prose ("… every day. What you'll be doing You'll be …") and bullets
flattened into capitalised phrases with no stops. Blocks keeps a wall as
one paragraph, so nothing in it is found as a heading, and boilerplate
sentences inside it aren't found either.

**What's known:** 80 ads have a paragraph over 1,500 characters. A
heading run into text follows the end of a sentence and comes before a
capital letter. Flattened bullets are harder: capitalised phrases with
no full stop between them.

**Steps:**

1. Agree the names first: a package for sentences, and what its pieces
   are called.
2. Split paragraphs into sentences with rules (stops, question marks,
   and the capital that follows), keeping each sentence's text as
   written.
3. Find headings inside walls: a known heading phrase (from
   `headings.txt`) between a sentence's end and a capital letter.
4. Let boilerplate match sentences as well as blocks.

**Done when** the walls are measured before and after (headings found in
them, boilerplate found in them), tests cover the splitting with
made-up text in the shapes seen, and PIPELINE.md has the new stage.

## What happened

- **Fewer walls than the plan says.** Its 80 ads with a paragraph over
  1,500 characters came from the first look at the sample, before
  blocks read a plain-text ad by its newlines. Counted again from
  blocks' output, 3 of the first sample's ads have one (22 over 800
  characters), 2 of the second's (59 over 800) and 12 of the third's
  (300 over 800). Walls are rarer than planned for, but the longest
  still hold headings run into them.
- **Many walls are folded layout, so the item waits on roadmap 24.**
  Of the six paragraphs with a heading run in after a sentence's end,
  five have it so in the ad's HTML. But looking for them showed that
  blocks folds an ad's newlines into spaces whenever the ad has any
  layout tag, as a browser would, and about 869 ads in the samples have
  newlines that look like layout folded so, headings and fields with
  them. Three ads have headings laid out by newlines and folded into a
  wall: one is a plain-text ad with a tagged footer, which makes blocks
  read all of it as HTML. Splitting sentences would patch that at the
  wrong stage, so keeping the layout became roadmap 24, and this item
  went back to Next, to be measured again on what's left once 24 lands.
  Nothing of it was built.
- **Back in Now, roadmap 24 shipped.** With the ads' layout kept, the
  walls are counted again: ads with a paragraph over 800 characters
  381 to 375, over 1,500 17 to 16. Roadmap 24 recovered the headings
  and fields folded into the text, but few whole walls were folded
  layout: what's left is prose, which this item's plan is for. Step 1,
  the names, is next.
- **Names agreed (step 1)**, after reading real walls with Iain:
  - a **wall** is a paragraph or line over 400 characters: over the
    three samples, about 11,000, among them the 487 that mix sentences
    repeated across ads with their own, and all five with a heading run
    in (466 to 1,619 characters);
  - a **sentence** runs up to a full stop, question mark or exclamation
    mark followed by a capital, or to the wall's end, kept as written;
  - a **run-in heading** is a known heading phrase (`headings.txt`) at
    a sentence's start, before a capital: in the walls it comes after a
    sentence's end ("… completion. The Role Reporting to …") or opens
    one with no stop at all ("Required Attributes and Experience The
    ideal …"), so the plan's "between a sentence's end and a capital"
    becomes "at a sentence's start";
  - `internal/sentences`, with `sentences.Split(text) []string`, pure
    text; and `internal/walls`, with `walls.Split([]blocks.Block)
    []blocks.Block`, a stage between blocks and headings, as headings
    can't import sections, where `headings.txt` is read;
  - a wall's sentences become Lines, as a `<br>` makes, so later stages
    need learn nothing new, and a run-in heading a Line of its own with
    the cue `run_in`. A wall is then printed a sentence to a line.
- **A second sign of a run-in heading** (Iain): a capital "The", "A"
  or "An" with no stop before it starts a sentence, and the title-case
  run before it, back to the last sentence's end, is a title over what
  follows ("Required Attributes and Experience The ideal Commercial
  Manager will have …"), whether or not `headings.txt` knows it.
  Counted in the walls: on its own the capital article misleads, as 78
  of them, in 72 walls, are mid-sentence names ("We are The …"); with a
  title-case run of up to six words before it, not opening with a word
  such as "As" or "At" ("As A Support Worker you will …" is 23 of 36),
  about 8 walls have one, besides the 5 with a known phrase. So a
  run-in heading is either a known heading phrase at a sentence's
  start, before a capital, or such a run before a capital article.
- **Seen in the walls, beyond this plan:** a prose wall's sentences run
  from company to duties ("Duties will include …") to requirements
  ("You will have …") to pay, and `wording`'s openings could place each
  once they're lines. Some of that may follow from step 4 on its own;
  what doesn't is for an idea after this item.
- **Sentences (step 2).** `internal/sentences` splits text at a full
  stop, question mark or exclamation mark followed by a space and a
  capital, kept as written, joined by spaces giving the text back. A
  stop before a lower-case word or a figure ends nothing ("example.
  com"), nor one after an abbreviation or an initial ("e.g. Sage", "J.
  Smith"); further stops and a closing quote or bracket stay with their
  sentence. Over the 10,874 walls in the samples, it makes 35,239
  sentences, about 3.2 a wall; a sample of 30 of its boundaries, read by
  eye, were all right, and the only stops before a capital it kept
  inside a sentence were 4 after "e.g.". Nothing calls it yet: the
  walls stage (step 3) will.
