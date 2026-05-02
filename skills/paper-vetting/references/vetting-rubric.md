# Vetting Rubric and Report Template

The three-lens rubric, plus the canonical output format. Each lens is a label with backing evidence — **not** a numeric score. Convergence/divergence across lenses is the point; collapsing into one number kills the signal.

## The three lenses

### Lens A — Team (who)

Score with **labels**, not 1–10 numbers. For corresponding + visibly senior co-authors only.

| Label         | When to use                                                                           |
| ------------- | ------------------------------------------------------------------------------------- |
| `heavyweight` | Full prof at S-tier dept, m-quotient ≥3 in this exact subfield, named in conferences. |
| `solid`       | Tenured/tenure-track, m≈2–3, regularly publishes in this subfield's S/A venues.       |
| `rising`      | Junior faculty / strong postdoc, m≥2.5, recent multi-S-tier output.                   |
| `tourist`     | Real credentials, but this paper is far from their published lane (low field-fit).    |
| `unknown`     | No public signal sufficient to label. **Don't invent.**                               |

For each labeled author, give 2–3 sentences of evidence: position, m-quotient (rounded), field-fit, recency, subfield reputation, practitioner cred (when relevant). Floor matters more than average — a `heavyweight` with practitioner-cred floor of "no shipped artifact" is still a theory voice on production claims.

**Lens A label** (overall team): pick the dominant pattern.

- `senior-led` (≥1 heavyweight + supporting cast)
- `solid-mid-career` (mostly solid, no heavyweights)
- `student-driven` (1 senior advisor + many junior authors)
- `tourist-team` (credentialed but off-lane)
- `unproven` (mostly unknown / early-career)

### Lens B — Citation context (how peers actually use it)

**Independent of who wrote it.** Don't peek at Lens A while assessing this.

| Label             | When to use                                                                           |
| ----------------- | ------------------------------------------------------------------------------------- |
| `well-received`   | scite supporting >> contrasting; Highly-Influential cites present; no PubPeer flags.  |
| `contested`       | scite contrasting >20% on a methods paper, OR PubPeer thread, OR retraction concern.  |
| `ignored`         | Older than 12 months, low/flat citations, no influential builds-on, no social signal. |
| `flagged`         | Retraction Watch hit, expression of concern, or formal correction.                    |
| `too-new-to-tell` | <6 months old; signals haven't accrued.                                               |

Evidence: scite ratio (or "scite unavailable"), Highly-Influential count, PubPeer status, Retraction Watch status, Connected-Papers neighborhood note (live-conversation node vs. isolated).

### Lens C — Claim-level evidence (does the result survive scrutiny)

| Label                 | When to use                                                                         |
| --------------------- | ----------------------------------------------------------------------------------- |
| `fully-evidenced`     | Code + weights + configs + leaderboard standing + reproducibility checklist clean.  |
| `partially-evidenced` | Some artifacts released; some claims unverifiable. Most papers land here.           |
| `claimed-only`        | Numbers in paper; nothing released; no leaderboard; no replications.                |
| `unverifiable`        | Method paper with no code, no benchmark contributing, no reproducibility checklist. |

For ML/Agent papers, **always** run the contamination + overclaim checklist in `references/ml-agent-pitfalls.md` and surface the result here.

Evidence: code release status, reproducibility checklist read, leaderboard standing, ml-agent-pitfalls checklist hits, self-citation %.

## Triangulation patterns

| Pattern                               | A                  | B                   | C                   | Verdict pattern                          |
| ------------------------------------- | ------------------ | ------------------- | ------------------- | ---------------------------------------- |
| **All-positive convergent**           | senior-led / solid | well-received       | fully-evidenced     | A trust band                             |
| **All-negative convergent**           | unproven / tourist | ignored / flagged   | claimed-only        | D trust band                             |
| **Hype paper from real team**         | senior-led         | contested / too-new | claimed-only        | B trust on method, C trust on claims     |
| **Underdog actually right**           | unproven           | well-received       | partially-evidenced | B trust; result probably real            |
| **Famous-name anchor risk**           | senior-led         | too-new             | partially           | B; flag that A is doing all the work     |
| **Survey snapshot, fast-decay field** | solid              | varies              | n/a (it's a survey) | C unless very recent; useful as map only |

## Trust band

| Band | Read posture                                                              |
| ---- | ------------------------------------------------------------------------- |
| A    | Read carefully; treat results as likely correct; mine for techniques.     |
| B    | Read with healthy skepticism; verify key claims against external sources. |
| C    | Skim for ideas; do not cite numbers without independent replication.      |
| D    | Treat as opinion piece / preliminary thinking; do not act on conclusions. |

Mandatory band downgrades (any one triggers):

- **Lens C is `claimed-only` or worse** on a methods paper → max B.
- **Lens B is `flagged`** → max C.
- **ML/Agent paper hits 6+ items** in the `ml-agent-pitfalls.md` checklist → drop one band.
- **Self-citation ratio ≥30%** → drop one band.

## Reading posture (use Keshav's 5 Cs)

After verdict, frame the read with the 5 Cs:

- **Category** — what kind of paper is this? (method / system / survey / position / dataset)
- **Context** — what literature does it sit in?
- **Correctness** — given the lens labels, where do we expect correctness to break?
- **Contributions** — what is genuinely new?
- **Clarity** — is the writing tight enough to trust?

Tie posture to lens labels: "Lens C `claimed-only` → don't trust implementation details; read method as taxonomy not validated design."

## Reporting template

```
# 论文 / Paper

**<Title>** (<arxiv id> v<n>, <year>)

<1–2 line plain-language statement of what the paper claims>

# Lens A — Team
**Label**: <senior-led / solid-mid-career / student-driven / tourist-team / unproven>

- **<Author 1>** (<position>, <institution>) — <heavyweight/solid/rising/tourist/unknown>.
  m≈<value>, h≈<value>. Field-fit: <high/med/low>. <one-sentence characterization>.
- **<Author 2>** ...
- (only corresponding + visibly senior; rest noted as "team labor")

Self-citation ratio: <%, or "not checked">

# Lens B — Citation context
**Label**: <well-received / contested / ignored / flagged / too-new-to-tell>

- scite ratio: <supporting:contrasting:mentioning, or "unavailable">
- Sem.Sch. Highly-Influential cites: <count, or "—">
- PubPeer / Retraction Watch: <none / link / flag>
- Connected-Papers neighborhood: <isolated / live-conversation node>

# Lens C — Claim-level evidence
**Label**: <fully-evidenced / partially-evidenced / claimed-only / unverifiable>

- Reproducibility checklist (NeurIPS/ICML/ICLR/JMLR): <present-and-clean / present-with-issues / missing>
- Code: <link, last-commit-recency / not released>
- Weights/configs: <released / partial / not released>
- Leaderboard standing: <link, position / no entry>
- ML/Agent pitfalls hits: <count/12, top concern: ...>

# Verdict
- **Trust band: <A / B / C / D>**
- **Triangulation pattern**: <which row of the patterns table>
- **Reading posture (5 Cs framing)**: <2–3 sentences tied to lens labels>
- **Skepticism flags**:
  1. <flag tied to Lens X>
  2. <flag tied to Lens Y>
  3. <flag tied to ml-agent-pitfalls if applicable>
- **Falsifier — what would change the verdict**: <concrete observable thing>

# Want me to proceed and read the paper through this lens?
```

## Reporting rules

- **Labels, not numbers.** Per-author 1–10 was tidy in v1 but encouraged false precision and let the user collapse three lenses into a mean.
- **Run the lenses in independent passes.** Don't let Lens A leak into Lens B writeup. Hide the team label until Lens B is locked in.
- **Round, don't fabricate.** "h ~30k+", "scite 80% supporting", not three-decimal precision.
- **Floors matter.** A `heavyweight` with no practitioner cred floor is still a theorist on production claims.
- **The falsifier is mandatory.** No verdict ships without "what would change this".
- **Stay under ~80 lines** for the whole report unless asked for depth. Calibration, not autobiography.
- **If a lens is genuinely thin, label `unknown` / `too-new-to-tell` and say so.** Don't pad with low-confidence guesses.
