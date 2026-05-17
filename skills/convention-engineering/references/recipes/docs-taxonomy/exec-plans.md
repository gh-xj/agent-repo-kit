# Execution Plans

Status: stable
Codified: 2026-05-17

Use execution plans for work where sequencing, delegation, or handoff matters
enough that chat history is not durable enough.

## Invariants

1. The plan is checked in under the repo's declared docs root.
2. The plan states the goal, affected files or modules, ordered work steps,
   and verification gates.
3. Progress changes are recorded in the plan or in a linked implementation
   report before the work is considered done.
4. Stale plans do not shadow current work.
5. Known small debt is tracked in one flat debt file instead of many plan
   stubs.

## Placement

Use the docs-taxonomy `plans/` subtree. Low-volume repos may keep dated plan
files directly under `plans/`. Repos with repeated plan traffic should split
current and shipped work:

```text
plans/
|-- active/
|-- completed/
`-- tech-debt-tracker.md
```

`active/` contains only work that is still being executed. `completed/` keeps
shipped or abandoned plans with their final outcome. `tech-debt-tracker.md`
lists many small known debt items that do not each need a full execution plan.

## Relationship To Work Items

When a repo also uses `.work/`, the work item owns durable demand, priority,
and coordination leases. The execution plan owns the implementation sequence.
Link the work item to the plan when both exist, but do not duplicate the same
state in both places.

## Completion

Before moving a plan out of `active/`, record:

- the final outcome
- the verification commands or observations
- follow-up debt that did not fit the completed operation

If the repo has no active/completed split, the same information still belongs
in the plan or in a linked implementation report.
