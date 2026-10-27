# Check the rules on a second sample

**Status:** shipped, unreleased

**Origin:** roadmap 1

**Started:** 2026-10-25, in dedfbc7b228ede0069cd12c0393301288fe88f0f

**Finished:** 2026-10-26

## The plan

**Serves:** structure

**From:** jdc-lab's session, 2026-10-25 (jdc-lab 167f3607190f0a42fe04c3131c4b96545d76c36c)
**Waiting on:** nothing

**What:** work down the second sample's unmapped table, and change any
rule that fitted only the first.

**Why:** every rule, phrase and threshold was tuned on the first
sample, which leans towards education jobs. The second leans towards
vehicle jobs.

**What's known:** both samples are in `data/ingestion/`, and their
figures in PIPELINE.md's What it measured. The rules mostly hold. Where
they don't: heading uses of a known kind fall from 76% to 73%, unmapped
text rises from 14% to 16%, and a quarter of that is boilerplate. The
biggest unmapped heading is "and" (39 ads, 82,000 characters), with
"OR" beside it: one advertiser's template puts them alone in bold
between requirements (Known gaps: lone conjunctions). Then phrases
that look like kinds: "who thrives here", "what to expect", "the
business", "what is in it for you", "vacancy type", "working
arrangement", "important notice". "The business" also maps "The
business has a strong presence across:", so check each in context.

**Steps:**

1. `go run ./cmd/pipeline run before`.
2. Stop a lone conjunction passing as a heading, in
   `internal/headings`, with a case in `TestRead`.
3. Work down `data/runs/before/ads-20261025/sections/unmapped.tsv` as
   PIPELINE.md's Improving it says, comparing each change on every
   sample.

**Done when** the second sample's known headings are at least the
first's 76%, neither sample's figures got worse, the lone conjunctions
gap is out of PIPELINE.md, and its figures are updated.

## What happened

- **Outcome:** done. Headings of a known kind on the second sample went
  from 73.4% of uses to 76.0% (12,212 of 16,062), past the first
  sample's 75.7% when the plan was written; the first rose to 76.7%
  (5,313 of 6,929). Unmapped text fell from 16% to 13% on the second,
  and 14% to 13% on the first. No heading of a known kind became
  unmapped on either. The lone conjunctions gap is out of PIPELINE.md,
  and its figures are updated. One figure fell, boilerplate on the
  second sample, from 13.3% to 12.5%: see Measurements, as it's partly
  a correction.
- **Changes from the plan:** step 3 found that a rule, not only
  phrases, had fitted the first sample: `MaxWords`. A bold line ending
  in a colon now has up to `MaxIntroWords` (20) words, as such lines
  introduce lists ("To be successful in this role as a Vehicle
  Technician, you should have:"); one advertiser's benefits list was
  being filed under requirements without it. "Either:" and "Plus:"
  joined "and" and "or" as conjunctions.
- **What changed,** each measured with `pipeline compare` on both
  samples, and each new heading's kind checked by eye:
  - `internal/headings`: a heading needs a word that isn't a
    conjunction (and, or, either, nor, but, plus); `MaxIntroWords`; the
    field labels "vacancy type", "working arrangement(s)", "hybrid
    working" and "reports to".
  - `headings.txt`: requirements "who thrives here", "what you
    need", "candidate(s) who", "teacher(s) who", "must have", "what we need", "qualities",
    "licence", "certification", "eligible"; benefits "what is in it for
    you", "what you can expect"; responsibilities "what to expect";
    company "the business", "our business", "the school"; apply
    "recruitment process"; summary "purpose of job", "main purpose".
    "Candidates who" also took eleven headings such as "We welcome
    applications from candidates who:" from apply to requirements,
    where they belong.
- **Measurements:** runs `before` and `now` (the same as `after`) on
  2026-10-26. Boilerplate on the second sample fell 0.8 points, and
  template rose from 2.6% to 3.3%: 35 repeated texts, 72,000
  characters, went from boilerplate to template, all of them from
  unmapped sections into known ones. About 41,000 characters of it is
  right: job text repeated across one advertiser's related jobs, under
  "What to expect" and "Who thrives here?", and other employers'
  requirements, counted as boilerplate before only because unmapped is
  outside the sections that describe a job. About 21,000 is wrong: the
  same advertiser's application line, and a development paragraph run
  into its licence requirement with no break in the HTML, now under
  requirements, with no heading of their own. On the first sample,
  boilerplate is unchanged at 13.2%.
- **A wrong turn, put right:** "what you need" was first left out, on
  the belief that it would also take "What you need to do now" (36
  ads, apply) to requirements, whose rule comes first. That was never
  tested, and it's wrong: where two phrases start at the same word, the
  longer is found, so "what you need to do now" is apply whatever the
  rule order. Added afterwards, "what you need" took 16 headings to
  requirements, all rightly, one of them from apply: "What you need to
  apply for the …:", which lists requirements.
- **Legal's name, put right too:** each rule's first word, the kind's
  name, was one of its phrases, so the word "legal" anywhere in a
  heading made it a legal notice. Four headings in both samples had it
  for another reason: three name the job ("The successful Legal
  Secretary will have:"), and one an agency. Matching against the ad's
  own title would have caught one, as the others word the job
  differently. Instead, legal's rule is now written `phrases => legal`,
  which the lab reads as leaving the name out (grammar didn't, so
  `internal/sections` built its list itself, until grammar v0.3.1's
  `MapOnly`, suggested from here, did it for the lab), with the only headings
  that rightly relied on the word, "legal information" and "the legal
  bit" (10 ads each), as phrases. Compared on both samples, exactly the
  four headings moved: one to responsibilities, one to company, two to
  unmapped, as no phrase fits "Temp Legal Secretary" or "The
  successful Legal Secretary will have:". Legal's boilerplate share
  rose a point on the first sample, 76% to 77%.
- **Tried and kept apart:** lines of job terms, such as
  "Permanent | Full-Time" (16 ads) as a heading, weren't made details:
  the text under them is a summary.
- **What's left:**
  - Text with no heading of its own, after the requirements, counts as
    template, as footers do
    ([roadmap 4](../../ROADMAP.md#4-catch-footers-filed-under-requirements)),
    though here it isn't at the end of the ad.
  - Phrases are found leftmost first, without overlapping (the longest
    where two start at the same word), so "To join
    us and be successful in this role you will need:" is
    responsibilities: "in this role you will" takes the words that "you
    will need" would have matched. Since put right, with grammar
    v0.3.2's `Outermost`, suggested from here
    ([record](2026-10-26-try-overlap-and-take-up-grammar-v0.3.2.md)).
  - "What to expect from our recruitment process:" (apply) and "What to
    Expect from …:" (an agency's pitch) are responsibilities, one ad
    each.
  - "The school" can head a job rather than the school, as in "The
    School - 1:1 SEN Teaching Assistant" (one ad), which is company: the
    risk that made legal's rule `=> legal`, but rare here, so kept.
  - Mixed headings stay unmapped, as "please note" does: "important
    notice", "additional information", "please read before applying",
    "footnote".
