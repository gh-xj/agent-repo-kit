# Convention Audit Workflow

Use this workflow to check an existing repo against convention standards. Produces a gap report — no auto-fixes without user approval.

## Table of Contents

- [How to Run](#how-to-run)
- [Audit Checklist](#audit-checklist) — Tiers 1-4
- [Output Format](#output-format)
- [Taskfile Integration](#taskfile-integration)
- [Applying Fixes](#applying-fixes)

## How to Run

1. Navigate to the repo root.
2. Run the contract checker for programmatic checks (assumes `ark` is on `PATH` after `install.sh`; set `ARK_BINARY` to override):

```bash
ark check --repo-root . --json
```

For open-source/local overlay mode, use:

```bash
ark check --repo-root . --config .docs/convention-engineering.overlay.json --json
```

3. Walk through the manual checklist below for items the checker doesn't cover.

## Audit Checklist

### Tier 1: Agent Legibility (highest impact)

- [ ] `CLAUDE.md` exists with architecture contract + commands + verification gates
- [ ] `AGENTS.md` mirrors `CLAUDE.md` (or single CLAUDE.md if only one AI tool used)
- [ ] Architecture layer diagram exists (ASCII, in CLAUDE.md or ARCHITECTURE.md)
- [ ] One canonical verify command is documented and functional
- [ ] Commands cheat sheet lists exact commands (not descriptions)

### Tier 2: Mechanical Enforcement

- [ ] Linter config exists (`.golangci.yml` / `eslint.config.js` / `pyproject.toml [tool.ruff]`)
- [ ] Linter enforces layer boundaries (depguard / eslint-boundaries / import-linter)
- [ ] Pre-commit hook exists and runs: format + lint + dep tidy
- [ ] Lockfile exists and is committed (`go.sum` / `bun.lock` / `uv.lock`)
- [ ] Type checking enabled in strict mode

### Tier 3: Verification Gates

- [ ] `task verify` (or equivalent) runs all gates in one command
- [ ] Unit tests exist and pass
- [ ] Architecture boundary check is automated (not prose-only)
- [ ] Pre-commit hook completes in <10 seconds

### Tier 4: Knowledge Architecture (for repos with cross-repo impact)

- [ ] Cross-repo skill exists if service has 3+ consumers
- [ ] Reference docs have freshness contracts (before-trusting, after-discovering)
- [ ] Changelog maintained in skill and/or ARCHITECTURE.md
- [ ] Docs taxonomy root chosen and consistent (`docs/` or `.docs/`, not both active)
- [ ] RFI/design/plan lifecycle paths exist: `requests/`, `planning/`, `plans/` under chosen docs root
- [ ] (git-exclude mode) `.git/info/exclude` includes local overlay entries (`.docs`, `CLAUDE.local.md`, `AGENTS.override.md`, local `Taskfile.yml`)

### Tier 5: Operational Conventions (conditional)

Only audit each subsection if the corresponding directory exists. See
`references/operations/work.md` and `references/operations/wiki.md` for
rationale and layout details.

If `.work/` exists:

- [ ] `CLAUDE.md` / `AGENTS.md` contains the work pointer snippet
- [ ] `.work/config.yaml` and `.work/views.yaml` exist
- [ ] root `Taskfile.yml` exposes `work:` when a Taskfile wrapper is expected
- [ ] `work --store .work view ready --json` or `task work -- view ready --json` passes

If `.wiki/` exists:

- [ ] `CLAUDE.md` / `AGENTS.md` contains the wiki pointer snippet
- [ ] `.wiki/scripts/lint.sh` exists and `task -d .wiki lint` passes
- [ ] `.wiki/raw/` and `.wiki/pages/` directories exist

Flag any adopted op whose pointer snippet is missing from the agent contract
files as a gap.

## Output Format

After running the audit, produce a gap report:

```
## Convention Audit: [repo-name]

### Present (green)
- [x] CLAUDE.md with architecture contract
- [x] golangci-lint with depguard
- [x] Pre-commit hook
...

### Missing (needs work)
- [ ] AGENTS.md mirroring (single CLAUDE.md only)
- [ ] Entropy/drift detection
...

### Partial (needs improvement)
- [~] Verification gate exists but missing smoke/regress steps
...

### Recommendation
Priority fixes: [top 3 items that would have the highest impact]
```

## Taskfile Integration

Add a `check:conventions` task to any repo's Taskfile for automated checking:

```yaml
# In Taskfile.yml or taskfiles/core.yml
tasks:
  check:conventions:
    desc: Check repo convention compliance
    vars:
      ARK: '{{default "ark" .ARK_BINARY}}'
      CONFIG_FILE: '{{default ".convention-engineering.json" .CONFIG}}'
    cmds:
      - "{{.ARK}} check --repo-root . --config {{.CONFIG_FILE}}"
    preconditions:
      - sh: command -v {{.ARK}} >/dev/null 2>&1
        msg: "ark CLI not found on PATH (install agent-repo-kit via ./install.sh or set ARK_BINARY)"
```

Include in the canonical verify gate:

```yaml
verify:
  cmds:
    - task: lint
    - task: test
    - task: check:conventions # convention compliance
    - task: smoke # if applicable
```

## Applying Fixes

After presenting the gap report, ask the user which gaps to address. Then:

1. Check which stack profiles apply (Go, TypeScript/React, Python, Research Corpus)
2. Read the relevant profile docs for specific patterns
3. Generate the missing files following profile conventions
4. Run the verify command to confirm everything works
