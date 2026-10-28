# Make boilerplate's verdict right

**Status:** shipped, unreleased

**Origin:** roadmap 4

**Started:** 2026-10-28, in e6226efeb90802722c14b59348f81a525fc3abca

**Finished:** 2026-10-28

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

- **Outcome:** done. Footers are boilerplate and common job lines
  aren't, as samples checked by eye both ways show; tests cover both
  rules, and PIPELINE.md's figures are updated.
- **Measured first,** as 18's figures came after the plan: of 40 lines
  the verdict took from embedded sections, 18 (45%) were about the job,
  so "about half" held; of 40 templates in an ad's ending run of
  repeated text, 26 (65%) were footers, the other 14 an advertiser's
  requirements running to the end of its ads.
- **Changes from the plan:**
  - The end run alone wasn't enough: with the wording guard only, 17 of
    30 texts newly taken as boilerplate were job lines, as the
    openings knew too few. So a footer must also talk as one, with a
    word of applying, getting in touch, the agency, its notices, "we",
    "our" or "us" (`FooterLine`, `footerWords`).
  - Step 2 needed `openings.txt` to know far more job lines: it caught
    about one in five of the job lines sampled. It gained the common
    openings of requirements ("The ability to", "Strong
    understanding", "Background in", "Experienced") and duties
    ("Provide", "Support", "Review", "Act as"). Checked on 3,000 lines
    from benefits, company, apply and legal sections: "Excellent",
    "Strong" and "High level of" alone opened perks, "Teach",
    "Teaching" and "Assist" agencies' names, so they went, or were
    narrowed to the skills they open.
  - The index learns the ending runs in a second pass (`Ends`, called
    by `Read`), once every ad is in, as which texts are repeated isn't
    known before.
- **Measurements**, the three samples, `claude/before` to
  `claude/after`: boilerplate 17.5% to 18.2% of the text, template for
  the job 3.3% to 2.6%; requirements' boilerplate share 11% to 14%;
  1,141 texts newly boilerplate, 153 newly templates. Read by eye, 28
  of 30 of the first were footers, notices or the advertiser's pitch
  (the two misses had "our" in a duty or a requirement), and 30 of 30
  of the second job lines. Afterwards, of 40 lines taken from embedded
  sections, about 9 were about the job (from 18); of 40 templates
  ending ads, about 8 were footers (from 26), from 1,457 such
  templates (3,363 before). Chunks: embedded 66.1% to 65.6% of tokens,
  the boilerplate chunk 9.6% to 10.2%; 21's guess places 0.65 points
  as duties, from 0.29, with the wider openings.
- **Tried and reverted:** the end run with only the wording guard,
  above; "excellent", "strong", "high level of", "teach", "teaching"
  and "assist" as openings.
- **What's left:** a duty or requirement with "our" or "we" in it, at
  the end of an advertiser's ads, is taken for a footer ("Utilise our
  CRM system…"); footers with none of the footer words stay templates.
- **The wash-up before the release:** the footer test's job line had no
  footer word, so it passed with the job-line exemption taken out, and
  nothing tested the ending counts summed over texts alike; both are
  tested now, each test failing without its code. Of templates opening
  with the new openings and holding a footer word, 26 texts, 7,681
  characters, all but one are job lines; the one is an agency whose
  name opens "Supporting".
