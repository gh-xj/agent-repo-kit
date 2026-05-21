#!/usr/bin/env bash
# Mirror canonical skills/<name>/ into per-harness adapter copies.
#
# Replaces the deleted ark skillsync runner. Drives off the live filesystem
# (every immediate skills/<name>/SKILL.md), not adapters/manifest.json — the
# manifest only lists harness skill roots, not the skill set.
#
# Modes:
#   --check   exit non-zero if any adapter copy differs from canonical
#   (none)    overwrite adapter copies to match canonical
#
# Asymmetry: convention-engineering is the "root" skill of each harness
# adapter (lives at adapters/<harness>/SKILL.md, not in a subdir). Every
# other skill mirrors at adapters/<harness>/<skill>/SKILL.md (claude-code,
# codex) or adapters/cursor/<skill>.md.
#
# Directory harnesses also mirror skill sidecars that the SKILL.md may route
# to: references/, schemas/, templates/, assets/, cli/, and maintenance/version
# markdown files. The flat Cursor placeholder remains a SKILL.md-only mirror.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

mode="write"
case "${1:-}" in
  --check) mode="check" ;;
  "") ;;
  *) echo "usage: $0 [--check]" >&2; exit 2 ;;
esac

# Harnesses that mirror per skill directory.
DIR_HARNESSES=("claude-code" "codex")
# Harness that mirrors per flat .md file.
FLAT_HARNESS="cursor"
# The one skill that lives at the adapter root rather than in a subdir.
ROOT_SKILL="convention-engineering"
SIDECARS=(
  references
  schemas
  templates
  assets
  cli
  MAINTENANCE.md
  maintenance.md
  CHANGELOG.md
  MIGRATION.md
)

drift=0
synced=0

sync_pair() {
  local src=$1 dst=$2
  mkdir -p "$(dirname "$dst")"
  if [ "$mode" = "check" ]; then
    if [ ! -f "$dst" ] || ! cmp -s "$src" "$dst"; then
      echo "drift: $dst" >&2
      drift=1
    fi
  else
    if [ ! -f "$dst" ] || ! cmp -s "$src" "$dst"; then
      cp "$src" "$dst"
      synced=$((synced + 1))
    fi
  fi
}

sync_sidecars() {
  local skill_dir=$1 dst_dir=$2
  local name src file rel

  for name in "${SIDECARS[@]}"; do
    src="$skill_dir/$name"
    [ -e "$src" ] || continue

    if [ -d "$src" ]; then
      while IFS= read -r file; do
        rel=${file#"$skill_dir/"}
        sync_pair "$file" "$dst_dir/$rel"
      done < <(find "$src" -type f | sort)
    elif [ -f "$src" ]; then
      sync_pair "$src" "$dst_dir/$name"
    fi
  done
}

for src in skills/*/SKILL.md; do
  [ -f "$src" ] || continue
  skill=$(basename "$(dirname "$src")")

  for harness in "${DIR_HARNESSES[@]}"; do
    if [ "$skill" = "$ROOT_SKILL" ]; then
      dst_dir="adapters/$harness"
    else
      dst_dir="adapters/$harness/$skill"
    fi

    sync_pair "$src" "$dst_dir/SKILL.md"
    sync_sidecars "skills/$skill" "$dst_dir"
  done

  sync_pair "$src" "adapters/$FLAT_HARNESS/$skill.md"
done

if [ "$mode" = "check" ]; then
  if [ "$drift" -ne 0 ]; then
    echo "sync-adapters: drift detected — run scripts/sync-adapters.sh" >&2
    exit 1
  fi
  echo "sync-adapters: in sync"
else
  echo "sync-adapters: synced $synced file(s)"
fi
