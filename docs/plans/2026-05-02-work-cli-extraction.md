# work-cli Extraction — Implementation Plan

**Status:** Ready to execute. Ark-deletion is done; `cli/` now contains only
the work CLI plus shared infra. No design questions remain — see
`docs/planning/2026-05-02_work-cli-extraction_design.md` §0 for locked
decisions.

**Goal:** Stand up `gh-xj/work-cli` as an independent OSS Go module with
feature parity to the in-tree `work` CLI, then strip the work CLI from
agent-repo-kit and rewire integrations to consume the external binary.

**Sequencing:** Phase 1 stands up the new repo with no ARK changes (safe to
land independently and revert if needed). Phase 2 starts only after
`work-cli v0.1.0` is published.

---

## Phase 1 — Stand up `gh-xj/work-cli`

Land in `gh-xj/work-cli`. No changes to `agent-repo-kit` in this phase.

### 1.1 New repo skeleton

- [ ] `gh repo create gh-xj/work-cli --public --description "Local-first work tracker CLI"`
- [ ] Initial files at root: `README.md`, `LICENSE` (MIT, matching ARK), `.gitignore` (Go default), `CHANGELOG.md`.
- [ ] `go.mod` with `module github.com/gh-xj/work-cli`, Go version matching ARK's `cli/go.mod`.

### 1.2 Code transplant

Copy verbatim, then rewrite import paths from
`github.com/gh-xj/agent-repo-kit/cli/internal/...` to
`github.com/gh-xj/work-cli/internal/...`:

- [ ] `cli/cmd/work/` → `cmd/work/`
- [ ] `cli/internal/work/` → `internal/work/` (includes `store.go`, `store_test.go`)
- [ ] `cli/internal/workcli/` → `internal/workcli/`
- [ ] `cli/internal/appctx/` → `internal/appctx/` (copy)
- [ ] `cli/internal/cliruntime/` → `internal/cliruntime/` (copy)
- [ ] `cli/internal/io/` → `internal/io/` (copy)
- [ ] `cli/internal/log/` → `internal/log/` (copy)

After copy: `go mod tidy && go build ./... && go test ./...` must pass.

### 1.3 QA skill transplant

Both harness mirrors move:

- [ ] `.claude/skills/work-cli-qa/` → `.claude/skills/work-cli-qa/` in new repo
- [ ] `.codex/skills/work-cli-qa/` → `.codex/skills/work-cli-qa/` in new repo
- [ ] Update any path references in those skills' `cli/main.go` and `references/*.md` from `agent-repo-kit` to `work-cli`.

### 1.4 Repo-local Taskfile

- [ ] `Taskfile.yml` at root with: `build`, `test`, `lint`, `smoke`, `ci` (deps: lint+test+build+smoke), `qa` (runs the work-cli-qa harness).
- [ ] CI surface matches ARK's `cli/Taskfile.yml` shape so the existing QA skill keeps working.

### 1.5 Release pipeline

- [ ] `.github/workflows/release.yml` adapted from ARK's existing release workflow. Output artifacts: `work-{version}-{darwin,linux}-{amd64,arm64}.tar.gz` + `checksums.txt`.
- [ ] Each tarball contains the `work` binary + `LICENSE` only.
- [ ] `install.sh` at repo root, modelled on ARK's old `install.sh` but `work`-only. Detects OS/arch, downloads tarball, verifies checksum, installs to `/usr/local/bin/work` (or `$PREFIX/bin`).
- [ ] First tag: `v0.1.0`. Verify the release workflow produces all four tarballs + checksums.

### 1.6 README content

- [ ] Single-page README: what it is (one paragraph), install (one-liner curl-to-shell), `.work/` quick-start (`work init`, `work inbox add`, `work view ready`, `work show`), file format pointer, link back to ARK as the "convention kit that integrates this." Keep it tight.
- [ ] No marketing copy. No roadmap section.

### 1.7 Phase 1 done = exit criteria

- `git clone gh-xj/work-cli && cd work-cli && task ci` passes from a clean checkout.
- `curl -sSL https://raw.githubusercontent.com/gh-xj/work-cli/v0.1.0/install.sh | sh` puts a working `work` binary on PATH.
- `work --version` prints `v0.1.0`.
- The work-cli-qa skill (now in the new repo) can be invoked against `./cmd/work` and passes.

---

## Phase 2 — agent-repo-kit consumes external `work`

Land in `gh-xj/agent-repo-kit` after `work-cli v0.1.0` is published. Single
commit, easy to revert.

### 2.1 Strip work-cli source from ARK

- [ ] Delete `cli/cmd/work/`.
- [ ] Delete `cli/internal/work/` and `cli/internal/workcli/`.
- [ ] **Keep** `cli/internal/{appctx,cliruntime,io,log}/` — they're now duplicated copies, but ARK no longer has a binary in `cli/`. If `cli/` becomes empty of executables, delete `cli/cmd/` entirely. The shared infra packages can be deleted from ARK only if no remaining code in ARK imports them; verify with `grep`.
- [ ] If nothing in `cli/` is left after the above, delete `cli/` entirely (including `cli/Taskfile.yml`, `cli/go.mod`, `cli/go.sum`).

### 2.2 Drop QA skill mirrors from ARK

- [ ] Delete `.claude/skills/work-cli-qa/`.
- [ ] Delete `.codex/skills/work-cli-qa/`.

### 2.3 Rewire `task work` and `task work:qa`

In root `Taskfile.yml`:

- [ ] Replace `task work` body. Old: `go run -C cli ./cmd/work --store ../.work {{.CLI_ARGS}}`. New: `work --store .work {{.CLI_ARGS}}` with a `preconditions:` check that `command -v work` exists (`msg: "install work from https://github.com/gh-xj/work-cli"`).
- [ ] Delete `task work:qa` — that skill now lives with the source it gates.
- [ ] Delete the `verify` task's `task -d cli ci` step if `cli/` is gone.
- [ ] `verify` task body becomes just `bash scripts/verify.sh`. The script already PATH-detects `work` and asserts the operation prerequisites (handles the hard-fail case).

### 2.4 `.conventions.yaml` updates

- [ ] Add `min_work_version: "0.1.0"` field at root. (`scripts/verify.sh` already reads this and compares against `work --version`.)
- [ ] Update the `checks:` entries that mention `task -d cli ci` and the embedded `work` build — replace with checks against the external binary.
- [ ] The "sweep gate" check stays as-is.

### 2.5 Reframe the kit-side `work-cli` skill

- [ ] `skills/work-cli/SKILL.md` — keep the workflow content (inbox/triage/view), add a top-of-file install pointer to `https://github.com/gh-xj/work-cli`. Remove any wording that implies the binary ships with ARK.
- [ ] `adapters/claude-code/work-cli/SKILL.md`, `adapters/codex/work-cli/SKILL.md` — same reframing pass. Keep these adapters because the _integration_ (how the harness invokes the CLI) is still kit-owned.
- [ ] `adapters/cursor/work-cli.md` — already absent in current tree; nothing to do here.

### 2.6 Documentation sweep

- [ ] `AGENTS.md`, `CLAUDE.md` — change the "Work" convention bullet from "task work -- ..." (which used the embedded build) to "task work -- ..." backed by the external `work` binary; add an install hint.
- [ ] `README.md` — reframe `work` from "bundled" to "supported integration"; link to `gh-xj/work-cli`.
- [ ] `skills/convention-engineering/references/operations/work.md` — point to the external repo.

### 2.7 Phase 2 done = exit criteria

- `go build ./...` from ARK root succeeds (or no Go code remains in ARK).
- `which work` returns the externally-installed binary.
- `task verify` exits 0 from a clean checkout: `scripts/verify.sh` passes, including the `min_work_version` check.
- `task work -- view ready` produces the same output as before extraction.
- `grep -r 'go run.*cmd/work' .` returns no hits in tracked files.

---

## Phase 3 — Cleanup

- [ ] CHANGELOG / release-note entry on the next ARK release explaining the split. One paragraph.
- [ ] Move `docs/implementation/work-cli-v0.md` and `docs/implementation/tickets-to-work-migration.md` out of ARK (into `gh-xj/work-cli/docs/`) or delete them, since they document a tool that no longer lives here.

---

## Out of scope for this plan

- Backwards-compatibility shims. None needed; no external adopters.
- Renaming `agent-repo-kit`. That's a separate question (after extraction, ARK ships no binaries — worth revisiting whether the name still fits, but not blocking).
- Distributing `work` via Homebrew or other package managers.
- Any change to the `.work/` file format or store schema.

---

## Rollback

- **Phase 1 rollback:** delete the `gh-xj/work-cli` repo. ARK is unchanged.
- **Phase 2 rollback:** revert the single Phase 2 commit in ARK. The `gh-xj/work-cli` repo and its release stay; ARK goes back to building `work` from source.
