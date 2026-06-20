# Session Workflow Packaging

Use this reference when a user asks to mine Codex, Claude, or other agent
session history for durable value.

The goal is not to summarize history and not to auto-create memory. The goal is
to find repeated, costly, stable, missing, and safe workflows that should become
skills, skill references, checks, evals, structured ledgers, or active work.

## Source Tiers

| Tier | Examples | Handling |
| --- | --- | --- |
| Raw local corpus | transcript JSONL, command logs, tool outputs, shell snapshots, task chunks | local only; no cloud upload by default |
| Sensitive local cache | SQLite/FTS indexes, candidate JSONL, source manifests, cost JSON | rebuildable; do not promote as-is |
| Durable review output | aggregate reports, package shortlist, routing ledger, proposals | paraphrase-first; source pointers only |
| Durable package | approved skill reference, check, eval, script, protocol, or work item | narrow owner; verified before completion |

## Operating Loop

1. Establish a privacy contract before reading raw sessions.
2. Create a source manifest and query log.
3. Build a local derived index when the corpus is too large to inspect
   manually.
4. Extract broad candidate streams:
   - user preferences and corrections
   - decisions and rejected alternatives
   - follow-ups and blockers
   - prompt and handoff patterns
   - command, tool, runtime, and verification failures
5. Treat regex and embedding hits as a triage queue, not truth.
6. Cluster candidates into repeated workflows.
7. De-duplicate against existing skills, scripts, checks, Taskfiles, docs, and
   repo contracts.
8. Package only candidates that pass all gates below.
9. Record routing decisions in a structured local ledger.
10. Promote only after review and verification.

## Package Gates

| Gate | Test |
| --- | --- |
| Repeated | Seen more than once, or predictably recurring and costly. |
| Stable | Inputs, procedure, output, and stopping condition are clear. |
| Valuable | Improves speed, quality, reliability, safety, or auditability. |
| Missing | Not already adequately covered by an existing artifact. |
| Safe | Can be represented without raw transcript or secret-bearing detail. |
| Owned | Has a narrow durable destination and maintainer surface. |

If a candidate fails a gate, keep it in the work space or reject it.

## Routing Defaults

| Candidate | Default destination |
| --- | --- |
| Reusable judgment workflow | skill reference |
| Deterministic repeated failure | check, lint, test, eval, or script |
| Active unfinished work | `.work` item in the owning repo |
| Cross-surface promotion decision | structured routing ledger |
| Repo invariant | AGENTS/CLAUDE only after explicit approval |
| One-off preference | do not promote unless xj explicitly asks |
| Sensitive personal fact | do not promote; keep pointer only if needed |

## Required Artifacts

| Artifact | Purpose |
| --- | --- |
| source manifest | lets reviewers reproduce corpus coverage |
| query log | exact commands and SQL for headline claims |
| candidate streams | structured local queues with source IDs and hashes |
| privacy scan evidence | prevents raw/private leakage into durable outputs |
| package shortlist | one row per candidate workflow with owner and gate result |
| routing ledger | approved/rejected/deferred lifecycle for package candidates |
| review packet | concise decision surface for the next reviewer |

## Eval Cases

When changing this reference or the `harness-router` trigger wording, validate
the boundary cases in `references/session-workflow-packaging-evals.md`.

## Verification Accounting

For command or tool failures, do not report raw nonzero counts as bugs. Classify
them first:

| Class | Meaning |
| --- | --- |
| benign probe | expected no-match/search/read failure |
| fixed | failure followed by a passing rerun or implemented fix |
| unresolved | failure still matters and needs reporting |
| reported | failure cannot be fixed in-session and is called out |

Durable reports should state which failures remain unresolved.

## Privacy Rules

- Do not upload raw transcripts, command logs, thinking blocks, attachments, or
  FTS indexes without explicit approval.
- Do not paste transcript excerpts into durable docs unless the reviewer
  explicitly asks for a narrow, redacted quote.
- Prefer `{harness, source path, row/line, timestamp, session id, hash}` over
  free-text evidence.
- Treat redacted candidate JSONL as sensitive local cache, not public output.

## Stop Conditions

Stop at a proposal when:

- the destination is always-loaded instructions
- the candidate contains sensitive evidence
- the schema is still experimental
- no command/check reads the proposed structured store
- ownership belongs to another repo or skill
