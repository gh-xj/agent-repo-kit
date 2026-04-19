# Report Template

Phase 6 of `attack-architecture` writes a file that matches this skeleton exactly. Placeholders in `{CURLY}` are filled at write time. Keep the H2 markers in place — the skill locates sections by heading, so renaming them breaks future iteration.

Target path: `.docs/arch-attacks/{YYYY-MM-DD}-{SCOPE_SLUG}.md`.

---

```markdown
---
scope: "{SCOPE}"
scope_slug: "{SCOPE_SLUG}"
date: "{YYYY-MM-DD}"
depth: "quick" | "normal" | "thorough"
lenses: ["L1", "L2", ...]
intake:
  context_constraints: [...]
  notes: "{optional free text}"
generated_by: "attack-architecture skill"
---

# Architecture Attack Report — {SCOPE}

## Executive summary

Top findings ranked by `severity × confidence × lens-coincidence`. One line each.

1. **{title}** — {one-line impact}. → {recommended action}.
2. ...

## Hotspots

Files or modules flagged by **2 or more** lenses. These are the principal-engineer signal — when overengineering and coupling converge on the same file, that file is the real problem.

| File / module | Lenses | Finding titles |
| ------------- | ------ | -------------- |
| {path}        | L1, L3 | ...            |

If no hotspots: `_No module was flagged by more than one lens. The findings below are orthogonal._`

## Findings by lens

### L1 — Overengineering

#### {finding title}

- **Evidence:**
  - `{file}:{line}` — `{quoted snippet}`
  - ...
- **Severity:** {low/med/high} · **Confidence:** {0–100} · **Blast radius:** {narrow/module/cross-cutting}
- **Why it hurts:** {≤40 words}
- **Minimal fix:** {from Phase 4, if present — else "Not expanded"}
- **Structural fix:** {from Phase 4, if present — else "Not expanded"}

_(Repeat per finding. Repeat section per lens. Omit lens sections with no findings — but list them in the "What was NOT attacked" section so the absence is explicit.)_

## Debate transcripts

Only populated when depth = `thorough`. One subsection per debated finding.

### {finding title}

**Verdict:** `confirmed` / `exaggerated` / `dismissed` · **Final confidence:** {0–100}
**Recommended action:** {≤30 words}

**Judge's rationale:** {≤100 words}

<details>
<summary>Attacker</summary>

{full attacker output}

</details>

<details>
<summary>Defender</summary>

{full defender output}

</details>

## Ranked mitigation plan

Excludes `dismissed` findings. Ranked by impact × (1 / effort).

| #   | Finding | Impact | Effort | Blast radius | Recommended action |
| --- | ------- | ------ | ------ | ------------ | ------------------ |
| 1   | {title} | high   | low    | module       | ...                |

## What was NOT attacked

Explicit out-of-scope list — so readers do not mistake silence for reassurance.

- **Out of scope paths:** paths not touched by this run.
- **Opted-out lenses:** {e.g. "L6 Concurrency was not selected — async / thread-safety issues are not reflected here."}
- **Lenses that returned zero findings:** {list — absence of findings is not proof of absence of smells; it means these specific patterns did not surface in the evidence read}.
- **Known limitations of this run:** {e.g. "generated files skipped", "tests not attacked"}.

## Handoff

Choose one:

1. Brainstorm a restructure for the top finding → invoke `superpowers:brainstorming`.
2. Plan concrete fixes across the mitigation plan → invoke `superpowers:writing-plans`.
3. Stop here.
```
