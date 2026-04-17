#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
SOURCE_TICKETS_DIR=$(cd "$SCRIPT_DIR/.." && pwd)

PASS_COUNT=0
FAIL_COUNT=0

fail() {
  printf 'FAIL: %s\n' "$1" >&2
  FAIL_COUNT=$((FAIL_COUNT + 1))
}

pass() {
  printf 'PASS: %s\n' "$1"
  PASS_COUNT=$((PASS_COUNT + 1))
}

new_workspace() {
  local ws
  ws=$(mktemp -d)
  mkdir -p "$ws/.tickets"
  # Copy only structural parts; never copy live tickets, locks, or audit log.
  cp "$SOURCE_TICKETS_DIR/Taskfile.yml" "$ws/.tickets/Taskfile.yml"
  cp -R "$SOURCE_TICKETS_DIR/harness" "$ws/.tickets/harness"
  if [ -f "$SOURCE_TICKETS_DIR/README.md" ]; then
    cp "$SOURCE_TICKETS_DIR/README.md" "$ws/.tickets/README.md"
  fi
  if [ -f "$SOURCE_TICKETS_DIR/.gitignore" ]; then
    cp "$SOURCE_TICKETS_DIR/.gitignore" "$ws/.tickets/.gitignore"
  fi
  mkdir -p "$ws/.tickets/all"
  printf '# Audit Log\n\n' > "$ws/.tickets/audit-log.md"
  printf '%s\n' "$ws"
}

assert_file_contains() {
  local file=$1
  local expected=$2
  if ! grep -Fq "$expected" "$file"; then
    fail "expected '$expected' in $file"
    return 1
  fi
}

assert_file_not_contains() {
  local file=$1
  local unexpected=$2
  if grep -Fq "$unexpected" "$file"; then
    fail "did not expect '$unexpected' in $file"
    return 1
  fi
}

test_rejects_invalid_transition() {
  local ws output
  ws=$(new_workspace)
  (
    cd "$ws"
    task -d .tickets create -- "Smoke ticket" --priority P1 --category bug --estimated 30m >/dev/null
  )

  if output=$(
    cd "$ws" &&
      task -d .tickets transition -- T-01 --to REVIEW 2>&1
  ); then
    fail "invalid transition OPEN -> REVIEW should fail"
    printf '%s\n' "$output" >&2
    return 1
  fi

  pass "rejects invalid transition"
}

test_rejects_unknown_state() {
  local ws output
  ws=$(new_workspace)
  (
    cd "$ws"
    task -d .tickets create -- "Smoke ticket" --priority P1 --category bug --estimated 30m >/dev/null
  )

  if output=$(
    cd "$ws" &&
      task -d .tickets transition -- T-01 --to WHATEVER 2>&1
  ); then
    fail "unknown state should fail"
    printf '%s\n' "$output" >&2
    return 1
  fi

  pass "rejects unknown state"
}

test_rejects_invalid_category() {
  local ws output
  ws=$(new_workspace)

  if output=$(
    cd "$ws" &&
      task -d .tickets create -- "Category test" --priority P1 --category not-a-real-category --estimated 30m 2>&1
  ); then
    fail "invalid category should fail"
    printf '%s\n' "$output" >&2
    return 1
  fi

  pass "rejects invalid category"
}

test_rejects_invalid_priority() {
  local ws output
  ws=$(new_workspace)

  if output=$(
    cd "$ws" &&
      task -d .tickets create -- "Priority test" --priority HIGH --category bug --estimated 30m 2>&1
  ); then
    fail "invalid priority should fail"
    printf '%s\n' "$output" >&2
    return 1
  fi

  pass "rejects invalid priority"
}

test_cancelled_ticket_records_cancel_reason() {
  local ws ticket_file
  ws=$(new_workspace)
  (
    cd "$ws"
    task -d .tickets create -- "Close test" --priority P1 --category bug --estimated 30m >/dev/null
    task -d .tickets close -- T-01 --reason "test close" >/dev/null
  )

  ticket_file=$(find "$ws/.tickets/all" -name ticket.md | head -1)
  assert_file_contains "$ticket_file" "status: CANCELLED" || return 1
  assert_file_contains "$ticket_file" 'cancel_reason: "test close"' || return 1
  assert_file_not_contains "$ticket_file" "resolved:" || return 1

  pass "cancelled ticket records cancel reason"
}

test_done_ticket_records_resolved_date() {
  local ws ticket_file
  ws=$(new_workspace)
  (
    cd "$ws"
    task -d .tickets create -- "Done test" --priority P1 --category bug --estimated 30m >/dev/null
    task -d .tickets transition -- T-01 --to IN_PROGRESS >/dev/null
    task -d .tickets transition -- T-01 --to DONE --note "completed" >/dev/null
  )

  ticket_file=$(find "$ws/.tickets/all" -name ticket.md | head -1)
  assert_file_contains "$ticket_file" "status: DONE" || return 1
  assert_file_contains "$ticket_file" "resolved:" || return 1
  assert_file_not_contains "$ticket_file" "cancel_reason:" || return 1

  pass "done ticket records resolved date"
}

test_list_output_format() {
  local ws output
  ws=$(new_workspace)
  (
    cd "$ws"
    task -d .tickets create -- "List format ticket" --priority P2 --category bug --estimated 30m >/dev/null
  )

  output=$(cd "$ws" && task -d .tickets list 2>&1)
  # Header row must include all four column labels
  if ! printf '%s\n' "$output" | grep -Eq '^ID +STATUS +PRI +CATEGORY +TITLE'; then
    fail "list header missing expected columns; got: $output"
    return 1
  fi
  # Ticket row must have ID, status, priority, category, title all populated (no blank columns)
  if ! printf '%s\n' "$output" | grep -Eq '^T-01 +OPEN +P2 +bug +List format ticket'; then
    fail "list row missing/blank columns; got: $output"
    return 1
  fi

  pass "list output format preserves all columns"
}

test_unsafe_title_is_yaml_escaped() {
  local ws ticket_file
  ws=$(new_workspace)
  # Title contains embedded newline, backslash, and double-quote — must be YAML-safe.
  local bad_title='Evil"Title\with
newline'
  (
    cd "$ws"
    task -d .tickets create -- "$bad_title" --priority P1 --category bug --estimated 30m >/dev/null
  )

  ticket_file=$(find "$ws/.tickets/all" -name ticket.md | head -1)
  # Title line must be single-line and begin with escaped quote sequence.
  if ! grep -Eq '^title: "Evil\\"Title\\\\with newline"$' "$ticket_file"; then
    fail "title not properly YAML-escaped; got:"
    grep '^title:' "$ticket_file" >&2 || true
    return 1
  fi
  # Frontmatter must remain a valid 2-delimiter block (exactly two `---` lines).
  local delim_count
  delim_count=$(grep -c '^---$' "$ticket_file")
  if [ "$delim_count" -ne 2 ]; then
    fail "frontmatter delimiters broken by unsafe title; found $delim_count lines"
    return 1
  fi

  pass "unsafe title is YAML-escaped"
}

test_multiline_note_sanitized_in_audit_log() {
  local ws audit baseline post added
  ws=$(new_workspace)
  (
    cd "$ws"
    task -d .tickets create -- "Note sanitize" --priority P1 --category bug --estimated 30m >/dev/null
  )
  audit="$ws/.tickets/audit-log.md"
  baseline=$(wc -l < "$audit" | tr -d ' ')
  # Inject a note with newlines that would forge extra audit rows if not sanitized.
  (
    cd "$ws"
    task -d .tickets transition -- T-01 --to IN_PROGRESS \
      --note "real note
2099-01-01  T-99  FORGED -> DONE" >/dev/null
  )
  post=$(wc -l < "$audit" | tr -d ' ')
  added=$((post - baseline))
  # Sanitized transition should add exactly one audit-log line.
  if [ "$added" -ne 1 ]; then
    fail "expected 1 new audit line, got $added"
    grep -n '' "$audit" >&2
    return 1
  fi
  # No line may start with the forged date prefix (would indicate a successful
  # audit-row injection via embedded newline in --note).
  if grep -Eq '^2099-' "$audit"; then
    fail "audit-log contains injected date-prefixed line"
    return 1
  fi
  # The sanitized note should still be recorded as a single concatenated line
  # on the legitimate transition row (not split across multiple rows).
  if ! grep -Eq '^[0-9]{4}-[0-9]{2}-[0-9]{2}.*T-01.*OPEN -> IN_PROGRESS.*real note' "$audit"; then
    fail "legitimate transition row not recorded as expected"
    grep -n '' "$audit" >&2
    return 1
  fi

  pass "multiline note is sanitized before audit-log write"
}

test_concurrent_new_produces_unique_ids() {
  local ws ids unique_count total
  ws=$(new_workspace)
  (
    cd "$ws"
    task -d .tickets create -- "Race A" --priority P1 --category bug --estimated 30m >/dev/null &
    task -d .tickets create -- "Race B" --priority P1 --category bug --estimated 30m >/dev/null &
    task -d .tickets create -- "Race C" --priority P1 --category bug --estimated 30m >/dev/null &
    task -d .tickets create -- "Race D" --priority P1 --category bug --estimated 30m >/dev/null &
    wait
  )
  # Collect all allocated IDs from ticket directories.
  ids=$(ls -1 "$ws/.tickets/all" 2>/dev/null | sed -nE 's/^(T-[0-9]+)-.*/\1/p' | sort)
  total=$(printf '%s\n' "$ids" | grep -c '^T-' || true)
  unique_count=$(printf '%s\n' "$ids" | sort -u | grep -c '^T-' || true)
  if [ "$total" -ne 4 ] || [ "$unique_count" -ne 4 ]; then
    fail "concurrent creates produced total=$total unique=$unique_count (expected 4/4)"
    printf '%s\n' "$ids" >&2
    return 1
  fi

  pass "concurrent task new produces unique IDs"
}

main() {
  test_rejects_invalid_transition
  test_rejects_unknown_state
  test_rejects_invalid_category
  test_rejects_invalid_priority
  test_cancelled_ticket_records_cancel_reason
  test_done_ticket_records_resolved_date
  test_list_output_format
  test_unsafe_title_is_yaml_escaped
  test_multiline_note_sanitized_in_audit_log
  test_concurrent_new_produces_unique_ids

  printf '\nSummary: %s passed, %s failed\n' "$PASS_COUNT" "$FAIL_COUNT"

  if [ "$FAIL_COUNT" -gt 0 ]; then
    exit 1
  fi
}

main "$@"
