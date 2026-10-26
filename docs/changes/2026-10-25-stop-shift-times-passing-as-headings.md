# Stop shift times passing as headings

**Status:** in progress

**Origin:** roadmap 2

**Started:** 2026-10-25, in dedfbc7b228ede0069cd12c0393301288fe88f0f

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

Nothing yet: the item has been in Now since the roadmap began, but no
work on it has been committed.
