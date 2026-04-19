# Python Convention Profile

Conventions for Python service/tool repositories.

## Table of Contents

- [Build System: Taskfile.dev + uv](#build-system-taskfiledev--uv)
- [Package Management: uv](#package-management-uv)
- [Linting: ruff](#linting-ruff)
- [Type Checking: mypy or pyright](#type-checking-mypy-or-pyright)
- [Testing: pytest](#testing-pytest)
- [Pre-Commit Hook](#pre-commit-hook)
- [File Organization](#file-organization)

## Build System: Taskfile.dev + uv

```yaml
# Taskfile.yml
tasks:
  setup:
    desc: Install dependencies
    cmds: [uv sync]
  test:
    desc: Run tests
    cmds: [uv run pytest]
  lint:
    desc: Run linter
    cmds: [uv run ruff check .]
  fmt:
    desc: Format code
    cmds: [uv run ruff format .]
  typecheck:
    desc: Type check
    cmds: [uv run mypy .]
```

## Package Management: uv

- Single config file: `pyproject.toml` (PEP 621)
- Lockfile: `uv.lock` (must be committed)
- Replaces pip, poetry, pipenv -- faster, more reliable

```toml
# pyproject.toml
[project]
name = "my-service"
version = "0.1.0"
requires-python = ">=3.12"
dependencies = [
    "fastapi>=0.100",
    "pydantic>=2.0",
]

[tool.uv]
dev-dependencies = [
    "pytest>=8.0",
    "ruff>=0.4",
    "mypy>=1.10",
]
```

## Linting: ruff

Replaces flake8 + isort + black in a single fast tool:

```toml
# pyproject.toml
[tool.ruff]
line-length = 88
target-version = "py312"

[tool.ruff.lint]
select = ["E", "F", "I", "N", "UP", "B", "C90"]

[tool.ruff.lint.mccabe]
max-complexity = 15
```

## Type Checking: mypy or pyright

Strict mode:

```toml
[tool.mypy]
strict = true
warn_return_any = true
warn_unused_configs = true
```

## Testing: pytest

- Fixtures in `conftest.py` (shared test setup)
- Coverage: `pytest-cov` with minimum threshold
- Structure: `tests/` mirroring `src/` layout

```toml
[tool.pytest.ini_options]
testpaths = ["tests"]
addopts = "--strict-markers -v"
```

## Pre-Commit Hook

```bash
#!/bin/bash
set -e
uv run ruff format .
uv run ruff check --fix .
uv run mypy .
```

## File Organization

```
my-service/
├── src/my_service/     # Application code
│   ├── __init__.py
│   ├── api/            # HTTP endpoints
│   ├── service/        # Business logic
│   ├── dal/            # Data access
│   └── model/          # Domain types
├── tests/              # Test files (mirrors src/)
├── pyproject.toml      # Single config file
├── uv.lock             # Lockfile
├── Taskfile.yml        # Task automation
└── AGENTS.md           # Agent instructions
```
