#!/usr/bin/env bash
# Deterministic floor of `task verify`: assert opt-ins declared in
# .conventions.yaml exist on disk. Free-form `checks:` prose entries are
# left to the convention-engineering agent on demand.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

if [ ! -f .conventions.yaml ]; then
  echo "verify: .conventions.yaml is missing" >&2
  exit 1
fi

if ! command -v yq >/dev/null 2>&1; then
  echo "verify: yq is required (https://github.com/mikefarah/yq)" >&2
  exit 1
fi

fail=0
fail() { echo "verify: $*" >&2; fail=1; }

# agent_docs[]: each declared file must exist.
while IFS= read -r doc; do
  [ -z "$doc" ] && continue
  [ -f "$doc" ] || fail "agent_docs: missing $doc"
done < <(yq '.agent_docs[]? // ""' .conventions.yaml)

# docs_root: directory exists with the canonical taxonomy subdirs.
docs_root=$(yq -r '.docs_root // ""' .conventions.yaml)
if [ -n "$docs_root" ]; then
  [ -d "$docs_root" ] || fail "docs_root: $docs_root is not a directory"
  for sub in requests planning plans implementation taxonomy; do
    [ -d "$docs_root/$sub" ] || fail "docs_root: $docs_root/$sub missing"
  done
fi

# taskfile: a Taskfile.yml exists with a verify target.
if [ "$(yq -r '.taskfile // false' .conventions.yaml)" = "true" ]; then
  [ -f Taskfile.yml ] || fail "taskfile: Taskfile.yml missing"
  grep -q '^[[:space:]]*verify:' Taskfile.yml || fail "taskfile: verify task not declared"
fi

# pre_commit: a hook script is wired up.
if [ "$(yq -r '.pre_commit // false' .conventions.yaml)" = "true" ]; then
  hooks_path=$(git config --get core.hooksPath || true)
  if [ -n "$hooks_path" ] && [ -x "$hooks_path/pre-commit" ]; then
    :
  elif [ -x .githooks/pre-commit ]; then
    :
  else
    fail "pre_commit: no executable pre-commit hook found"
  fi
fi

# skill_roots[]: each declared root exists.
while IFS= read -r root; do
  [ -z "$root" ] && continue
  [ -d "$root" ] || fail "skill_roots: $root missing"
done < <(yq '.skill_roots[]? // ""' .conventions.yaml)

# operations[]: each declared operation has its prerequisites.
while IFS= read -r op; do
  [ -z "$op" ] && continue
  case "$op" in
    work)
      command -v work >/dev/null 2>&1 \
        || fail "operations.work: 'work' binary not on PATH (install from https://github.com/gh-xj/work-cli)"
      [ -f .work/config.yaml ] || fail "operations.work: .work/config.yaml missing"

      min=$(yq -r '.min_work_version // ""' .conventions.yaml)
      if [ -n "$min" ] && command -v work >/dev/null 2>&1; then
        cur=$(work version 2>/dev/null | awk '{print $NF}' | sed 's/^v//' || true)
        if [ -n "$cur" ]; then
          # lexical compare is good enough for semver MAJOR.MINOR.PATCH
          # without prerelease suffixes; refine if/when we ship those.
          if [ "$(printf '%s\n%s\n' "$min" "$cur" | sort -V | head -n1)" != "$min" ]; then
            fail "operations.work: work version $cur < required $min"
          fi
        fi
      fi
      ;;
    wiki)
      [ -d .wiki ] || fail "operations.wiki: .wiki/ missing"
      ;;
    *)
      echo "verify: unknown operation '$op' (no built-in check)" >&2
      ;;
  esac
done < <(yq '.operations[]? // ""' .conventions.yaml)

# epic.leaves[]: each declared leaf must exist as a sibling dir AND be linked
# from this repo under repo/<leaf>. See references/patterns/epic-wrapper.md.
while IFS= read -r leaf; do
  [ -z "$leaf" ] && continue
  if [ ! -d "../$leaf" ]; then
    fail "epic.leaves: ../$leaf not found — clone it as a sibling and run 'task bootstrap'"
    continue
  fi
  link="repo/$leaf"
  if [ ! -L "$link" ]; then
    fail "epic.leaves: ./$link is not a symlink — run 'task bootstrap' to recreate it"
    continue
  fi
  # Resolve and ensure it points at the sibling.
  target=$(readlink "$link")
  if [ "$target" != "../../$leaf" ]; then
    fail "epic.leaves: ./$link -> $target (expected ../../$leaf)"
  fi
done < <(yq '.epic.leaves[]? // ""' .conventions.yaml)

if [ "$fail" -ne 0 ]; then
  exit 1
fi
echo "verify: opt-ins ok"
