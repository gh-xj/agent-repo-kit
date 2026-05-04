---
name: paper-vetting
description: "Use BEFORE reading or summarizing any research paper. Triangulate trust across three independent lenses (team / citation context / claim-level evidence), then decide how skeptically to read it. Triggers: 'research this paper', 'read this paper', 'summarize this arxiv', 'what's the reputation of', 'is this paper credible', 'vet this paper', '评估这篇论文', '这篇论文靠谱吗', '这个团队厉害吗', any arxiv.org / openreview.net / aclanthology.org / proceedings URL handed in for review."
---

# Paper Vetting

Vet a paper through **three independent lenses**, then triangulate. Output is a calibrated reading recommendation, not a single trust score. Reputation alone tells you _how seriously_ to read; only the convergence of lenses tells you _whether the claims survive_.

Methodology-first. Same flow regardless of field, but the ML/Agent lens has its own pitfall list — see `references/ml-agent-pitfalls.md`.

## Why three independent lenses

Single-lens vetting fails predictably:

- **Team-only** → famous-name halo: senior teams write wrong papers and you anchor on credentials.
- **Citation-count-only** → hype amplification: viral wrong papers and ceremonial citations both inflate the number.
- **Self-reported-rigor-only** → checklist theater: authors check boxes without substance.

Independence prevents each failure mode. **Mismatch between lenses is itself the signal** — strong team + critical citation context + no released code = a different problem than weak team + replicated result + open code.

## When to use

- User hands over an arxiv / OpenReview / venue URL and asks to "research", "read", "summarize", or "vet" it.
- User asks about a paper's team or credibility.
- User is deciding whether to invest time in a paper.

## When NOT to use

- User wants you to _implement_ a method from a paper they already trust → just read.
- User wants a literature review across many papers → use a survey workflow.
- The "paper" is a blog post — most lenses still apply, but venue/checklist signals will be empty.

## Process

```
Phase 1: Identify        →  metadata via arXiv/OpenAlex (cheap), see sources.md tool ladder
Phase 2: Three lenses    →  run in parallel; each lens votes independently
   2a Team               →  position, m-quotient, field-fit, recency, Connected-Papers neighborhood
   2b Citation context   →  scite supporting/contrasting, Sem.Sch. highly-influential cites,
                            PubPeer flags, Retraction Watch
   2c Claim-level        →  reproducibility checklist (in-paper), code/data/weights, leaderboard,
                            for ML/Agent: contamination check (see ml-agent-pitfalls.md)
Phase 3: Triangulate     →  convergence → confident verdict; divergence → flag the mismatch
Phase 4: Verdict         →  trust band + reading posture + skepticism flags + falsifier
Phase 5: Read (opt-in)   →  read paper through the calibrated lens
```

**Default**: run Phases 1–4, present verdict, then ask before Phase 5. Skip vetting only on explicit "skip vetting".

## Phase 1 — Identify

Use the cheapest tool that works (see `references/sources.md` tool ladder):

- arXiv id → `https://arxiv.org/abs/<id>` (canonical) or arXiv API.
- DOI → OpenAlex API (`/works/doi:<doi>`).
- OpenReview → use the forum page; reviews + scores are gold.

Extract: title, abstract, version, year, all authors with affiliations, corresponding-author flags, venue (or "preprint"), and — for ML papers — whether a NeurIPS/ICML/ICLR/JMLR-style **author-completed reproducibility checklist** is included. The checklist is its own signal regardless of content.

## Phase 2 — Three lenses (run in parallel)

Dispatch lens lookups concurrently. Cap each lens at 3–4 web calls; stop early when the picture is clear.

### Lens A — Team (who)

For corresponding + visibly senior co-authors only (not all 20):

- Position + institution + tenure status.
- **m-quotient = h-index / years-since-first-publication**. Better than raw h for career-stage normalization. Round, don't fabricate.
- Field-fit: is this paper in their published lane? (Connected Papers neighborhood reality-check.)
- Recency: first/corresponding output in last 18 months.
- Subfield reputation: invited talks, keynotes, named in conferences.
- Practitioner credibility: code shipped, frameworks in use, production deployment.

Output: 2–3 sentences per author + a one-word characterization (`heavyweight` / `solid` / `rising` / `unknown` / `tourist`). Mark `?` when signal is thin — do **not** invent.

### Lens B — Citation context (how peers actually use it)

The lens v1 missed. Citation count is volume; citation **context** is direction.

- **scite.ai Smart Citations** — supporting / contrasting / mentioning ratio. <5% contrasting on a methods paper with >100 cites is a healthy sign; >20% contrasting is a red flag.
- **Semantic Scholar Highly-Influential Citations** — citations where the citing paper _built on_ this one, not just referenced it.
- **Connected Papers / ResearchRabbit** — is this paper a node in the live conversation, or isolated?
- **PubPeer** — search the title or DOI. Any post-publication concerns?
- **Retraction Watch** — retracted? expression of concern?
- **Replication record** — has anyone independently reproduced the headline claim?

Output: one short paragraph + a label (`well-received` / `contested` / `ignored` / `flagged` / `too-new-to-tell`).

### Lens C — Claim-level evidence (does the result survive scrutiny)

Reading what the paper _self-discloses_ + what's externally checkable.

- **Reproducibility checklist** (NeurIPS / JMLR / MLRC) — is one included? what's checked yes/no? (Authors are rewarded for honesty about limitations — "no" answers aren't disqualifying, but **missing fields** are.)
- **Code / data / weights / configs** — released? last commit recent? stars? independent forks?
- **Leaderboard / benchmark** — entry on Papers With Code? still SOTA, or has the field moved on?
- **For ML/Agent papers** — run the contamination + overclaim checks in `references/ml-agent-pitfalls.md`.
- **Methodology section** — sample size, ablations, statistical tests, preregistration.

Output: short paragraph + label (`fully-evidenced` / `partially-evidenced` / `claimed-only` / `unverifiable`).

## Phase 3 — Triangulate

Plot the three labels side-by-side. Three patterns:

| Pattern    | Lenses                   | Read                                                                      |
| ---------- | ------------------------ | ------------------------------------------------------------------------- |
| Convergent | All three positive       | High trust; mine for techniques.                                          |
| Convergent | All three negative       | Low trust; skim for ideas, do not cite numbers.                           |
| Divergent  | Team strong, others weak | "Hype paper from a real team" — read the _method_, distrust the _claims_. |
| Divergent  | Team weak, others strong | "Underdog actually right" — rare but real; trust the _result_.            |
| Mixed      | One ambiguous            | Use the ambiguous lens to define skepticism flags.                        |

The **mismatch is the story**. Don't average lenses into a single number.

## Phase 4 — Verdict

Output four things:

1. **Trust band**: `A` (read closely, mostly trust) / `B` (read with healthy skepticism) / `C` (skim, verify externally) / `D` (don't rely on without independent confirmation).
2. **Reading posture**: which sections to read carefully, skim, or cross-check. Tie posture to specific lenses (e.g., "Lens C flagged no code → don't trust the implementation details").
3. **Skepticism flags**: 2–4 specific things to watch for, each tied to _which lens_ surfaced it.
4. **Falsifier — what would change the verdict**: a concrete observable thing (replication on a leaderboard, retraction posted, code released, etc.). Without this, the verdict is static instead of a live hypothesis.

See `references/vetting-rubric.md` for the report template.

## Phase 5 — Read (only if user confirms)

Read the paper with the calibrated posture. Map skepticism flags onto specific sections. For ML/Agent papers, do the contamination check before believing any benchmark number.

## Non-negotiables

- **Three lenses run independently** before triangulation. Do not let the team lens leak into the citation-context lens — that's anchoring.
- **m-quotient over raw h-index** for career-stage fairness. Mention raw h only when career length is unknown.
- **Round, don't fabricate.** `h ~30k+` not `h = 29,847`. Scite ratios as percentages, not three-decimal floats.
- **Author-count > 10 is a yellow flag.** Score only corresponding + visibly senior co-authors; name the rest as "team labor".
- **Self-citation pattern matters.** ≥30% of citations being the team's own work is a flag, surface it explicitly.
- **Field decay**: AI/LLM/Agent papers older than 12 months → discount Lens B's volume signal; the field moves faster than citations accrue.
- **Don't conflate seniority with correctness.** Famous teams write wrong papers; surface the specific reasoning, not the credentials.
- **Honest uncertainty**: "no public signal on <name>" beats an invented score.
- **The falsifier is mandatory**: every verdict ends with what would change it.

## References

| File                              | Use for                                                                                                       |
| --------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| `references/sources.md`           | Tool ladder — cheap → better → best path for each lens. Includes APIs (arXiv, OpenAlex, Semantic Scholar).    |
| `references/influence-metrics.md` | h-index, m-quotient, g-index, FWCI, scite Smart Citations, Altmetric — what each means and where it fails.    |
| `references/vetting-rubric.md`    | The three-lens scoring rubric + report template + Keshav's 5 Cs as a complementary frame.                     |
| `references/ml-agent-pitfalls.md` | ML/LLM/Agent-specific failure modes: benchmark contamination, leakage, overclaiming, NeurIPS-checklist reads. |

## Common mistakes

- Reading the paper before vetting it (cognitive anchoring).
- Collapsing the three lenses into a weighted sum — kills the mismatch signal that's the whole point.
- Treating raw h-index as career-stage-normalized. Use m-quotient.
- Skipping Lens B because "the team is famous" — that's exactly when Lens B matters most.
- Skipping the NeurIPS reproducibility checklist when one is included — it's a free signal sitting in the paper.
- Forgetting the ML-specific contamination check; in 2024–2026 it's the dominant failure mode for benchmark claims.
- Skipping the falsifier line — without it the verdict is a static judgment, not a live hypothesis.
