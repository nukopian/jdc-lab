# Stop shift times passing as headings

**Status:** shipped, unreleased

**Origin:** roadmap 2

**Started:** 2026-10-25, in dedfbc7b228ede0069cd12c0393301288fe88f0f

**Finished:** 2026-10-27

## The plan

**Serves:** structure

**From:** jdc-lab's session, 2026-10-25 (jdc-lab 167f3607190f0a42fe04c3131c4b96545d76c36c)
**Waiting on:** nothing

**What:** don't take a bold time range, such as "2:00pm–10:00pm", for a
heading.

**Why:** it starts a section of no known kind, splitting the section
it's in.

**What's known:** `headingShaped` refuses money and runs of three
digits or more, but a time has only pairs. The headings with figures in
are listed by a command in PIPELINE.md's Improving it.

**Steps:** in `internal/headings`, refuse text whose key is mostly
times (digits with am or pm, or hh:mm); add cases to `TestRead`.

**Done when** no time range is in `data/runs/RUN/all/headings/headings.tsv`, and no
other heading was lost (compare the table before and after).

## What happened

The item was in Now from the roadmap's start, and its work was done on
2026-10-27, by then with three samples, 25,954 distinct ads in all.

- **Outcome:** done. `headingShaped` refuses text holding a time of
  day, so no time range is left in
  `data/runs/after/all/headings/headings.tsv`. The only figures with a
  point left there are pay and hours ("Pay Rate – 13.10", "37.50 hours
  per week"), which aren't times. No heading without a time was lost:
  the table lost 365 headings and gained none, and `compare` shows
  none newly unmapped.
- **Changes from the plan:** the rule refuses a heading holding any
  time, not one that's mostly times. Every heading in the samples with
  a time in is about when the work is: shift times ("- 6am to 2pm"),
  hours ("Monday to Friday, 8:00am - 5:00pm"), start times ("08:00
  starts"), and lines of facts that include one ("Full-time |
  Monday–Friday | 9am–5pm"). Measuring "mostly" would have kept those
  last. A time is a figure with am or pm ("2pm", "7.30am"), hh:mm
  ("17:00"), or a range of hh.mm ("8.30-3.30"); a figure with a point
  and nothing more ("37.50") is not, as it's more often pay or hours.
  `TestRead` has a case for times, and one for figures that aren't.
- **Measurements:** runs `before` and `after`, compared. Across all
  three samples, heading uses fell from 95,174 to 94,674: 500 uses of
  365 headings, 326 of them unmapped. The other 39 had been mapped to
  details, by a word such as "hours" ("Working Hours - Monday to
  Friday, 08:00 - 16:30"). They're one-line facts: the two looked at
  with `heading-uses` had nothing under them, and opened an empty
  section. Known-kind uses fell by 52, from 72,167 to 72,115.
  Ads with every heading known rose from 13,565 to 13,672 (52% to
  53%), and ads with an unmapped section fell from 48% to 47%. On the
  second sample, ads with every heading known rose from 51% to 52%,
  and PIPELINE's What it measured has the new figures.
- **Tried and reverted:** nothing.
- **What's left:** the 39 details lines would be better read as
  fields, a label and its value, but a dash isn't a field's separator,
  as a colon or bar is. Pay written without a currency sign ("Pay Rate
  – 13.10", in 25 ads) still passes as a heading. Neither is on the
  roadmap yet.
