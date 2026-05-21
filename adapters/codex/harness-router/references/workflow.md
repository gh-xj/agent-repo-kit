# Workflow

Use this workflow after meaningful work, a user correction, or an explicit
request to decide what should persist.

## 1. Gather Evidence

Read only what is needed:

- latest user correction or session summary
- changed files and verification output
- active work item and work space
- relevant instruction files
- related or used skills
- docs or references that already own the topic
- target-local rules for any surface you will write into, such as `RULES.md`,
  `README.md`, scaffold notes, or convention blocks

## 2. Extract Candidate Learnings

Write each learning as one durable sentence. Split unrelated ideas. Mark weak
or untrusted claims before routing them.

## 3. Classify

Apply `externalization-model.md` first:

- What cognitive burden would the artifact remove?
- What artifact class should carry that burden?
- Is the learning captured, staged, proposed, implemented, verified, rejected,
  or deprecated?

Then apply `routing-taxonomy.md` to each learning. If a dimension is unclear,
say so in the proposal instead of guessing.

## 4. Route

Use `target-surfaces.md` to choose the narrowest destination. Prefer:

- one canonical owner over duplicated instructions
- references over bloated router files
- checks/hooks/evals over prose when enforcement matters
- `.work` over docs for active unresolved follow-up
- target-local rules over generic assumptions when writing into a work space,
  docs area, skill, or generated surface

## 5. Propose

Use `proposal-format.md`. Include at least:

- source and date
- durable lesson
- candidate destination
- externalized burden and artifact class
- reason
- confidence
- risks

## 6. Stop Before Mutation

Do not edit durable surfaces unless the user approves the proposal or the user
already requested a specific edit.

When an edit is approved, re-read the target's local rules before writing. For
example, a typed `.work` space may require raw captures under `raw/`,
synthesized notes under `pages/`, and conclusions in `findings.md`.
