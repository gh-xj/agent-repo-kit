# Work CLI Extraction — Decoupling `work` into its own OSS Project

**Status:** Decisions recorded 2026-05-02 — ready to break down into a Phase 1 implementation plan
**Date:** 2026-05-02
**Owner:** xj
**Scope:** Move the `work` CLI and its supporting surfaces out of `agent-repo-kit` into a new standalone open-source repository (`gh-xj/work-cli`), and reshape `agent-repo-kit` to consume it as an external dependency.

**Cross-reference:** `2026-05-02_ark-deletion_design.md` deletes the `ark` CLI.
Recommended sequencing is **ark-deletion first, then this extraction** — once
`ark` is gone, `cli/` contains only the work CLI plus shared infra, and many
of the integration touchpoints below (e.g. `internal/scaffold`) no longer
exist to reshape.

## §0. Decisions Locked In (2026-05-02)

- **Repo name:** `gh-xj/work-cli`.
- **Shared infra strategy:** Option A — duplicate the ~210 LoC of `appctx`, `cliruntime`, `io`, `log` into `work-cli`. Revisit if drift becomes painful.
- **Install UX:** two separate installers. ARK's `install.sh` no longer ships or installs `work`. Adopters run `work-cli`'s installer independently.
- **`ark init --operations work` when `work` binary is missing:** **hard fail** with an actionable error pointing to `work-cli`'s install URL.
- **`work-cli-qa` skill ownership:** moves with `work-cli`. The new repo adopts the Claude Code project-skill convention (`.claude/skills/`).
- **Existing adopters:** none outside this machine. We have full freedom; no backwards-compatibility shim or migration note needed for downstream consumers. ARK release notes still describe the split for the public record.
- **Versioning compatibility:** add a `min_work_version` field to `.conventions.yaml` when the `work` opt-in is declared. ARK's `task verify` (and the `scripts/verify.sh` shipping as part of the ark-deletion epic) runs `work --version` and compares against the declared minimum, hard-failing on skew. Cheap insurance against silent version-skew bugs once `work` releases independently.

---

## §1. Motivation

`agent-repo-kit` (ARK) currently bundles two products in one repository:

1. **The convention kit** — `ark` CLI, scaffolding, contract verification, skills, adapters, install pipeline.
2. **The work tracker** — `work` CLI, the `.work/` store format, the `work-cli` skill, the `work-cli-qa` self-evolving release gate.

This bundling has costs:

- **Release coupling.** A bug in `work` blocks an `ark` release and vice versa. A `work` consumer who doesn't want the rest of the convention has to install the whole kit.
- **Conceptual coupling.** New adopters read the kit and infer that `.work/` is mandatory. The current `convention-engineering` skill treats `work` as an _operation_ — but the kit's docs and scaffolds present it as the default.
- **OSS surface.** `work` is the part most likely to attract independent users (it's a self-contained, language-agnostic local task tracker). It cannot be discovered, starred, or contributed to as its own thing while it lives inside ARK.
- **Skill ownership.** The `work-cli-qa` skill is a release gate for `work`'s implementation. It belongs _with_ the implementation it gates, not in the kit that consumes it.

**Goal:** `work` becomes its own repo with its own release cadence, README, install path, and issue tracker. ARK remains the convention kit and _integrates_ `work` as one supported (but no longer bundled) operation.

---

## §2. Current Coupling Audit

### §2.1 Go code (light, almost clean)

`work` CLI imports from ARK:

| Package               | Used by `work`           | Reverse-coupled?           |
| --------------------- | ------------------------ | -------------------------- |
| `internal/appctx`     | yes (config root, env)   | no                         |
| `internal/cliruntime` | yes (cobra/kong wrapper) | imports `appctx` only      |
| `internal/io`         | yes (json/text output)   | no                         |
| `internal/log`        | yes (slog setup)         | no                         |
| `internal/work`       | yes (store)              | one ark caller (see below) |
| `internal/workcli`    | yes (cobra commands)     | none                       |

ARK code imports from `work` in exactly **one place**:

- `cli/internal/scaffold/docs.go:98` — `work.New(...).Init()` during `ark init` to bootstrap a `.work/` store in target repos.

**Implication:** code separation is straightforward. The shared infra packages (`appctx`, `cliruntime`, `io`, `log`) total ~210 LoC and have no ARK-specific concepts; they can be (a) duplicated into `work-cli`, (b) extracted into a third shared module, or (c) kept inside ARK and imported by `work-cli` as a Go module dependency. See §3.

### §2.2 Non-code surfaces (heavy)

These all reference `work` and need a story:

- `skills/work-cli/` — kit-owned skill teaching agents how to use the CLI.
- `adapters/{claude-code,codex,cursor}/work-cli*` — adapter-specific mirrors.
- `.claude/skills/work-cli-qa/` — self-evolving release gate for `work` source.
- `cli/internal/scaffold/{config,support,docs}.go` — emits `task work` task, `.gitignore` rules, README convention block when `work` is in operations list.
- `cli/internal/contract` — `work` is referenced as a recognized operation.
- `Taskfile.yml` — exposes `task work -- ...` for the kit's own dogfooding.
- `docs/implementation/work-cli-v0.md`, `docs/implementation/tickets-to-work-migration.md`.
- `skills/convention-engineering/references/operations/work.md`.
- `examples/demo-repo/` — fully wired demo using `work`.
- `.convention-engineering.json` — declares `work` as an active operation here.

---

## §3. Architecture Options

### Option A — Hard split, duplicate shared infra (recommended)

`work-cli` becomes self-contained: it copies the small shared infra packages (`appctx`, `cliruntime`, `io`, `log` — ~210 LoC total) into its own repo and goes its own way. ARK keeps its own copies. The `internal/scaffold/docs.go` call to `work.New().Init()` is replaced by either (i) shelling out to `work init` if installed, or (ii) writing the empty `.work/config.yaml` directly via templates (the init logic is small).

**Pros:**

- Zero cross-repo Go module dependency. Each repo has independent release cadence.
- `work-cli` has no awareness of ARK. Cleanest OSS story.
- Simple to refactor later if shared concerns emerge.

**Cons:**

- ~210 LoC duplicated across two repos. Drift risk on infra changes.

### Option B — Extract shared infra into a third Go module

Create `gh-xj/cli-infra` (or similar) hosting `appctx`, `cliruntime`, `io`, `log`. Both `agent-repo-kit` and `work-cli` depend on it.

**Pros:**

- No duplication.

**Cons:**

- Three repos to release-coordinate for what is currently 210 LoC of glue. Premature abstraction.
- Public Go module surface for utility code that has no external consumers.

### Option C — `work-cli` depends on ARK as a Go module

`work-cli` imports `github.com/gh-xj/agent-repo-kit/cli/internal/...` as an external module.

**Pros:**

- No duplication, no third repo.

**Cons:**

- `internal/` packages can't be imported externally — would have to promote them to public packages with a stability contract.
- Reverses the conceptual hierarchy: the standalone tool depends on the kit.
- Defeats the goal of decoupling release cycles.

### Recommendation

**Option A.** The duplication is small and stable; the simplicity buys real decoupling. If infra grows or drift becomes painful in 6+ months, revisit Option B.

---

## §4. What Moves vs. What Stays

### §4.1 Moves to new `work-cli` repo

```
cli/cmd/work/                       → cmd/work/
cli/internal/work/                  → internal/work/
cli/internal/workcli/               → internal/workcli/
cli/internal/appctx/        (copy)  → internal/appctx/
cli/internal/cliruntime/    (copy)  → internal/cliruntime/
cli/internal/io/            (copy)  → internal/io/
cli/internal/log/           (copy)  → internal/log/
.claude/skills/work-cli-qa/         → .claude/skills/work-cli-qa/
docs/implementation/work-cli-v0.md  → docs/design/work-cli-v0.md
docs/implementation/tickets-to-work-migration.md → docs/design/migration-from-tickets.md
```

Plus new files: `README.md`, `LICENSE`, `Taskfile.yml`, `install.sh` (or rely on `go install`), GitHub release workflow, `CHANGELOG.md`.

### §4.2 Stays in `agent-repo-kit`, gets reshaped

- `skills/work-cli/` — keep, but reframe from "use the bundled CLI" to "if you adopt the `work` operation, here's the workflow contract." Add an install pointer.
- `adapters/*/work-cli*` — same reframing. These are still kit-owned because the _integration_ is what the kit teaches.
- `cli/internal/scaffold/{config,support,docs}.go` — keep `work` as a recognized operation, but:
  - Drop the Go-level `work.New().Init()` call.
  - Detect `work` on PATH. If present, shell out to `work init` for store bootstrap. If absent, **hard fail** with: "operation `work` requires the `work` CLI on PATH; install from https://github.com/gh-xj/work-cli".
  - No auto-install, no template-only fallback. Hard fail keeps the convention contract honest: an active operation must have its tool present.
- `skills/convention-engineering/references/operations/work.md` — keep, point to the external repo.
- `Taskfile.yml` — `task work` becomes a passthrough that requires `work` on PATH (no longer builds it locally).
- `examples/demo-repo/` — keep, with a setup note that `work` must be installed first.
- `.convention-engineering.json` — unchanged in shape, but the kit no longer dogfoods `work` as a _source-controlled component_; it's just an operation we adopt.

### §4.3 Deletes from `agent-repo-kit`

Anything that becomes pure pointer to the new repo and adds no value (e.g., `docs/implementation/work-cli-v0.md` once it lives at `work-cli/docs/`).

---

## §5. Convention Contract & Verification Implications

The `convention-engineering` skill currently presents `work` as an _operation_ you can opt into. That framing already supports decoupling — we just need to honor it more rigorously:

- **Operation activation should not assume the binary is bundled.** `ark check` for `work`-active repos should produce a clear "install `work` from <url>" diagnostic when the binary is missing, not a generic failure.
- **Verification gate.** `task verify` in ARK currently runs `ark check --repo-root .` plus `task work -- view ready` and similar. Post-split, these still work _if `work` is on PATH_ — CI installs it as a prereq.
- **The `work-cli-qa` skill leaves with `work`.** ARK's verification gate stops covering `work` source quality; that becomes `work-cli`'s own CI.

---

## §6. Migration Plan (Phased)

### Phase 0 — Decide & document

- Review and approve this proposal. Record the chosen option in `docs/plans/`.

### Phase 1 — Stand up the new repo, no ARK changes

- Create `gh-xj/work-cli`. Copy listed files. Add `LICENSE`, README, install script, release workflow.
- First release: `work-cli v0.x.0` with feature parity to current `cli/cmd/work`.
- Set up issue tracker, basic contributing guide.

### Phase 2 — ARK consumes the external CLI

- In ARK, drop the Go import of `internal/work` from `internal/scaffold/docs.go`. Replace with PATH-detection + shell-out to `work init`, hard-failing if absent.
- Update `task work` to require external `work` binary; remove the local build dependency.
- Update install pipeline: ARK release stops shipping the `work` binary in its tarball. `install.sh` no longer attempts to install `work`; it prints a final-stanza pointer to `work-cli`'s installer for users who want the `work` operation.
- Update `ark init`/scaffolding: when `work` is in the operations list, verify the binary on PATH up front and hard-fail with the install URL if missing.
- Update skills/adapters to reframe from "bundled" to "external dependency."
- Delete the now-moved code from `cli/internal/work` and `cli/internal/workcli`. Keep the shared infra packages (`appctx`, `cliruntime`, `io`, `log`) — both repos now own their own copies.

### Phase 3 — Cleanup & reframing

- Move `.claude/skills/work-cli-qa/` to `work-cli` repo.
- Reduce `docs/implementation/work-cli-v0.md` to a stub linking out, or delete.
- Update `convention-engineering` skill text and `examples/demo-repo/README.md`.
- Add a CHANGELOG entry / release note for ARK explaining the split for existing users.

### Phase 4 — Backfill (optional)

- If drift between the two copies of shared infra becomes painful, reopen Option B.

---

## §7. Open Questions

None remaining. All design decisions are locked in §0 above. The doc is ready
to be broken down into a `docs/plans/` implementation plan once the
ark-deletion epic lands.

---

## §8. Out of Scope

- Changing the `.work/` file format or store schema.
- Rewriting the `work` CLI in another language.
- Decoupling `convention-engineering` itself from ARK.
- Building a Homebrew tap or other package manager distribution for `work` (can come after the split lands).

---

## §9. Success Criteria

- `gh-xj/work-cli` exists, has tagged releases, and can be installed without any reference to ARK.
- `agent-repo-kit` builds and `task verify` passes without including any `work-*` Go source.
- `ark init --operations work` produces a working `.work/` store on a host that has `work` on PATH, and a clear actionable error on a host that doesn't.
- The `work-cli` skill in ARK is accurate against the externally-released `work` binary.
- No regressions in the `examples/demo-repo` walkthrough.
