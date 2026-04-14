# convention-engineering

Unified repo convention knowledge with multi-stack profiles.

## Read Order

1. The convention doc (routing table)
2. Target reference for your question
3. Stack profiles for your repo's tech stack

## Run Contract Checker

```bash
SKILL_DIR="<path-to-this-convention>"
GO111MODULE=off go run "$SKILL_DIR/scripts" --repo-root . --json
```

This reads the tracked contract at `.convention-engineering.json`. The checker
now fails if that machine artifact is missing.

With an explicit tracked contract path:

```bash
GO111MODULE=off go run "$SKILL_DIR/scripts" --repo-root . --config .convention-engineering.json --json
```

With an overlay contract:

```bash
GO111MODULE=off go run "$SKILL_DIR/scripts" --repo-root . --config .docs/convention-engineering.overlay.json --json
```

## Run Orchestrated Evaluation

```bash
GO111MODULE=off go run "$SKILL_DIR/scripts" \
  --repo-root . \
  --orchestrate \
  --topic convention-run \
  --generated-artifacts README.md,OWNERSHIP.md
```

This writes the convention brief, raw evidence bundle, handoff manifest, and launcher receipt under `<docs_root>/planning/` and `<docs_root>/reviews/`, then launches `convention-evaluator` and leaves its report/result artifacts beside them.

## Core Checks

- Required files (CLAUDE.md, Taskfile.yml, etc.)
- Task tokens with include-aware taskfile resolution
- Canonical pointer contracts (`all` / `any` modes)
- Docs content markers
- `.git/info/exclude` pattern contracts (`git_exclude_checks`)
- Optional invariant contract checks

## Open-Source Local Overlay

Use `references/operations/open-source-git-exclude-workflow.md` when working in another/open-source repo with local-only files.

This mode uses `.git/info/exclude` + `.docs/` (instead of tracked `docs/` paths) for personal convention scaffolding.

## Config

- Primary tracked contract: `.convention-engineering.json`
- Overlay contract: `.docs/convention-engineering.overlay.json`
- Scoring surface: `convention-evaluator`
- Example: `references/convention-config.example.json`
- Schema: `references/config.schema.json`

Contract fields introduced by the checker:

- `contract_version` gates major compatibility and currently must be `1`
- `mode` is `tracked` or `overlay`
- `docs_root` is restricted to `docs` or `.docs`
- `ownership_policy` defines portable and repo-local authorship ownership
- `mirror_policy` defines mirrored-doc behavior for files such as `CLAUDE.md` and `AGENTS.md`
- `evaluation_inputs` is an evaluator-signal object; the minimum supported field is `repo_risk`
- `chunk_plan` uses the spec shape with `enabled` plus `chunks[]` records of `{id, scope, completion_criteria, depends_on}`
- additive fields are allowed within `contract_version: 1`; unknown major versions fail closed

## Structure

```
references/
├── core/           # Universal patterns (agent docs, docs taxonomy, arch contracts, gates, supply chain)
├── profiles/       # Stack-specific (Go, TypeScript/React, Python)
├── contracts/      # Verification gate contracts
└── operations/     # Audit and bootstrap workflows
scripts/            # Go contract checker CLI
```
