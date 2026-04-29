#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
SKILLS_ROOT="$ROOT_DIR/skills"
AGENTS_ROOT="${HOME:?}/.agents/skills"
CLAUDE_ROOT="${HOME:?}/.claude/skills"

skill_names=()
shopt -s nullglob
for skill_file in "$SKILLS_ROOT"/*/SKILL.md; do
  skill_dir=${skill_file%/SKILL.md}
  skill_names+=("${skill_dir##*/}")
done
shopt -u nullglob

if [ "${#skill_names[@]}" -eq 0 ]; then
  printf 'no skill sources found under: %s\n' "$SKILLS_ROOT" >&2
  exit 1
fi

mkdir -p "$AGENTS_ROOT" "$CLAUDE_ROOT"

for name in "${skill_names[@]}"; do
  src="$SKILLS_ROOT/$name"
  dst="$AGENTS_ROOT/$name"

  if [ ! -d "$src" ]; then
    printf 'missing skill source: %s\n' "$src" >&2
    exit 1
  fi
  if [ -e "$dst" ] && [ ! -L "$dst" ]; then
    printf 'refusing to replace non-symlink path: %s\n' "$dst" >&2
    exit 1
  fi

  rm -f "$dst"
  ln -s "$src" "$dst"
done

for name in "${skill_names[@]}"; do
  src="$AGENTS_ROOT/$name"
  dst="$CLAUDE_ROOT/$name"

  if [ -e "$dst" ] && [ ! -L "$dst" ]; then
    printf 'refusing to replace non-symlink path: %s\n' "$dst" >&2
    exit 1
  fi

  rm -f "$dst"
  ln -s "$src" "$dst"
done

printf 'linked %d repo skills into %s\n' "${#skill_names[@]}" "$AGENTS_ROOT"
printf 'linked Claude compatibility entries into %s\n' "$CLAUDE_ROOT"
printf 'left %s untouched\n' "${HOME}/.codex/skills"
