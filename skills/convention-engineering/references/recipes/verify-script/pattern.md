# Recipe: `scripts/verify.sh`

Status: stable
Codified: 2026-05-02

A small shell script that reads `.conventions.yaml` via `yq` and asserts each declared opt-in resolves to a real artifact. Pairs with the agent who interprets the free-form `checks:` list at audit / evaluation time.

## When To Use

Adopt this recipe when the repo declares any typed keys in `.conventions.yaml` and you want CI to fail on regressions without round-tripping through an agent. Skip it when the repo has only free-form `checks:` and audits run agent-only.

## Why

`.conventions.yaml` has two shapes of rule:

- **Typed keys** with documented semantics: `agent_docs`, `docs_root`, `taskfile`, `pre_commit`, `skill_roots`, `operations`, `core_beliefs`, `epic`. Each has exactly one truth check.
- **Free-form `checks:`** entries written in prose. Their truth depends on judgment that an agent applies during audit.

Without a script, the agent is the only gate. CI without an agent is silent; nightly runs cannot tell whether opt-ins regressed. The script gives CI a deterministic floor; the agent stays the ceiling.

## Boundary Rule

A typed key earns a script assertion. A free-form `checks:` entry does not. If a free-form rule deserves machine enforcement, **promote it to a typed key first** — extend the schema, document the key, then add the assertion. Do not script-assert prose; the script becomes a parser of English and rots.

## Minimal Recipe

```bash
#!/usr/bin/env bash
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

[ -f .conventions.yaml ] || { echo "verify: .conventions.yaml missing" >&2; exit 1; }
command -v yq >/dev/null 2>&1 || { echo "verify: yq required" >&2; exit 1; }

fail=0
fail() { echo "verify: $*" >&2; fail=1; }

# agent_docs[]: each declared file exists.
while IFS= read -r doc; do
  [ -z "$doc" ] && continue
  [ -f "$doc" ] || fail "agent_docs: missing $doc"
done < <(yq '.agent_docs[]? // ""' .conventions.yaml)

# docs_root + canonical taxonomy subdirs.
docs_root=$(yq -r '.docs_root // ""' .conventions.yaml)
if [ -n "$docs_root" ]; then
  for sub in requests planning plans implementation taxonomy; do
    [ -d "$docs_root/$sub" ] || fail "docs_root: $docs_root/$sub missing"
  done
fi

# core_beliefs trio.
if [ "$(yq -r '.core_beliefs.enabled // .core_beliefs // false' .conventions.yaml)" = "true" ]; then
  beliefs_path=$(yq -r '.core_beliefs.path // .docs_root // "docs"' .conventions.yaml)
  for f in core-belief.md invariants.md anti-patterns.md; do
    [ -f "$beliefs_path/$f" ] || fail "core_beliefs: $beliefs_path/$f missing"
  done
fi

# … one block per typed key …

[ "$fail" -eq 0 ] || exit 1
echo "verify: opt-ins ok"
```

The full reference implementation lives in `agent-repo-kit/scripts/verify.sh`.

## Properties

- **Pure shell + `yq`.** No language runtime. Drops into any repo.
- **One assertion per declared opt-in.** Missing key → silently skip (key was opt-out). Present key → assert exactly once.
- **Loud failures.** Each gap prints `verify: <category>: <what>`; CI output reads top-to-bottom without grep.
- **Idempotent.** No state written. Safe to run on every commit.
- **Short.** ~80 lines for the typed-key set. If it grows past ~150 lines, the descriptor's typed-key set is probably too ambitious; trim.

## Adopt

1. Drop in `scripts/verify.sh`. Start with the recipe above.
2. Add a Taskfile target: `task verify` runs `bash scripts/verify.sh`.
3. Wire it into CI. The script must exit non-zero on any gap.
4. As the descriptor's typed-key set grows, add one assertion block per new key in the same shape.

## What the agent still owns

- Every entry under `checks:`. The agent reads the prose, applies the rule against the live repo, reports findings.
- Graded scoring (judgment beyond pass/fail) — see `references/meta/evaluation-rubric.md`.
- Refactor-grade audits where typed assertions pass but the structure is still wrong.

## Anti-Patterns

- **Asserting prose.** Don't grep for sentences from the `checks:` list. Let the agent interpret it.
- **Embedding policy.** Don't hard-code paths the descriptor doesn't declare.
- **Silent skips of unknown ops.** When `operations:` declares a value the script doesn't know about, log "unknown op" rather than succeed-silently or hard-fail. Opt-ins evolve faster than the script.
