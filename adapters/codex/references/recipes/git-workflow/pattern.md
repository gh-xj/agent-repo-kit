# Recipe: Git Workflow

Status: experimental
Codified: 2026-06-22

Reusable Git discipline for agent-operated repos: inspect first, stage narrowly,
verify before completion, and push only through a declared durability policy.

## When To Use

Adopt when a repo or wrapper coordinates agent commits, generated outputs,
private data, local-only leaves, or multiple remotes.

Skip when the repo is read-only, externally owned, or has no agent commit
workflow. For one-off external checkouts, keep the convention local in
`.git/info/exclude` if needed rather than committing project policy.

## The Convention

Every agent Git operation starts with:

```bash
git status --short --branch
```

The agent identifies the operation boundary before staging. Stage explicit paths
only:

```bash
git add -- path/to/file another/path
git diff --cached
```

Each commit is one review boundary: one behavior, policy, data refresh,
generated-output refresh, or documentation decision. Split commits when either
part could be reverted or reviewed independently. Combining is acceptable only
when a split would leave one commit broken.

Run the repo's canonical gate before claiming completion:

```bash
task verify
git status --short
```

If verification writes generated artifacts, either commit the intended generated
output in the owning repo or fix the gate so it is read-only. Do not report
"verified and clean" until the post-verify status has been inspected.

Push only when the repo's declared policy allows it. Do not infer a remote name
from the directory or GitHub org. Repos without a remote use their declared
backup path, such as git bundles, instead of a guessed push destination.

The final operation report names:

- repo
- commit hash
- staged paths or path groups
- verification command and result
- push, bundle, or explicit skip reason
- any remaining dirty paths and why they remain uncommitted

## Adopt

1. Add the workflow to `AGENTS.md` or point to a repo doc when it is too long.
2. Add repo-specific policy under `.conventions.yaml`.
3. Add a read-only audit task when the repo coordinates more than one checkout.
4. Keep hard enforcement in checks; keep human judgment in prose.

Suggested extension:

```yaml
git_policy:
  default_branch: main
  durability: remote # remote | local_bundle | dropbox_remote | none
  push_policy: declared_upstream_only # declared_upstream_only | bundle_only | explicit_only
  hooks: repo_local_required # repo_local_required | external_allowed | none
  post_verify_clean: required # required | manual
  commit_subject: domain_imperative # domain_imperative | conventional | custom
```

For wrappers, put per-leaf durability in the wrapper registry and audit the
live checkout state against it.

## Anti-Patterns

- Guessing `origin` or a GitHub repository from the directory name.
- Using `git add .` when unrelated dirty files exist.
- Treating local hooks as sufficient verification.
- Claiming a wrapper verified every leaf when the leaf was skipped for missing
  `.conventions.yaml` or `Taskfile.yml`.
- Hiding warning-only verify output in the final report.

## Eval Cases

See `evals.md`.
