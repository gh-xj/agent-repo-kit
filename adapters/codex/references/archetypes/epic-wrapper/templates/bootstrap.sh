#!/usr/bin/env bash
# Recreate the sibling symlinks declared in epic.leaves under repo/.
# Idempotent. Skips a leaf cleanly if its sibling dir is missing.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

command -v yq >/dev/null 2>&1 || {
  echo "bootstrap: yq required (https://github.com/mikefarah/yq)" >&2
  exit 1
}

mapfile -t leaves < <(yq '.epic.leaves[]? // ""' .conventions.yaml | sed '/^$/d')

if [ "${#leaves[@]}" -eq 0 ]; then
  echo "bootstrap: no epic.leaves declared in .conventions.yaml" >&2
  exit 1
fi

mkdir -p repo

for leaf in "${leaves[@]}"; do
  if [ ! -d "../$leaf" ]; then
    echo "warn: ../$leaf not found -- clone it as a sibling and re-run" >&2
    continue
  fi
  link="repo/$leaf"
  if [ -e "$link" ] && [ ! -L "$link" ]; then
    echo "skip: $link exists and is not a symlink -- leaving it alone" >&2
    continue
  fi
  ln -sfn "../../$leaf" "$link"
  echo "linked: $link -> ../../$leaf"
done
