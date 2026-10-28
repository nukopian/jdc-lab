# Make boilerplate's verdict right

**Status:** in progress

**Origin:** roadmap 4

**Started:** 2026-10-28, in e6226efeb90802722c14b59348f81a525fc3abca

## The plan

**Serves:** structure, embeddings

**From:** jdc-lab's session, 2026-10-25 (jdc-lab 167f3607190f0a42fe04c3131c4b96545d76c36c), widened 2026-10-27
**Waiting on:** nothing

**What:** make the verdict on repeated text right both ways: footers
filed under requirements counted as boilerplate, and common job wording
no longer counted as it.

**Why:** footers (an agency's notices, equal opportunities statements)
often have no heading, so they're filed under the last one, usually
requirements, and under five title groups they're counted as templates
for the job and embedded. The other way, text under five title groups
or more is boilerplate wherever it is, so job lines common to many ads
("High levels of accuracy and attention to detail") are taken out of
what's embedded: about half of a sample of what's taken from summaries
and requirements is about the job.

**What's known:** the largest templates in the first sample's
`boilerplate/repeated.tsv` are almost all such footers. `chunk`'s
boilerplate chunk lists, ad by ad, every line the verdict takes from an
embedded chunk, with the section it came from: where to check it.

**Steps:**

1. Record where each text comes in its ad: whether it's in the run of
   repeated texts that ends the ad; count a repeated text found mostly
   in such end runs as boilerplate, whatever its section.
2. Keep a widespread text that's found mostly in the sections that
   describe the job, and is phrased as a requirement or a duty, out of
   boilerplate.
3. Check the templates under three or more groups, and a sample of the
   boilerplate chunks, before and after.

**Done when** those footers are boilerplate, common job lines aren't, a
sample of the newly flipped texts is checked by eye both ways, tests
cover it, and PIPELINE.md's figures are updated.

## What happened
