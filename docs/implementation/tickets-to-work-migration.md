# Migrating From `.tickets` To `work`

The new `work` CLI is a replacement product, not a compatibility layer for
the legacy `.tickets` tracker. There is no automatic `.tickets` importer in
v0.

Use this guide when a repo is ready to move from `.tickets/` to `.work/`.

## Stance

- `work init` creates `.work/`; it must not read, rewrite, or import
  `.tickets/`.
- Keep `.tickets/` in git history, but remove it from the active tree once
  all still-relevant tickets have been moved or deliberately left behind.
- Move only still-relevant legacy tickets. Closed or obsolete tickets should
  remain historical records.
- Legacy tickets should go through inbox/triage or be manually recreated as
  work items.
- Preserve useful legacy metadata such as `legacy_id` and `legacy_path`.
- Treat old terminal states as historical context. Evidence-gated closure is
  later work, so do not translate old `DONE` directly into accepted proof.

The migration is intentionally human-reviewed. Old ticket bodies often mix
intent, notes, stale action items, and status history; blindly importing them
would preserve the wrong abstraction.

## Recommended Paths

### Inbox and Triage

Use this path when the legacy ticket needs judgment, splitting, rewriting, or
deduplication:

1. Read `.tickets/all/T-NN-*/ticket.md`.
2. Run `work inbox add` with the ticket title, useful context, source path, and
   old ID.
3. Review the inbox entry with `work inbox` or `work show`.
4. Run `work triage accept` only if the repo should track it as active work.
5. Remove the old active tracker only after the new work item exists.

### Manual Recreation

Use this path when the ticket is already clean, current, and clearly scoped:

1. Read `.tickets/all/T-NN-*/ticket.md`.
2. Run `work new` with a cleaned title, context, priority, and area.
3. Copy only the useful summary and acceptance criteria.
4. Add legacy metadata to the new work item.

Example metadata:

```yaml
legacy_id: T-01
legacy_path: .tickets/all/T-01-example/ticket.md
legacy_source: .tickets
```

## What Not To Build In v0

- Do not add a bulk `.tickets` importer.
- Do not add a migration command that promotes every open ticket into
  `.work/items/`.
- Do not infer dependency relations from old free-form ticket text. Relations
  are later work.
- Do not treat `.tickets` audit history as evidence-gated `work` closure.
  Evidence is later work.

## When To Stop Creating `.tickets`

A repo can stop creating new `.tickets` entries once:

- `work inbox`, `work inbox add`, and `work triage accept` cover incoming
  request capture.
- `work view` covers the daily scan of canonical active work.
- `work show` covers direct lookup by ID.
- active `.tickets` entries have either been moved or deliberately left behind.
- AGENTS.md/CLAUDE.md point agents at `work` instead of `.tickets`.

At that point, remove `.tickets/` from the active tree and update verification
so `task verify` checks `.work/` instead of the legacy tracker.
