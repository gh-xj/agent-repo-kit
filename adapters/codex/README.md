# adapters/codex/

No packaged Codex adapter is shipped yet.

Current Codex skill installs live under `~/.agents/skills/<skill>/SKILL.md`,
but this repo does not yet provide a ready-to-install wrapper there.

If you add one, keep it thin: point it at the repo-root
`convention-engineering/` and `convention-evaluator/` surfaces rather than
duplicating content. `install.sh` does not support `--target codex` today.
