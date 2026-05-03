# Work CLI

Operate the local-first `.work/` tracker as an agent-facing work ledger.

## When To Use

Use this skill when the task is to:

- Inspect or operate a repo's `.work/` state.
- Capture an untriaged request with `work inbox add`.
- Promote accepted work with `work triage accept`.
- Create a direct work item with `work new`.
- Show work records or scan views with `work show` and `work view`.
- Use a typed work item with `--type`, including scaffolded item spaces.

Do not use this skill for:

- Changing `cmd/work`, `internal/work`, or `internal/workcli`. Those live in
  the external [`gh-xj/work-cli`](https://github.com/gh-xj/work-cli) repo;
  edit them there.
- Redesigning the `.work/` filesystem contract.
- Adding verification gates or repo convention docs.

Use `convention-engineering` for convention design and `go-scripting` for CLI
implementation work.

## Install

`work` is an external dependency. Install once per machine:

```bash
go install github.com/gh-xj/work-cli/cmd/work@latest
```

Or grab a tarball release from
https://github.com/gh-xj/work-cli/releases. The repo's
`.conventions.yaml` may declare a `min_work_version` that
`scripts/verify.sh` enforces.

## First Actions

1. Find the repo root and check whether the work store config exists.
2. Prefer the repo wrapper when present:

```bash
task work -- view ready
```

3. Otherwise call the binary directly:

```bash
work --store .work view ready
```

4. Use `--json` whenever another tool or agent will parse the result.

## Operating Loop

Read `references/operator-workflow.md` before making changes to `.work/`.

Default flow:

1. Capture uncertain requests in the inbox.
2. Triage only when the repo should track the work.
3. Work from canonical records under `.work/items/*.yaml`.
4. Put item-owned notes and small supporting artifacts under
   `.work/spaces/<W-ID>/` when a work item needs a directory.
5. Use views as derived read models, not as storage.
6. Run the repo's verification gate before claiming completion.

## Storage Model

- Inbox entry: unaccepted demand signal under `.work/inbox/IN-NNNN.yaml`.
- Work item: accepted canonical record under `.work/items/W-NNNN.yaml`.
- Work space: optional item-owned directory under `.work/spaces/W-NNNN/` for
  notes, research captures, plans, and other supporting files.
- Work type: optional scaffold/template under `.work/types/<type>/` that can
  create an initial work space for typed items.

## Core Commands

```bash
work init
work inbox
work inbox add "Title" --body "Context" --source "user"
work triage accept IN-0001 --area cli --priority P1
work new "Title" --area docs --priority P2
work view ready
work show W-0001
```

## Boundaries

- Say "work item" in user-facing text; avoid bare "item" when clarity matters.
- Inbox entries are demand signals, not accepted work.
- Work items are the durable source of truth.
- Work spaces support a work item, but do not replace the canonical item YAML.
- Large logs, browser captures, and bulky evidence payloads belong outside
  `.work/`; store only pointers in the work item or its space.
- Do not create legacy compatibility paths unless the user explicitly asks for
  migration documentation.
