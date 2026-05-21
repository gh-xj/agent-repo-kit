# Influence Metrics — What They Mean and Where They Fail

A reference for **interpreting** the numbers. None of these is a single source of truth; their value comes from cross-checking across the three lenses.

## Author-level metrics

### h-index

**Definition.** A researcher has h-index `h` if `h` papers each have ≥`h` citations.
**Use it for.** Rough single-number sense of an active researcher's combined productivity + impact, _only when you also know career length_.
**Where it fails.** Field-dependent (bio inflates, theory CS deflates). Career-stage-biased (only grows with time). Gameable via salami-slicing and self-citation rings. Useless for industry researchers without curated profiles.

### m-quotient (preferred default)

**Definition.** `m = h-index / years-since-first-publication`. Hirsch's own normalization.
**Use it for.** Comparing across career stages. A 5-year postdoc with `h=15, m=3.0` outpaces a 30-year senior with `h=40, m=1.3`.
**Calibration (CS / ML, 2026):**
| m | Read |
| --- | ----------------------------------------------------- |
| <1 | Below typical productivity for the field |
| 1–2 | Typical active researcher |
| 2–3 | Strong; visible in subfield |
| 3+ | Top performer; rising star or established luminary |

Use m-quotient as the **default** Lens A citation metric. Mention raw h only when career length is unknown.

### g-index

**Definition.** Largest `g` such that the top `g` papers together have ≥`g²` citations. Weights highly-cited papers more than h-index does.
**Use it for.** Catching researchers whose impact is concentrated in a few landmark papers (which h-index hides).
**Where it fails.** Same field-dependence as h-index.

### i10-index

**Definition.** Number of papers with ≥10 citations (Google Scholar's metric).
**Use it for.** Productivity sanity check. Mostly redundant with h-index.

### hI-index (author-count-adjusted)

**Definition.** Adjusts h-index for the number of authors per paper.
**Use it for.** ML/Bio papers with 20+ authors where h-index is inflated by mass collaborations.
**Where it fails.** Multiple competing formulae; not standardized.

### Total citations

**Use it for.** Magnitude check.
**Where it fails.** One viral paper dominates. Look at the _distribution_, not just the total. "30k cites from 200 papers, top paper 5k" is healthier than "30k from 4 papers, top paper 25k" (one-hit-wonder territory).

## Paper-level metrics

### Raw citation count

**Use it for.** Floor-level signal that anyone has noticed the paper.
**Where it fails.** Volume only, no direction. Wrong-but-cited papers and right-but-cited papers look identical.

### Field-Weighted Citation Impact (FWCI)

**Definition.** Citations normalized against the world average for the same field, year, document type. FWCI=1 = field average. Available in OpenAlex (free) and Scopus (paid).
**Use it for.** Cross-field comparison and flagging unusually high-impact papers.
**Calibration.** FWCI > 1.5 = strong; > 3 = exceptional; > 5 = field-changing.
**Where it fails.** Field classification fuzzy at boundaries. Newer papers can't have stable FWCI yet.

### Highly Influential Citations (Semantic Scholar)

**Definition.** Algorithmic subset of citations where the citing paper _built on_ the cited work, not just referenced it.
**Use it for.** Stripping ceremonial citations. **Strong** Lens B signal.
**Calibration.** A 500-cite paper with only 3 highly-influential citations has much shallower impact than the headline number suggests.

### scite.ai Smart Citations ★ Lens B primary

**Definition.** Each citation classified as `supporting`, `contrasting`, or `mentioning`. Built on a fine-tuned classifier reading the actual citing-sentence context.
**Use it for.** The single most actionable Lens B signal — tells you whether peers used the paper or critiqued it.
**Calibration:**
| Pattern | Read |
| ------------------------------------------------------ | ------------------------------------- |
| supporting >> contrasting; >100 cites | well-received |
| contrasting >20% on a methods paper | contested — read the contrast carefully |
| mostly `mentioning` | ceremonial citation; shallow impact |
| Almost no citations, all `supporting` | too-new-to-tell |
**Where it fails.** Free tier limited; paid for serious vetting. Classifier has false positives. Coverage thin in some subfields.

### Altmetric Attention Score

**Definition.** Source-weighted aggregate of news, blogs, X, Reddit, policy mentions, Wikipedia, Mendeley. Author-bias-adjusted (self-promotion downweighted).
**Use it for.** Cross-disciplinary public attention.
**Calibration.** >50 = real attention; >500 = viral; >2000 = mainstream news.
**Where it fails.** Doesn't measure correctness. Strong English-language bias. Score=0 is common for genuinely good but technical papers.

## Self-disclosed rigor

### NeurIPS / ICML / ICLR / JMLR Reproducibility Checklist

**What it is.** Author-completed yes/no/justification table embedded in the paper since 2021. Covers code release, data, ethics, limitations, compute, reproducibility evidence.
**Use it for.** **Free Lens C signal sitting inside the PDF.** Read it. Authors are rewarded for honest "no" answers, so:

- Honest "no" with justification → maturity.
- "Yes" with handwave → red flag (claim without backing).
- Missing entries on a NeurIPS-track paper → process failure or hidden weakness.
  **Where it fails.** Doesn't apply outside top ML venues. Some authors check "yes" optimistically.

### MLRC / Papers With Code

Independent reproducibility reports filed against published papers. When present, they're the strongest claim-level signal possible.

## Social / qualitative signals

### X / Twitter

**Use it for.** Real-time peer review. For LLM/AI work in 2024–2026, frontier-lab researchers and independent practitioners post critique within days. **Who** is talking matters more than **how many**.

- Skeptical reply from one known researcher >> 50 cheerleading retweets.
- Quote tweets with critique >> retweets without context.
  **Where it fails.** Hype amplification, echo chambers, selection bias.

### Hacker News + r/MachineLearning

The top comment is often a sharp methodology challenge — read it before forming a view. Mid-quality on average; high-quality at the top.

## How to combine across lenses

Don't compute a weighted sum. Use the metrics as **independent witnesses inside each lens**, then triangulate the lenses.

- **Lens A (Team)**: prefer m-quotient over raw h; cross-check with field-fit and Connected Papers neighborhood.
- **Lens B (Citation context)**: scite ratio + Highly-Influential count + PubPeer flag + Retraction Watch. Volume is last, not first.
- **Lens C (Claim-level)**: reproducibility checklist + code release + leaderboard standing + ML-pitfalls check.

Convergence across lenses → confident verdict. Divergence → name the mismatch and define the falsifier (what would resolve it).

## Calibration tables

### m-quotient bands (CS/ML 2026, repeated for ease)

- <1 below typical / 1–2 typical / 2–3 strong / 3+ top performer

### Venue tiers

| Tier | Examples                                               |
| ---- | ------------------------------------------------------ |
| S    | NeurIPS, ICML, ICLR, ACL, CVPR, SIGMOD, SOSP, OSDI     |
| A    | EMNLP, AAAI, KDD, WWW, RecSys, TMLR, JMLR, COLM        |
| B    | Workshop papers at S/A venues, second-tier conferences |
| C    | Arxiv-only with no venue, low-tier conferences         |

### Field decay (LLM/Agent specifically)

- Paper age <6 months → citations not yet meaningful; weight Lens C heavily.
- 6–12 months → citations meaningful but volume still growing.
- > 12 months → citation volume relevant but discount headline impact: the field has moved.
- > 24 months → for LLM/Agent, treat as "historical interest" unless still on a leaderboard.

The goal is **calibrated reading**, not a score. Metrics tell you _how hard to push back_, not _whether the paper is right_. The "right" question is settled by independent replication, not citations.
