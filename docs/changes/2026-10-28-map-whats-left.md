# Map what's left

**Status:** shipped in v0.5.2

**Origin:** roadmap 21

**Started:** 2026-10-28, in cbfca304a81ae66e2f2d6cf5dcdfc4b8d8048603

**Finished:** 2026-10-28

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

- **The list with no heading** (2026-10-28), named in What's known as
  a different change, is idea 23.
- **Outcome:** shipped short of its Done when, on both measures, by the
  maintainer's choice. Text under headings of no known kind fell from
  11.6% to 9.8% on the three samples, not below 9%. Of 50 headings
  newly mapped, read by eye, none had a wrong kind (after a first 50
  found four, and three phrases were made narrower). The guess was
  right for 39 to 43 sections of each fresh 50, not 45. The chunker
  shows what placed every guess.
- **Why the targets weren't reached:** the heavy end of what's left is
  advertisers' names, which `headings.txt`, being public, can't hold,
  and catch-alls ("Please note:", "Additional Information:") whose text
  is of several kinds, so no phrase can map it. The guess's misses are
  a long tail of the employer's culture, perks and notices under
  headings no rule covers.
- **Changes from the plan:**
  - Step 1 and 2 went into `internal/headings` rather than `sections`:
    a bold line made only of facts (`onlyFacts`, by the shared
    grammars) and one holding a placeholder for removed contact
    details (`hasRemoved`) are no headings, but text, so their lines
    join the section before and the chunker takes them as facts. A
    line of facts ending in a colon still heads what follows. "37.50
    hours per week" alone is now text, where a test had it a heading.
  - New field labels: "reporting to", "responsible to", "reference
    number", "key words". New phrases in `headings.txt` for each
    kind, narrow enough to keep the 50 read by eye right: "candidate"
    and "setting" alone reached too far, and went.
  - The guess is in the chunker (`chunk.guess`), as it needs the
    boilerplate verdict and the wording rules, which `sections`
    can't import. So the sections' unmapped share counts guessed text
    as unmapped still; the chunks' figures show the guess.
  - The default: the plan left a section with no evidence out. A
    sample of 40 such sections found about 31 about the job (its
    pitch, duties and requirements in words `openings.txt` hasn't), so,
    with the maintainer's agreement, the default is the summary. A
    label over a short value ("IR35 Status:", "Inside IR35") is
    details, and headings opening "who are" or "join" speak for the
    advertiser, both added after the second sample of guesses.
- **Measurements**, the three samples, `claude/before` to `claude/after`:

  | | Before | After |
  |---|---:|---:|
  | Heading uses of a known kind | 78% | 81% |
  | Ads with every heading known | 56% | 62% |
  | Text under headings of no known kind | 11.6% | 9.8% |
  | Ads with an unmapped section | 43.6% | 38.5% |
  | Tokens embedded | 57.5% | 66.1% |
  | Tokens in the boilerplate chunk | 19.2% | 9.6% |
  | Of them, unmapped text | 10.7 | 2.9 |
  | Tokens placed by the guess | | 6.4 |

  Of the 2.9 points left out: 2.3 mostly boilerplate, 0.4 the
  advertiser's, 0.2 after applying. The guesses read: 42, 32, 43 and 39
  of 50, the second before the label and advertiser rules, the last
  after the final phrases and tuned on none.
- **Tried and reverted:** "candidate", "setting" and "overtime" as
  details, all reaching too far; the label rule before the
  after-applying one.
- **What's left:** the guess's long tail; `openings.txt` lacks the bare
  verbs ("Teach", "Provide") that would place more duties by wording
  than as summary; advertisers' names need a general rule, perhaps
  the boilerplate verdict on their heading across ads.
- **The wash-up before the release** (an independent review, each finding
  reproduced by a failing test, then fixed): a field's label that names
  a work pattern, "Hybrid Working" in bold, had become text, as
  `onlyFacts` took it for a line of facts, so its section ran into the
  one before; it's a heading again (16 heading uses). "We are seeking:"
  headed requirements, but its "we" had the guess leave them out as the
  advertiser's; it's now a requirements phrase, as "we're looking for"
  is (16 headings), and "someone who" a requirements opening. The
  guess's doc and PIPELINE's table now say the evidence is weighed in
  order, that "at least half" is meant where they said "most", and that
  a section of fields alone is left out; `hasRemoved` ignores case.
