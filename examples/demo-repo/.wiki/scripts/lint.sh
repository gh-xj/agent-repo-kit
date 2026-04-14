#!/usr/bin/env bash
# Minimal wiki health check. Validates:
#   - Unique IDs per type prefix
#   - Filename matches frontmatter `id`
#   - Every [[S-NNN]] wikilink resolves to an existing source page
#   - Every source page has a `raw_path` that starts with `raw/` and exists

set -euo pipefail

PAGES_DIR=${PAGES_DIR:-pages}
RAW_DIR=${RAW_DIR:-raw}

errors=0
err() { printf 'ERROR: %s\n' "$1" >&2; errors=$((errors + 1)); }

# Collect all existing page IDs
declare -a ids=()
if ls "$PAGES_DIR"/*.md >/dev/null 2>&1; then
  for f in "$PAGES_DIR"/*.md; do
    base=$(basename "$f" .md)
    fm_id=$(awk -F': *' '/^id:/{print $2; exit}' "$f" | tr -d '"')
    fm_id_prefix="${base%%-*}-${base#*-}"; fm_id_prefix="${fm_id_prefix%%-*}"
    expected_id_in_name=$(printf '%s' "$base" | awk -F- '{print $1"-"$2}')

    if [ -z "$fm_id" ]; then
      err "$f: missing frontmatter id"
      continue
    fi
    if [ "$fm_id" != "$expected_id_in_name" ]; then
      err "$f: frontmatter id '$fm_id' does not match filename id '$expected_id_in_name'"
    fi
    ids+=("$fm_id")
  done
fi

# Check ID uniqueness
if [ "${#ids[@]}" -gt 0 ]; then
  dupes=$(printf '%s\n' "${ids[@]}" | sort | uniq -d)
  if [ -n "$dupes" ]; then
    while IFS= read -r d; do err "duplicate id: $d"; done <<<"$dupes"
  fi
fi

# Check wikilinks and raw_path
if ls "$PAGES_DIR"/*.md >/dev/null 2>&1; then
  for f in "$PAGES_DIR"/*.md; do
    # Every [[S-NNN]] or [[N-NNN]] must resolve to a page starting with that ID.
    # Tolerate "no matches" (grep exits 1) without tripping pipefail, and avoid a
    # subshell so err()'s increments to $errors are visible to the final check.
    refs=$(grep -oE '\[\[[SN]-[0-9]+' "$f" 2>/dev/null | sed 's/^\[\[//' | sort -u || true)
    if [ -n "$refs" ]; then
      while IFS= read -r ref; do
        [ -z "$ref" ] && continue
        if ! ls "$PAGES_DIR/$ref-"*.md >/dev/null 2>&1; then
          err "$f: wikilink [[$ref]] has no target page"
        fi
      done <<<"$refs"
    fi
  done
fi

# Enforce raw_path on source pages: required, must start with raw/, must exist.
if ls "$PAGES_DIR"/*.md >/dev/null 2>&1; then
  for f in "$PAGES_DIR"/*.md; do
    type=$(awk -F': *' '/^type:/{print $2; exit}' "$f")
    [ "$type" = "source" ] || continue
    raw=$(awk -F': *' '/^raw_path:/{print $2; exit}' "$f" | tr -d '"')
    if [ -z "$raw" ]; then
      err "$f: source page missing raw_path"
    elif [[ "$raw" != raw/* ]]; then
      err "$f: raw_path '$raw' must start with 'raw/'"
    elif [ ! -e "$raw" ]; then
      err "$f: raw_path '$raw' does not exist"
    fi
  done
fi

if [ $errors -eq 0 ]; then
  n_pages=$({ ls "$PAGES_DIR"/*.md 2>/dev/null || true; } | wc -l | tr -d ' ')
  n_raw=$(ls "$RAW_DIR" 2>/dev/null | { grep -v '^\.' || true; } | wc -l | tr -d ' ')
  printf 'OK: %s pages, %s raw sources\n' "$n_pages" "$n_raw"
  exit 0
fi

printf '\n%d error(s)\n' "$errors" >&2
exit 1
