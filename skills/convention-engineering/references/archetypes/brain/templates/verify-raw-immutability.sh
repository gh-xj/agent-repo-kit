#!/usr/bin/env bash
# verify-raw-immutability.sh — assert no tracked file under the captured realm
# has been modified after its first commit. README.md files are exempt.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

CAPTURED_REALM="${1:-raw}"

git rev-parse HEAD >/dev/null 2>&1 || { echo "no commits — skip"; exit 0; }
mapfile -t files < <(git ls-files "$CAPTURED_REALM/" 2>/dev/null || true)
[ "${#files[@]}" -eq 0 ] && { echo "$CAPTURED_REALM/ empty — skip"; exit 0; }

failed=0
for f in "${files[@]}"; do
  [[ "$f" == */README.md ]] && continue
  first=$(git log --diff-filter=A --follow --format=%H -- "$f" | tail -1)
  [ -z "$first" ] && continue
  first_blob=$(git rev-parse "${first}:${f}" 2>/dev/null || echo "")
  head_blob=$(git rev-parse "HEAD:${f}" 2>/dev/null || echo "")
  [ -z "$first_blob" ] || [ -z "$head_blob" ] && continue
  if [ "$first_blob" != "$head_blob" ]; then
    echo "FAIL: $f modified" >&2
    failed=1
  fi
done

[ "$failed" -ne 0 ] && exit 1
echo "OK: ${#files[@]} captured file(s) unchanged"
