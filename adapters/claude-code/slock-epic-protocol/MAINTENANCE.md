# Slock Epic Protocol Maintenance

This skill is self-evolving. When a real Slock/epic run exposes protocol drift,
update the smallest owning file and record the lesson here in the same change.

## When To Update

Update this skill when:

- an agent writes durable protocol into Slock memory or notes
- a repo cannot verify its Slock agent map
- a Slock agent follows `repo` but cannot discover the right skills or `.work`
- a handoff depends on Slock history instead of target repo inbox/work
- a bootstrap/check script pattern repeats across projects
- a user corrects this skill's assumptions

## Entry Format

```markdown
### YYYY-MM-DD - short title

**Trigger**: what real run exposed the issue.
**Change**: files changed and rule added/removed.
**Verification**: command or inspection that proves the fix.
**Falsifier**: what would show the rule is wrong or incomplete.
**Regression case**: future prompt/state that should now behave differently.
```

## Lessons

### 2026-05-06 - Pointer-only Slock memory must be exact

**Trigger**: A federated Slock/epic project had correct `repo` symlinks, but
mapped Slock `MEMORY.md` files accumulated daily scan logs and active handoff
rules. The descriptor-level check still passed.

**Change**: The protocol now requires generated `MEMORY.md` to match the
pointer-only form exactly during local Slock checks.

**Verification**: `task slock:check` must fail on memory drift and pass after
regeneration from the descriptor.

**Falsifier**: If a mapped agent genuinely needs non-pointer startup behavior
that cannot live in repo docs, skills, or work, the pointer-only rule is too
strict and needs a named extension field.

**Regression case**: A future audit finds `## Recent scans` or `## Current
pipeline rule` in generated memory; the skill should move the rule into the
repo and regenerate memory, not bless the drift.

### 2026-05-06 - Slock agents need runtime skill mirrors in the repo

**Trigger**: A leaf repo had a Claude-local skill but no Codex-compatible mirror,
while Slock agents could run through different runtimes.

**Change**: The protocol now requires declared skill roots for every runtime
used by mapped agents, with symlink or generated mirrors for the same portable
skill intent.

**Verification**: Leaf verification should check declared skill roots exist, and
manual inspection should confirm `SKILL.md` resolves through each mirror.

**Falsifier**: If a project pins every mapped agent to one runtime and documents
that boundary, missing mirrors are not drift.

**Regression case**: A mapped Codex agent follows `repo` and cannot discover the
repo-local skill; the skill should add `.agents/skills` or document why Codex is
not a supported runtime for that leaf.

### 2026-05-06 - Migrate Slock notes by ownership, not by copying

**Trigger**: Slock `notes/` contained valuable source maps, quality bars,
content calibration, and process evolution, but also stale paths and retired
roles.

**Change**: The migration rule is extractive: durable rules move into the
owning repo docs, skills, or `.work`; historical logs remain local unless the
user asks for an archive.

**Verification**: Active repo surfaces contain the durable rule, stale old-path
refs are absent from contract surfaces, and local Slock memory is pointer-only.

**Falsifier**: If auditors repeatedly need historical note details to operate
current work, the migration was too thin and should create explicit historical
repo records.

**Regression case**: A future migration copies an entire Slock note with old
repo paths into a leaf doc; the skill should reject that and synthesize a clean
repo-owned rule instead.
