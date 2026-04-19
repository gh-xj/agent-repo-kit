# Supply Chain Hygiene

Dependency management patterns per stack. The goal: reproducible builds, known vulnerabilities surfaced, licenses auditable.

## Lockfile Expectations

| Stack | Lockfile | Auto-Generated | Must Commit |
|-------|----------|----------------|-------------|
| Go | `go.sum` | Yes (go mod tidy) | Yes |
| TypeScript (Bun) | `bun.lock` | Yes (bun install) | Yes |
| TypeScript (npm) | `package-lock.json` | Yes (npm install) | Yes |
| TypeScript (Yarn) | `yarn.lock` | Yes (yarn install) | Yes |
| Python (uv) | `uv.lock` | Yes (uv lock) | Yes |
| Python (poetry) | `poetry.lock` | Yes (poetry lock) | Yes |

Never `.gitignore` lockfiles. They are the reproducibility contract.

## Vulnerability Scanning

| Stack | Tool | Command | When to Run |
|-------|------|---------|-------------|
| Go | govulncheck | `govulncheck ./...` | Before release, weekly CI |
| TypeScript | npm audit | `npm audit` / `bun audit` | Before release, weekly CI |
| Python | pip-audit | `pip-audit` | Before release, weekly CI |
| Python | safety | `safety check` | Alternative to pip-audit |

## License Checking

| Stack | Tool | Command |
|-------|------|---------|
| Go | go-licenses | `go-licenses check ./...` |
| TypeScript | license-checker | `npx license-checker --summary` |
| Python | liccheck | `liccheck -s strategy.ini` |

## Pinning Strategy

- **Go**: Module system handles exact pinning via `go.sum`. No additional config needed.
- **TypeScript**: Lockfile pins exact versions. Use `--save-exact` for new deps.
- **Python**: `uv.lock` or `poetry.lock` pins exact versions. Avoid pip freeze without a lockfile.

## System Dependencies

Document required system-level dependencies in setup tasks:
- C libraries (e.g., libyara for YARA bindings)
- Runtime requirements (e.g., Node.js version, Go version)
- Optional services (e.g., MongoDB for persistence)

Use a `task setup` command that installs or verifies all prerequisites.

## Pre-Commit Dep Tidy

Include dependency tidying in pre-commit hooks:
- Go: `go mod tidy` (removes unused, adds missing)
- TypeScript: Not typically needed (lockfile managed by install)
- Python: `uv lock` if using uv
