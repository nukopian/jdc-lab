# Map what's left

**Status:** in progress

**Origin:** roadmap 21

**Started:** 2026-10-28, in 7dbc459e5efbfa5891397f2b2510ee61bd316c67

## The plan

**Serves:** structure, embeddings

**From:** jdc-lab's session, 2026-10-28 (jdc-lab 7dbc459e5efbfa5891397f2b2510ee61bd316c67)
**Waiting on:** 22 for step 2, as a heading made only of facts is
told by its grammars; 17 before the final figures, as it moves text
between sections

**What:** give a kind to the headings of no known kind that hold the
most text. Then place the text that's still unmapped by an educated
guess, from what it says, instead of leaving all of it out.

**Why:** on the three samples, 12% of the text sits under headings of
no known kind, in 44% of ads. The chunker gathers that text into the
boilerplate chunk, so it isn't embedded: 10.7 points of an ad's tokens,
on the figures when 18 closed. Some of it is duties and requirements,
which search needs. Some of it really is notices and the advertiser's
pitch. At the moment all of it is treated as the second.

**What's known:**

- **It's a long tail.** The three samples have 10,639 distinct
  unmapped headings. Ranked by the text under them, the top 100 hold
  22% of the unmapped text, the top 200 29%, the top 500 41%, and the
  top 1,000 52%. Mapping headings one by one gives less each time, so
  this item maps the heavy end and guesses at the rest.
- **The heavy end** includes:
  - notices ("Please note:", "Important Information")
  - details ("Additional Information:", "Reporting To:", "OTE:",
    "Overtime:")
  - headings made only of facts ("Permanent", "Full-Time | Permanent")
  - requirements in a form `headings.txt` doesn't yet have ("The
    successful Teaching Assistant will have:")
  - the placeholders an export puts in for removed contact details
    ("(phone number removed)", "(url removed)"), which aren't headings
  - advertisers' names used as headings
- **Advertisers' names can't be mapped by name.** `headings.txt` is
  public, so they need a general rule or are left to the guess.
- **What 21 was thought to cover, but doesn't:** two of the pages' ads
  done badly were put down to headings. In one, "Why …?" takes in the
  pay, but that's text out of order, which 17 fixes. The other has no
  headings after its first, only a second list with no heading of its
  own whose lines read as requirements. Splitting a list like that by
  its wording is a different change.
- **The evidence a guess can use** is already in the lab:
  - the openings that place a line's wording as duties or requirements
    (`internal/wording`)
  - the boilerplate index, which says whether a text is repeated
    across ads
  - the heading's own words ("we", "our", "us" for the advertiser)
  - its place in the ad, as after the apply and legal text at the end

**Steps:**

1. Not headings: the removed-details placeholders, and anything else
   the heavy end shows that only looks like a heading.
2. Headings made only of facts become details, told by 22's grammars.
   Labels such as "OTE" and "Reporting To" join the field labels.
3. Map the unmapped headings from the top, by the text under them,
   through the top 200, each with the kind a reader would give it. A
   rule goes into `headings.txt` as a phrase that will reach others
   ("will have" with requirements, say), not as the heading word for
   word. Advertisers' names get a general rule, if one holds, or are
   left.
4. The guess: for each section still unmapped, weigh the evidence above
   and place it, as a kind or as boilerplate. With no evidence either
   way it's left out, as now. The chunker says what placed it, as it
   does for every chunk ("guessed duties: 5 of its 7 lines open as
   duties do").
5. Save runs before and after each step and compare them. Read 50 of
   the headings newly mapped, and 50 of the sections guessed, as the
   chunks show them.
6. PIPELINE.md: the guess, and the figures that move. README: Where it
   stands. Update the pages' notes in the data store.

**Done when** the text under headings of no known kind is below 9% on
the three samples, from 12%; of the 50 headings newly mapped read by
eye, none has a kind a reader wouldn't give it; of the 50 sections
guessed, at least 45 are placed as a reader would place them; and the
chunker shows what placed every guess.

## What happened
