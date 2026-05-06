# Migration: Brain Repo

When migrating live personal data into the brain (an old-format archive, a sibling repo's content, scattered notes from another tool), the migration script MUST follow the safety rules below.

## The Five Migration Rules

1. **Default to `--dry-run`.** Never mutate the destination on bare invocation. Real migrations require an explicit `--apply` flag.
2. **Write dry-run output to a STAGING directory** under `.work/spaces/W-NNNN/migration-output/` (gitignored). Never to the real `human/` destination in dry-run mode. The owner spot-checks staging before committing.
3. **Treat the source directory as READ-ONLY.** No writes ever, not even to the source's own `.git/`.
4. **Log every per-file decision** (kept, dropped, anomaly) to a `MIGRATION_LOG.md` in the staging dir.
5. **Be re-runnable.** `--apply` should be idempotent — safe to re-run if the first attempt was interrupted mid-batch.

## Two-Phase Shipping Pattern

Pair the rules above with two-phase shipping:

- **Phase 1:** commit the script + template + dry-run sample as safe artifacts. Owner reviews the staging output.
- **Phase 2:** owner says "apply"; you run `--apply` and commit the data as a separate commit (typically much larger than Phase 1).

Two phases keep the high-risk operation reversible until the last moment, and produce a clean git history with the artifacts and the data-mass separated.

Proven on the canonical instance's bulk daily-log migration (69 files transformed from an older legacy format, zero owner regret).

## Legacy Content Preservation

For *legacy* personal-content directories (old journals, accumulated notes), prefer **preserve as-is in a `legacy-*/` subdirectory** over lossy schema conversion. The archetype's gates exempt `legacy-*/`. Start the new format going forward and let the legacy content age.
