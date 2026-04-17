# Stack Pattern: Python + uv

Advisory guidance for uv-backed Python projects. **Not lint-enforced.** The
core insight: `uv run` handles environment management for you, so the
Taskfile stays very small.

## Keep The Task Count Tiny

Aim for 1–5 user-facing tasks. If you are past 5, you are probably wrapping
things uv already does.

## One `uv run` Per Task

No separate "setup" or "venv" step. `uv run <cmd>` lazily syncs the env from
`pyproject.toml` + `uv.lock` on every invocation. Adding a `task setup` that
does `uv sync` is redundant.

```yaml
tasks:
  run:
    desc: Run the CLI
    cmds:
      - uv run main.py {{.CLI_ARGS}}
```

## Pass Config Via Taskfile `vars:`

Configurable runs should thread values through Taskfile vars into CLI args,
**not** env vars, and **not** inline Python. Users override with
`task run KEY=value`:

```yaml
vars:
  INPUT: ""
  MODEL: claude-opus

tasks:
  run:
    cmds:
      - uv run main.py --model={{.MODEL}} {{if .INPUT}}--input={{.INPUT}}{{end}} {{.CLI_ARGS}}
```

## Dependencies Live In `pyproject.toml`

Add/remove with `uv add <pkg>` / `uv remove <pkg>`. Commit `uv.lock`. Do not
call `pip install` from the Taskfile — it poisons the environment and
bypasses the lockfile.

## Recommended (Not Required) Python Stack

- [`typer`](https://typer.tiangolo.com/) for CLI parsing — fewer lines and
  better `--help` than `argparse`.
- [`loguru`](https://github.com/Delgan/loguru) for logging — zero-config,
  structured where needed.
- `if __name__ == "__main__":` guard on entry files. Typer picks it up
  cleanly and `python -m` still works.

None of these are enforced; swap in your own preferences.

## Anti-Patterns

- **Inline Python in Taskfile** — `cmds: [python -c "import x; x.run()"]`.
  Hard to test, hard to read, escapes poorly. Put code in a `.py` file.
- **`pip install` or manual `python -m venv`** — duplicates what uv already
  does. Breaks reproducibility when the lockfile is the source of truth.
- **`argparse` boilerplate for a 3-flag CLI** — typer is a one-decorator
  win. Use argparse only if you have a compelling reason.
- **Env vars for config** — `ENV_FOO=bar uv run ...`. Taskfile vars with
  CLI args are more discoverable (`task run FOO=bar` shows up in history
  and `--help` output).
- **`task setup` that shells to `uv sync`** — `uv run` already does this.
  Delete the task.
