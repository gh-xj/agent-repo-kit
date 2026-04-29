# Work CLI Operator Workflow

Use this workflow to operate `.work/` without taking ownership of the work CLI
implementation or convention design.

`.work/` is local operational state and should be ignored by the repo root
`.gitignore`.

## Store Discovery

Prefer the repo wrapper if it exists:

```bash
task work -- view ready
task work -- show W-0001
```

If there is no wrapper, use the binary directly:

```bash
work --store .work view ready
work --store .work show W-0001
```

If the repo is the `agent-repo-kit` checkout itself, the root wrapper runs the
source CLI:

```bash
task work -- view ready
```

## Lifecycle

Inbox entries are pre-commit capture. Work items are committed work.

```text
.work/inbox/IN-0001.yaml --triage accept--> .work/items/W-0001.yaml
```

Use inbox when the request is raw, duplicated, exploratory, or may not deserve
tracking:

```bash
work inbox add "Investigate failing smoke run" \
  --body "Captured from user report" \
  --source "chat"
```

Accept only when the repo should track it:

```bash
work triage accept IN-0001 --area qa --priority P1
```

Create directly when the work is already understood:

```bash
work new "Document migration path" --area docs --priority P2
```

## Reading Work

Use views for scans:

```bash
work view ready
work view active
work view blocked
work view done
```

Use show for exact lookup:

```bash
work show W-0001
work show IN-0001
```

Use JSON for automation:

```bash
work view ready --json
work show W-0001 --json
```

## Typed Work Items

`work init` installs the built-in `research` type when absent. Repos may define
additional work types under `.work/types/<type>/`. The type is a scaffold
lookup key, not a native class in the core work model.

```bash
work new "Research agent inbox UX" --type research
work triage accept IN-0002 --type research
```

The work item remains flat under `.work/items/W-NNNN.yaml`. Type-owned files
belong under `.work/spaces/W-NNNN/`.

## Verification

Before claiming completion, run the repo's verification gate. If the repo has
a Taskfile wrapper, prefer:

```bash
task verify
```

For this repository's work CLI implementation, use the maintainer QA harness:

```bash
task work:qa
```
