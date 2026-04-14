# adapters/codex/

**TBD — contributions welcome.**

The Codex adapter will expose `contract/` and `evaluator/` as Codex agent
definitions (typically TOML under `~/.codex/agents/`). See
`../claude-code/` for the pattern: a thin shim pointing at the harness-free
source, plus install.sh wiring.
