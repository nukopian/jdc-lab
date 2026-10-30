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
