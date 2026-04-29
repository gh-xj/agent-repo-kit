# install-v2: Installation Overhaul Architecture Blueprint

**Status:** Phase 2 — Architecture
**Date:** 2026-04-18
**Workstreams:** W1 (release pipeline), W2 (install.sh v2), W3 (adapters subcommand), W4 (upgrade subcommand), W5 (ark init wizard)

> 2026-04-29 implementation note: this is a historical architecture plan.
> The current installer ships both `ark` and `work`; skills are installed
> separately through the open `skills` CLI. Some sections below preserve the
> original design discussion rather than the final shipped behavior.

---

## §1. Release Artifact Naming Scheme

**Archive filename template:** `ark-{version}-{os}-{arch}.tar.gz`
`{version}` = semver without `v` prefix (e.g., `0.4.0`). `{os}` = `darwin` | `linux`. `{arch}` = `amd64` | `arm64`.

**Checksum file:** `checksums.txt` — single file per release, SHA-256, `sha256sum`-compatible format: `<hash>  <filename>` per line.

**Download URL shape:**

```
https://github.com/gh-xj/agent-repo-kit/releases/download/v{version}/ark-{version}-{os}-{arch}.tar.gz
https://github.com/gh-xj/agent-repo-kit/releases/download/v{version}/checksums.txt
```

**Archive contents:** binaries `ark` and `work` + `LICENSE`. README omitted to prevent stale docs drift.

**POSIX-portable checksum detection in install.sh:**

```sh
_sha256() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$@"
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$@"
  else
    warn "no sha256 tool found; skipping checksum verification"; return 0
  fi
}
# verify:
grep "ark-${VERSION}-${OS}-${ARCH}.tar.gz" "$TMPDIR/checksums.txt" \
  | (cd "$TMPDIR" && _sha256 --check --status) \
  || die "checksum mismatch for ark-${VERSION}-${OS}-${ARCH}.tar.gz"
```

---

## §2. `.goreleaser.yml` Shape

Full file, commit at repo root:

```yaml
version: 2

project_name: ark

before:
  hooks:
    - go mod tidy

builds:
  - id: ark
    binary: ark
    dir: cli
    gomod:
      dir: cli
    main: .
    env:
      - CGO_ENABLED=0
    flags:
      - -trimpath
    ldflags:
      - -s -w
      - -X github.com/gh-xj/agent-repo-kit/cli/cmd.appVersion={{.Version}}
      - -X github.com/gh-xj/agent-repo-kit/cli/cmd.appCommit={{.Commit}}
      - -X github.com/gh-xj/agent-repo-kit/cli/cmd.appDate={{.Date}}
    goos:
      - darwin
      - linux
    goarch:
      - amd64
      - arm64

archives:
  - id: ark
    builds: [ark]
    name_template: "ark-{{ .Version }}-{{ .Os }}-{{ .Arch }}"
    format: tar.gz
    files:
      - LICENSE

checksum:
  name_template: "checksums.txt"
  algorithm: sha256

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^chore:"
      - "^ci:"
      - Merge pull request
      - Merge branch
```

**Required change in `cli/cmd/root.go:14-19`** — `const` block → split into `const` + `var`:

```go
// Before:
const (
    binaryName = "ark"
    appVersion = "dev"
    appCommit  = "none"
    appDate    = "unknown"
)

// After:
const binaryName = "ark"

var (
    appVersion = "dev"
    appCommit  = "none"
    appDate    = "unknown"
)
```

`binaryName` stays `const`; `appVersion`/`appCommit`/`appDate` become `var` so `-X` ldflags overwrite them at link time. No other files reference these identifiers — they are consumed only at `cli/cmd/root.go:106-109` (inside `newLeafCmd`) and `cli/cmd/version.go:17-20`.

---

## §3. GitHub Actions Release Workflow

**File:** `.github/workflows/release.yml`

```yaml
name: release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version-file: cli/go.mod
          cache-dependency-path: cli/go.sum

      - uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

`GITHUB_TOKEN` is sufficient — it is auto-provisioned by GitHub Actions with `contents: write` permission. No additional secrets required.

**Helper script:** `scripts/tag-release.sh`

```sh
#!/bin/sh
# Usage: scripts/tag-release.sh <version>
# Example: scripts/tag-release.sh 0.4.0
set -eu
VERSION="${1:?usage: tag-release.sh <version>}"
git tag -a "v${VERSION}" -m "release v${VERSION}"
git push origin "v${VERSION}"
printf '[tag-release] pushed v%s\n' "$VERSION"
```

---

## §4. Adapter Manifest Schema

**File:** `adapters/manifest.json`

JSON chosen over YAML: no runtime `yq` dependency, parseable by Go's stdlib `encoding/json`, and the Go helper `ark adapters list-links` is the single consumer — `install.sh` never needs to parse JSON itself.

**Schema:**

```json
{
  "schema_version": 1,
  "harnesses": [
    {
      "name": "claude-code",
      "skill_root": "~/.claude/skills",
      "links": [
        {
          "source": "convention-engineering",
          "dest": "convention-engineering"
        },
        { "source": "convention-evaluator", "dest": "convention-evaluator" },
        { "source": "skill-builder", "dest": "skill-builder" },
        { "source": "taskfile-authoring", "dest": "taskfile-authoring" }
      ]
    }
  ]
}
```

Fields:

- `schema_version` — integer, must equal 1.
- `harnesses[].name` — matches `--target` flag values.
- `harnesses[].skill_root` — destination directory; `~` expanded at runtime.
- `harnesses[].links[].source` — path relative to repo root.
- `harnesses[].links[].dest` — symlink name relative to `skill_root`.

Note: current `install.sh` only links three of these; `taskfile-authoring` is a fourth skill directory that exists in the repo but is not yet wired. Adding it to the manifest fixes that drift.

**Go structs:** new package `cli/internal/adapters`, file `cli/internal/adapters/manifest.go`:

```go
package adapters

import (
    "encoding/json"
    "os"
    "path/filepath"
    "strings"
)

type Manifest struct {
    SchemaVersion int       `json:"schema_version"`
    Harnesses     []Harness `json:"harnesses"`
}

type Harness struct {
    Name      string `json:"name"`
    SkillRoot string `json:"skill_root"`
    Links     []Link `json:"links"`
}

type Link struct {
    Source string `json:"source"`
    Dest   string `json:"dest"`
}

func Load(path string) (*Manifest, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var m Manifest
    return &m, json.Unmarshal(data, &m)
}

func (h *Harness) ExpandSkillRoot() (string, error) {
    if strings.HasPrefix(h.SkillRoot, "~/") {
        home, err := os.UserHomeDir()
        if err != nil {
            return "", err
        }
        return filepath.Join(home, h.SkillRoot[2:]), nil
    }
    return h.SkillRoot, nil
}
```

---

## §5. `install.sh` v2 Flow

New flags added to existing flag parser:

```
--from-source      force source build even when prebuilt download is available
--prefix <dir>     binary install directory (default: ~/.local/bin)
--skip-symlinks    skip adapter link step
--manifest <path>  adapter manifest (default: $DIR/adapters/manifest.json)
```

`--dry-run`, `--target`, existing `log`/`warn`/`die`/`run` helpers are preserved unchanged.

**Default `--prefix`:** `~/.local/bin`. Justification: no sudo required; XDG standard user bin dir; present on PATH in most modern Linux/macOS setups. Users needing system-wide install pass `--prefix /usr/local/bin`.

**Pseudocode for full flow:**

```sh
detect_os()   # uname -s: Darwin→darwin, Linux→linux; else die
detect_arch() # uname -m: x86_64→amd64, arm64|aarch64→arm64; else die

resolve_strategy():
  if FROM_SOURCE=1 OR (command -v go AND [ -d "$DIR/.git" ]):
    STRATEGY=source
  else:
    STRATEGY=download

install_binary():
  if STRATEGY=source:
    run "(cd '$DIR/cli' && mkdir -p '$PREFIX' && go build -o '$PREFIX/ark' .)"
  else:
    VERSION=$(curl -sfL https://api.github.com/repos/gh-xj/agent-repo-kit/releases/latest \
              | grep '"tag_name"' | sed 's/.*"v\([^"]*\)".*/\1/')
    [ -n "$VERSION" ] || die "could not determine latest release version"
    ARCHIVE="ark-${VERSION}-${OS}-${ARCH}.tar.gz"
    TMPDIR=$(mktemp -d); trap 'rm -rf "$TMPDIR"' EXIT
    run "curl -sfL '.../releases/download/v${VERSION}/${ARCHIVE}' -o '$TMPDIR/$ARCHIVE'"
    run "curl -sfL '.../releases/download/v${VERSION}/checksums.txt' -o '$TMPDIR/checksums.txt'"
    verify_checksum "$TMPDIR/checksums.txt" "$TMPDIR/$ARCHIVE"
    run "tar -xzf '$TMPDIR/$ARCHIVE' -C '$TMPDIR'"
    run "mkdir -p '$PREFIX'"
    run "mv '$TMPDIR/ark' '$PREFIX/ark.new'"
    run "mv '$PREFIX/ark.new' '$PREFIX/ark'"  # atomic same-fs rename
    run "chmod +x '$PREFIX/ark'"

link_adapters():
  [ "$SKIP_SYMLINKS" -eq 1 ] && return 0
  ARK_BIN=$(command -v ark 2>/dev/null || echo "$PREFIX/ark")
  [ -x "$ARK_BIN" ] || die "ark binary not found at $ARK_BIN"
  run "$ARK_BIN adapters link \
    --target '$TARGET' \
    --manifest '$MANIFEST' \
    --repo-root '$DIR' \
    $([ $DRY_RUN -eq 1 ] && echo '--dry-run')"

main():
  detect_os; detect_arch
  resolve_target  # existing auto-detect logic
  resolve_strategy
  install_binary
  link_adapters
  log "done. ark installed at $PREFIX/ark"
  [ "$TARGET" = "claude-code" ] && log "restart Claude Code to pick up skills."
```

---

## §6. `ark` Go-Side Changes

### 6a. New subcommand: `ark adapters link`

**New files:**

- `cli/cmd/adapters.go` — registers `adapters` group + `link` and `list-links` subcommands; follows same registration pattern as `newSkillCmd()` at `cli/cmd/root.go:75`.
- `cli/internal/adapters/link.go` — `RunLink(repoRoot, manifestPath, target string, dryRun bool) error`.

`link` flags: `--target` (required), `--manifest` (default `adapters/manifest.json` relative to `--repo-root`), `--repo-root` (default `.`), `--dry-run`.

`RunLink` logic:

1. `adapters.Load(manifestPath)`
2. Find harness by name; return `fmt.Errorf("unknown target %q", target)` if not found.
3. `harness.ExpandSkillRoot()`
4. For each link: `srcAbs = filepath.Join(repoRoot, link.Source)` — verify exists; `dstAbs = filepath.Join(skillRoot, link.Dest)` — call `ensureSymlink`.
5. `ensureSymlink(src, dst, dryRun)`: existing symlink → remove + recreate; non-symlink file/dir → warn + skip; absent → create. Mirrors `install.sh:85-98` semantics exactly.

`list-links` subcommand prints `<srcAbs>\t<dstAbs>` per line, one per link entry. Used as belt-and-suspenders for shell consumers.

`adapters` group is registered in `registerBuiltins()` at `cli/cmd/root.go:51`.

### 6b. New subcommand: `ark upgrade`

**New files:**

- `cli/cmd/upgrade.go` — command registration.
- `cli/internal/upgrade/upgrade.go` — `DetectFlavor`, `RunUpgrade`.

Flags: `--target` (default: auto-detect same as install.sh), `--dry-run`, `--prefix` (default: directory of `os.Executable()`).

**Flavor detection** (`cli/internal/upgrade/upgrade.go`):

```go
type Flavor int
const (FlavorClone Flavor = iota; FlavorPrebuilt)

func DetectFlavor(selfPath string) Flavor {
    dir := filepath.Dir(selfPath)
    for {
        if fi, err := os.Stat(filepath.Join(dir, ".git")); err == nil && fi.IsDir() {
            return FlavorClone
        }
        parent := filepath.Dir(dir)
        if parent == dir { break }
        dir = parent
    }
    return FlavorPrebuilt
}
```

**Clone upgrade sequence:**

1. Resolve `repoRoot` = highest `.git` parent of `selfPath`.
2. `exec.Command("git", "-C", repoRoot, "pull", "--ff-only")` — run, stream output.
3. `exec.Command("go", "build", "-o", selfPath+".new", ".")` in `repoRoot/cli/`.
4. `os.Rename(selfPath+".new", selfPath)`.
5. `adapters.RunLink(target, manifest, repoRoot, dryRun)`.

**Prebuilt upgrade sequence:**

1. Fetch latest tag from GitHub API; compare to `appVersion`; if equal, log + return.
2. Download archive + checksum to `os.MkdirTemp`.
3. Verify checksum.
4. `os.Rename(selfPath, selfPath+".old")`.
5. Extract `ark` from archive to `selfPath+".new"`.
6. `os.Rename(selfPath+".new", selfPath)`.
7. Remove `selfPath+".old"` on success; on any error, log that `.old` is the recovery point and return error.
8. `adapters.RunLink(target, manifest, resolvedRepoRoot, dryRun)`.

**Interrupted-upgrade recovery:** if `selfPath+".new"` exists at startup of `RunUpgrade`, resume from step 6 (rename .new → selfPath). If `selfPath+".old"` exists and `selfPath` is absent, rename `.old` → selfPath and return error with instructions to re-run upgrade.

### 6c. `ark init` wizard rewrite

**Dependency to add** in `cli/go.mod`: `github.com/charmbracelet/huh` (see §7 open question for pinning).

`init.go` gains `--non-interactive` flag. When `--non-interactive` is set, or `os.Stdin` is not a terminal (`!term.IsTerminal(int(os.Stdin.Fd()))`), the current flag-driven path runs unchanged.

When interactive, build a single `huh.NewForm(...)` with three groups:

**Group 1 — Overwrite guard** (skipped if no existing contract):

- `huh.NewConfirm().Title("Existing .convention-engineering.json found. Overwrite?")` → `&confirmOverwrite`; abort on false.

**Group 2 — Configuration:**

- `huh.NewMultiSelect[string]().Title("Profiles").Options(...)` with auto-detected value pre-selected; assigns `&profilesSelection`.
- `huh.NewMultiSelect[string]().Title("Operations").Options(work, wiki, taskfile).Value(&opsSelection)` with `work` pre-selected and `wiki` opt-in.
- `huh.NewSelect[string]().Title("Repo risk").Options(standard, elevated, critical).Value(&repoRisk)` with `standard` pre-selected.

**Group 3 — Confirmation:**

- `huh.NewConfirm().Title("Scaffold with the above settings?")` → `&confirmScaffold`; abort on false.

After `form.Run()` completes: construct `scaffold.Options` from selections, call `scaffold.RunInit(repoRoot, opts, stdout, stderr)` unchanged. Attempt `exec.Command("task", "verify")` as non-fatal smoke test. Print next-steps summary.

### 6d. ldflags wiring

Change `cli/cmd/root.go:14-19` as shown in §2. This is the only required file change for ldflags — the `-X` path `github.com/gh-xj/agent-repo-kit/cli/cmd.appVersion` matches the package path of that `var`.

---

## §7. Dependency Graph + Open Questions

**Dependency arrows:**

```
§1 (naming) ──► §2 (.goreleaser.yml)  ──► §3 (release.yml)
§1           ──► §5 (install.sh download URLs)
§4 (manifest) ──► §6a (adapters link Go impl)
§6a           ──► §5 (install.sh calls ark adapters link)
§6a           ──► §6b (upgrade calls adapters.RunLink)
§6d (var)     ──► §2 (ldflags target must be var)
§6c (wizard)   — independent of §1–§4; blocked only on huh dep addition
```

**W1** owns §1 + §2 + §3 + §6d. May start immediately.
**W3** owns §4 + §6a. May start immediately (no W1 dependency).
**W2** owns §5. Blocked on W3 (`ark adapters link` must exist to test link step).
**W4** owns §6b. Blocked on W3 (`adapters.RunLink` shared logic).
**W5** owns §6c. Independent; may start immediately.

**Open questions:**

1. Owner slug confirmed as `gh-xj`? (release URLs use `github.com/gh-xj/agent-repo-kit`)
2. Sign release archives with cosign, or skip entirely?
3. Default `--prefix`: `~/.local/bin` or interactive prompt?
4. Should `ark upgrade` re-render `check.sh` after binary relocation?
5. Pin `huh` to `v0.6.x` or track `v0.6` minor?
6. Add `codex` harness entry to manifest now or later?

---

## §8. Non-Goals (explicitly out of scope)

- Windows builds, MSYS2, or WSL install paths.
- Homebrew tap formula.
- npm/npx wrapper.
- Cosign or GPG archive signing.
- Auto-rollback beyond `.old` file recovery.
- Copy-based (non-symlink) skill installation.
- Cursor adapter wiring (placeholder docs only, no install wiring).
