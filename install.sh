#!/bin/sh
# install.sh — install agent-repo-kit into a supported harness
#
# Default flow: download the prebuilt ark binary for your OS/arch from the
# latest GitHub release, verify its SHA-256 checksum, install it to --prefix,
# then shell out to `ark adapters link` to wire skill symlinks into the
# detected harness.
#
# Pass --from-source to build from the local checkout instead (requires Go
# and a .git directory). Pass --skip-symlinks to install the binary only.

set -eu

DIR=$(cd "$(dirname "$0")" && pwd)
TARGET=""
DRY_RUN=0
FROM_SOURCE=0
SKIP_SYMLINKS=0
PREFIX="$HOME/.local/bin"
MANIFEST="$DIR/adapters/manifest.json"
RELEASE_OWNER="gh-xj"
RELEASE_REPO="agent-repo-kit"

log()  { printf '[install] %s\n' "$*"; }
warn() { printf '[install] WARN: %s\n' "$*" >&2; }
die()  { printf '[install] ERROR: %s\n' "$*" >&2; exit 1; }

run() {
    if [ "$DRY_RUN" -eq 1 ]; then
        log "DRY-RUN: $*"
    else
        log "$*"
        eval "$@"
    fi
}

usage() {
    cat <<'EOF'
Usage: ./install.sh [options]

Options:
  --target <name>     install target: claude-code or none (auto-detected)
  --prefix <dir>      binary install directory (default: ~/.local/bin)
  --manifest <path>   adapter manifest (default: <repo>/adapters/manifest.json)
  --from-source       build ark from local checkout instead of downloading
  --skip-symlinks     install binary only; skip the adapter link step
  --dry-run           preview actions without making changes
  -h, --help          show this message and exit

Default behavior:
  - Auto-detect target: ~/.claude/skills exists -> claude-code, else none.
  - Strategy: if --from-source, or (go on PATH AND .git present) -> source
              else -> download latest release archive and verify checksum.
EOF
}

while [ $# -gt 0 ]; do
    case "$1" in
        --target=*)   TARGET="${1#*=}";   shift ;;
        --prefix=*)   PREFIX="${1#*=}";   [ -n "$PREFIX" ]   || die "empty --prefix";   shift ;;
        --manifest=*) MANIFEST="${1#*=}"; [ -n "$MANIFEST" ] || die "empty --manifest"; shift ;;
        --target)     [ $# -ge 2 ] || die "missing value for --target";   TARGET="$2";   shift 2 ;;
        --prefix)     [ $# -ge 2 ] && [ -n "$2" ] || die "missing value for --prefix";   PREFIX="$2";   shift 2 ;;
        --manifest)   [ $# -ge 2 ] && [ -n "$2" ] || die "missing value for --manifest"; MANIFEST="$2"; shift 2 ;;
        --from-source)   FROM_SOURCE=1;   shift ;;
        --skip-symlinks) SKIP_SYMLINKS=1; shift ;;
        --dry-run)       DRY_RUN=1;       shift ;;
        -h|--help)       usage; exit 0 ;;
        *) die "unknown arg: $1" ;;
    esac
done

detect_os() {
    case "$(uname -s)" in
        Darwin) OS=darwin ;;
        Linux)  OS=linux ;;
        *) die "unsupported OS: $(uname -s) (expected Darwin or Linux)" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)  ARCH=amd64 ;;
        arm64|aarch64) ARCH=arm64 ;;
        *) die "unsupported arch: $(uname -m)" ;;
    esac
}

# auto-detect target (unchanged semantics from v1: lines 62-74)
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

resolve_strategy() {
    if [ "$FROM_SOURCE" -eq 1 ]; then
        STRATEGY=source
    elif command -v go >/dev/null 2>&1 && [ -d "$DIR/.git" ]; then
        STRATEGY=source
    else
        STRATEGY=download
    fi
    log "install strategy: $STRATEGY"
}

_sha256() {
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$@"
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$@"
    else
        warn "no sha256 tool found; skipping checksum verification"
        return 0
    fi
}

verify_checksum() {
    checksums="$1"; archive_name="$2"; workdir="$3"
    if ! command -v sha256sum >/dev/null 2>&1 && ! command -v shasum >/dev/null 2>&1; then
        warn "no sha256 tool found; skipping checksum verification"
        return 0
    fi
    grep "  ${archive_name}\$" "$checksums" >"$workdir/expected.txt" \
        || die "no checksum entry for $archive_name in checksums.txt"
    ( cd "$workdir" && _sha256 --check --status expected.txt ) \
        || die "checksum mismatch for $archive_name"
    log "checksum verified: $archive_name"
}

install_binary() {
    if [ "$STRATEGY" = source ]; then
        command -v go >/dev/null 2>&1 \
            || die "go toolchain not found; remove --from-source or install Go"
        run "(cd \"$DIR/cli\" && mkdir -p \"$PREFIX\" && go build -o \"$PREFIX/ark\" .)"
        return 0
    fi

    api_url="https://api.github.com/repos/${RELEASE_OWNER}/${RELEASE_REPO}/releases/latest"
    if [ "$DRY_RUN" -eq 1 ]; then
        VERSION="<latest>"
        log "DRY-RUN: curl -sfL $api_url | grep tag_name | sed ... -> VERSION"
    else
        VERSION=$(curl -sfL "$api_url" 2>/dev/null \
            | grep '"tag_name"' \
            | sed -e 's/.*"tag_name"[[:space:]]*:[[:space:]]*"v\{0,1\}//' -e 's/".*//' \
            | head -n1) || VERSION=""
        [ -n "$VERSION" ] || die "could not determine latest release version from $api_url"
        log "latest release: v$VERSION"
    fi

    ARCHIVE="ark-${VERSION}-${OS}-${ARCH}.tar.gz"
    base_url="https://github.com/${RELEASE_OWNER}/${RELEASE_REPO}/releases/download/v${VERSION}"
    TMPDIR=$(mktemp -d 2>/dev/null || mktemp -d -t ark-install)
    trap 'rm -rf "$TMPDIR"' EXIT INT TERM

    run "curl -sfL \"${base_url}/${ARCHIVE}\" -o \"$TMPDIR/$ARCHIVE\""
    run "curl -sfL \"${base_url}/checksums.txt\" -o \"$TMPDIR/checksums.txt\""
    if [ "$DRY_RUN" -eq 1 ]; then
        log "DRY-RUN: verify_checksum $TMPDIR/checksums.txt $ARCHIVE"
    else
        verify_checksum "$TMPDIR/checksums.txt" "$ARCHIVE" "$TMPDIR"
    fi
    run "tar -xzf \"$TMPDIR/$ARCHIVE\" -C \"$TMPDIR\""
    run "mkdir -p \"$PREFIX\""
    run "mv \"$TMPDIR/ark\" \"$PREFIX/ark.new\""
    run "mv \"$PREFIX/ark.new\" \"$PREFIX/ark\""
    run "chmod +x \"$PREFIX/ark\""
}

link_adapters() {
    if [ "$SKIP_SYMLINKS" -eq 1 ]; then
        log "skipping adapter link step (--skip-symlinks)"
        return 0
    fi
    if [ "$TARGET" = none ]; then
        return 0
    fi
    ARK_BIN=$(command -v ark 2>/dev/null || true)
    [ -n "$ARK_BIN" ] || ARK_BIN="$PREFIX/ark"
    if [ "$DRY_RUN" -eq 0 ] && [ ! -x "$ARK_BIN" ]; then
        die "ark binary not found at $ARK_BIN"
    fi
    dry_flag=""
    [ "$DRY_RUN" -eq 1 ] && dry_flag=" --dry-run"
    run "\"$ARK_BIN\" adapters link --target \"$TARGET\" --manifest \"$MANIFEST\" --repo-root \"$DIR\"$dry_flag"
}

main() {
    detect_os
    detect_arch
    log "platform: $OS/$ARCH"
    resolve_strategy
    install_binary

    case "$TARGET" in
        claude-code)
            link_adapters
            log "done. ark installed at $PREFIX/ark"
            log "restart your Claude Code session to pick up the skills."
            ;;
        none)
            cat <<EOF
No harness detected (or --target none). To adopt manually:

  1. Ensure $PREFIX is on PATH (or copy ark onto PATH yourself).
  2. Scaffold the tracked contract into your repo:
       ark init --repo-root /path/to/your-repo
  3. Read convention-engineering/ and convention-evaluator/ for the
     full convention and scoring surfaces.
  4. In the target repo, run: task verify
EOF
            log "done. ark installed at $PREFIX/ark"
            ;;
        *)
            die "unknown target: $TARGET (expected claude-code|none)"
            ;;
    esac
}

main
