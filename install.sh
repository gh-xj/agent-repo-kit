#!/bin/sh
# install.sh — install agent-repo-kit into a detected harness
#
# Usage:
#   ./install.sh [--target claude-code|codex|none] [--dry-run]
#
# Default: auto-detect.
#   - ~/.claude/skills/ exists             -> claude-code
#   - ~/.codex/ exists (and no Claude)     -> codex
#   - neither                              -> none (prints instructions)

set -eu

DIR=$(cd "$(dirname "$0")" && pwd)
TARGET=""
DRY_RUN=0

log()  { printf '[install] %s\n' "$*"; }
warn() { printf '[install] WARN: %s\n' "$*" >&2; }
die()  { printf '[install] ERROR: %s\n' "$*" >&2; exit 1; }

while [ $# -gt 0 ]; do
    case "$1" in
        --target)   TARGET="${2:-}"; shift 2 ;;
        --target=*) TARGET="${1#--target=}"; shift ;;
        --dry-run)  DRY_RUN=1; shift ;;
        -h|--help)
            sed -n '2,10p' "$0" | sed 's/^# \{0,1\}//'
            exit 0
            ;;
        *) die "unknown arg: $1" ;;
    esac
done

if [ -z "$TARGET" ]; then
    if [ -d "$HOME/.claude/skills" ]; then
        TARGET="claude-code"
    elif [ -d "$HOME/.codex" ]; then
        TARGET="codex"
    else
        TARGET="none"
    fi
    log "auto-detected target: $TARGET"
else
    log "using explicit target: $TARGET"
fi

run() {
    if [ "$DRY_RUN" -eq 1 ]; then
        log "DRY-RUN: $*"
    else
        log "$*"
        eval "$@"
    fi
}

link() {
    src="$1"; dst="$2"
    if [ -e "$dst" ] || [ -L "$dst" ]; then
        warn "$dst already exists — skipping (remove it to re-link)"
        return 0
    fi
    run "mkdir -p \"$(dirname "$dst")\""
    run "ln -s \"$src\" \"$dst\""
}

case "$TARGET" in
    claude-code)
        DEST="$HOME/.claude/skills"
        log "installing symlinks into $DEST/"
        link "$DIR/contract"  "$DEST/convention-engineering"
        link "$DIR/evaluator" "$DEST/convention-evaluator"
        log "done. restart your Claude Code session to pick up the skills."
        ;;
    codex)
        log "codex adapter is a placeholder — see adapters/codex/README.md"
        log "no files were installed."
        ;;
    none)
        cat <<EOF
No harness detected (or --target none). To adopt manually:

  1. Copy examples/demo-repo/.tickets and .wiki into your repo.
  2. Paste the "## Conventions" block from examples/demo-repo/AGENTS.md
     into your repo's AGENTS.md and CLAUDE.md (dual-write).
  3. Read contract/ for the full convention surface.
  4. Run:  task -d .tickets test   and   task -d .wiki lint
EOF
        ;;
    *) die "unknown target: $TARGET (expected claude-code|codex|none)" ;;
esac
