#!/bin/sh
# install.sh — install agent-repo-kit into a supported harness
#
# Usage:
#   ./install.sh [--target claude-code|none] [--dry-run]
#
# Default: auto-detect.
#   - ~/.claude/skills exists -> claude-code
#   - otherwise               -> none (prints manual instructions)

set -eu

DIR=$(cd "$(dirname "$0")" && pwd)
TARGET=""
DRY_RUN=0

log()  { printf '[install] %s\n' "$*"; }
warn() { printf '[install] WARN: %s\n' "$*" >&2; }
die()  { printf '[install] ERROR: %s\n' "$*" >&2; exit 1; }

while [ $# -gt 0 ]; do
    case "$1" in
        --target)
            [ $# -ge 2 ] || die "missing value for --target"
            TARGET="${2:-}"
            [ -n "$TARGET" ] || die "missing value for --target"
            shift 2
            ;;
        --target=*)
            TARGET="${1#--target=}"
            [ -n "$TARGET" ] || die "missing value for --target"
            shift
            ;;
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
    else
        TARGET="none"
        if [ -d "$HOME/.codex" ] || [ -d "$HOME/.agents/skills" ]; then
            log "detected Codex state, but no packaged Codex adapter is shipped yet"
        fi
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
    if [ ! -e "$src" ]; then
        die "source path missing: $src"
    fi
    if [ -L "$dst" ]; then
        warn "$dst is a symlink — re-linking"
        run "rm \"$dst\""
    elif [ -e "$dst" ]; then
        warn "$dst already exists and is not a symlink — skipping (remove it to re-link)"
        return 0
    fi
    run "mkdir -p \"$(dirname "$dst")\""
    run "ln -s \"$src\" \"$dst\""
}

case "$TARGET" in
    claude-code)
        DEST="$HOME/.claude/skills"
        log "installing symlinks into $DEST/"
        link "$DIR/convention-engineering" "$DEST/convention-engineering"
        link "$DIR/convention-evaluator" "$DEST/convention-evaluator"
        link "$DIR/skill-builder" "$DEST/skill-builder"
        log "done. restart your Claude Code session to pick up the skills."
        ;;
    codex)
        die "unsupported target: codex (no packaged Codex adapter yet; see adapters/codex/README.md)"
        ;;
    cursor)
        die "unsupported target: cursor (no packaged Cursor adapter yet; see adapters/cursor/README.md)"
        ;;
    none)
        cat <<EOF
No harness detected (or --target none). To adopt manually:

  1. From this checkout, bootstrap your repo with:
     (cd "$DIR" && GO111MODULE=off go run ./convention-engineering/scripts \\
       --repo-root /path/to/your-repo --init)
  2. Read convention-engineering/ and convention-evaluator/ for the full
     convention and scoring surfaces.
  3. In the target repo, run: task verify
  4. If you later move this checkout, set:
     CONVENTION_ENGINEERING_DIR="$DIR/convention-engineering"
EOF
        ;;
    *) die "unknown target: $TARGET (expected claude-code|none)" ;;
esac
