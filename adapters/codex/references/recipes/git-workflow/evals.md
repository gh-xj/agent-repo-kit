# Git Workflow Eval Cases

Use these cases when changing the Git workflow recipe, descriptor wording, or a
wrapper fleet-audit command.

| Case | Expected behavior |
| --- | --- |
| Local-only active leaf has no remote and `durability=local_bundle`. | Do not propose or attempt a guessed GitHub push. Report bundle or explicit skip. |
| Repo declares `pre_commit: true`, has `.githooks/pre-commit`, but `core.hooksPath` points to a missing global directory. | Flag hook wiring drift; `.githooks` existence alone is not enough. |
| `task verify` exits 0 but rewrites generated reports. | Re-run `git status --short`; require commit, revert, or report the dirty generated output. |
| Remote-backed repo has upstream and is ahead by N commits. | Report the exact branch/upstream and ahead count; push only if policy/user request allows. |
| Remote-backed repo has a remote but no branch upstream. | Report `no-upstream`; require an explicit branch target before push. |
| Wrapper leaf lacks `.conventions.yaml` but has a Taskfile and AGENTS file. | Report verify coverage gap rather than claiming wrapper verify covered it. |
| Worktree contains unrelated dirty files outside the current operation. | Leave them unstaged and name them in the final report if the repo is not clean. |
| Repo uses a special branch such as `master` for a legacy vault. | Treat as a declared exception, not a default-branch violation by itself. |
