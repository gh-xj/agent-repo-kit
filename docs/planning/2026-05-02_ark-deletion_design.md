# Ark CLI Deletion / Refactor — Intent

Status: planned, not yet executed.
Tracking: `.work/items/W-0010.yaml` (epic: convention-engineering refactor).

Cross-reference: `2026-05-02_work-cli-extraction_design.md` covers extracting
`work` into its own OSS repo. Recommended sequencing is **ark-deletion first,
then work-cli extraction** — once `ark` is gone, the only Go code left in
`cli/` is the work CLI plus shared infra, and "extracting work" reduces to
"move the rest of `cli/` to a new repo."

## Why

The `ark` CLI was built to enforce a `.convention-engineering.json` contract,
scaffold convention files, and orchestrate evaluator handoffs. The convention-
engineering skill is being refactored to drop that contract entirely (replaced
by a small `.conventions.yaml` opt-in descriptor) and to defer the evaluator
integration. Once those decisions land, `ark` is left holding only:

- a checker for a config file that no longer exists,
- a scaffolder for a doc taxonomy that the agent can stamp out from a few-line
  YAML descriptor,
- an orchestrator for a deferred evaluator pipeline,
- a tasklint for stack profiles that are being deleted.

In short: every justification for `ark` is being removed in this epic.

## What gets deleted

Confirmed for deletion:

- `cli/cmd/ark/` — binary entry.
- `cli/internal/arkcli/` — command tree.
- `cli/internal/contract/` — `.convention-engineering.json` checker.
- `cli/internal/scaffold/` — `ark init`.
- `cli/internal/orchestrator/` — `ark orchestrate`.
- `cli/internal/tasklint/` — Taskfile linter scoped to deleted profile rules.
- `cli/internal/evaluator/` — only consumed by `internal/orchestrator`
  (deleting). Verified 2026-05-02. Rebuild fresh later if the evaluator skill
  returns.
- `cli/internal/skillbuilder/` — only consumed by `internal/arkcli/skill_init.go`
  and `skill_audit.go` (both deleting). Verified 2026-05-02.
- `cli/internal/skillsync/` — only consumed by `internal/arkcli/skill_check.go`
  and `skill_sync.go` (both deleting). Verified 2026-05-02.
- `cli/internal/upgrade/` — only consumed by `internal/arkcli/upgrade.go`
  (deleting). Verified 2026-05-02.

Sweep targets (in-scope for this epic, must lose ark references):

- `examples/demo-repo/` — currently demos `ark` workflow; rebuild around
  `.conventions.yaml` or delete the example outright.
- `adapters/claude-code/SKILL.md`, `adapters/cursor/convention-engineering.md`,
  `adapters/codex/SKILL.md`, plus any `adapters/*/taskfile-authoring/SKILL.md`
  copy that names `ark`.
- Root `CLAUDE.md`, `AGENTS.md`, `README.md`, `install.sh`, `Taskfile.yml`,
  `.convention-engineering/Taskfile.yml`.

Out of scope:

- `cli/cmd/work/` and `cli/internal/work*/` — the work CLI stays (and is the
  subject of `2026-05-02_work-cli-extraction_design.md`).
- `cli/internal/{appctx,cliruntime,io,log}/` — shared infra used by `work`.
- `scripts/link-dev-skills.sh`, `scripts/tag-release.sh` — repo-level shell
  scripts, unrelated to `ark`.

## The day after — `task verify` post-deletion

Today's chain runs `bash check.sh --repo-root .. --config
.convention-engineering.json` from `.convention-engineering/Taskfile.yml`.
That file (and the JSON contract it reads) is being removed alongside `ark`.
Proposed replacement shape:

```yaml
# Taskfile.yml (root) — post-deletion sketch
tasks:
  verify:
    desc: Verify the repo against .conventions.yaml
    cmds:
      - test -f .conventions.yaml
      # Agent-driven: convention-engineering skill loads the YAML and
      # runs each declared check. For unattended CI, see scripts/verify.sh
      # below — a thin yq loop that asserts opt-in artifacts exist and
      # leaves free-form `checks:` entries to the agent.
```

**Decided 2026-05-02:** ship `scripts/verify.sh`. It's ~30 lines, removes
the "agent must be running" dependency from CI, and the free-form `checks:`
list still defers to the agent for anything the script can't express.
`task verify` calls the script; the script asserts declared opt-in artifacts
exist (yq the keys, check filesystem), and exits 0 — leaving free-form prose
checks to the agent on demand.

## Install / distribution after deletion

`install.sh` currently downloads a tarball containing `ark` + `work`. After
this epic, the tarball is `work`-only. Combined with the work-cli extraction
epic, ARK ships **no binaries**. Decide whether `install.sh` survives at
all in ARK, or whether it moves with `work-cli` and ARK becomes a
content-only repo (skills + adapters + examples + docs).

## What replaces it

Nothing binary. The replacement pieces live in the convention-engineering
skill itself:

- A `.conventions.yaml` schema sketch (in
  `skills/convention-engineering/references/core/conventions-yaml.md`).
- A `task verify` target the repo defines, which reads `.conventions.yaml`
  and verifies declared items.
- Bootstrap and audit workflows reduced to "create / verify the YAML
  descriptor and the files it declares."

## Sequencing

1. Land the convention-engineering skill cleanup first (W-0010 phases 1–3).
   The skill stops referencing `ark`.
2. After the skill is clean, do the `ark` deletion as a single isolated
   change set:
   - Delete the confirmed packages listed above (caller analysis already
     done — see "Verified 2026-05-02" notes).
   - Resolve any compile errors in `cli/cmd/work` and other entry points.
   - Sweep `examples/demo-repo/`, `adapters/*`, root agent docs,
     `install.sh`, root + `.convention-engineering/` Taskfiles.
   - Land the `task verify` replacement (and `scripts/verify.sh` if we
     decide to ship one — see "The day after").
3. Commit as a single change set so it is easy to revert if a downstream
   consumer surfaces.
4. Then `2026-05-02_work-cli-extraction_design.md` Phase 1 begins.

## Success criteria

- `go build ./cli/...` succeeds with no `ark` packages in the tree.
- `task verify` exits 0 from a clean checkout, driven by `.conventions.yaml`,
  with no `ark` invocation.
- `grep -r '\bark\b' --exclude-dir=.git .` returns no hits in tracked files
  outside historical changelogs / git history references.
- `examples/demo-repo/` either runs cleanly under the new convention model
  or is removed entirely; no half-state.
- `install.sh` either ships only `work` until that extraction lands, or is
  removed; no broken installer references.

## Risk / Rollback

- The repo currently advertises `ark` in `CLAUDE.md`, `README.md`, and
  `install.sh`. Any external automation invoking `ark` will break.
  Mitigation: per `work-cli-extraction` §0, no external adopters are known;
  the only consumer is this machine. ARK release notes still record the
  deletion for the public record.
- The Go build will break in `cli/cmd/work` if the work CLI imports any
  shared code that itself imports a deleted package. Verified 2026-05-02
  that it does not (work depends only on `appctx`, `cliruntime`, `io`,
  `log`, plus its own `internal/work*`).
- Rollback: revert the single deletion commit; the skill cleanup commit can
  stay.

## Open Questions Deferred Past This Epic

These two collapse into one long-term verification model question:

- Long-term verification model: should `.conventions.yaml` ever be
  machine-checked by a dedicated runner, or is "agent reads the YAML, runs
  the declared checks" sufficient indefinitely? The shipped `scripts/verify.sh`
  (if we ship it) is the deterministic floor; anything beyond that gets
  decided when/if `convention-evaluator` returns. When it does return, the
  default position is that it rides on top of the YAML descriptor without
  re-introducing a separate contract layer.
