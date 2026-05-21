# Bootstrap

Use this when adding the protocol to a repo or epic wrapper.

## Sequence

1. Confirm the target shape:
   - single repo or epic wrapper
   - leaf repos and prefixes
   - Slock agents and handles
   - runtime skill roots actually used
2. Add or update the canonical descriptor map.
3. Add leaf `domain.*` claims.
4. Add repo-owned install/check scripts.
5. Add Taskfile commands.
6. Install generated local projections.
7. Run verification.
8. Commit repo-owned changes by ownership boundary.

## Required Epic Commands

Recommended Taskfile surface:

```yaml
tasks:
  bootstrap:
    desc: Recreate repo symlinks and install local Slock projections.
    cmds:
      - bash scripts/bootstrap.sh
      - bash scripts/install-slock.sh

  agents:check:
    desc: Verify canonical Slock-to-repo mapping.
    cmds:
      - bash scripts/check-agent-map.sh

  slock:install:
    desc: Install local Slock MEMORY.md pointers and repo symlinks.
    cmds:
      - bash scripts/install-slock.sh {{.CLI_ARGS}}

  slock:check:
    desc: Verify installed local Slock pointers and repo symlinks.
    cmds:
      - CHECK_LOCAL_SLOCK=1 bash scripts/check-agent-map.sh
```

`task verify` should include descriptor-level registry checks. Local Slock
checks may be opt-in because they depend on machine-local agent dirs.

## Install Script Contract

The target repo owns the script. A correct `install-slock` script:

- reads `slock_agent_registry`
- refuses unsupported `symlink_mode`
- creates or refreshes `<agent-dir>/<symlink_name>` as a symlink to the repo
- refuses to overwrite non-symlink paths without approval
- writes generated pointer-only `MEMORY.md`
- adds the generated symlink to local git exclude when the Slock dir is a git
  repo
- never copies tokens, runtime sessions, channel exports, or notes into the
  repo

## Check Script Contract

A correct `check-agent-map` script verifies:

- required registry fields are present
- `key`, `agent_id`, `repo`, `prefix`, and current handle are unique
- mapped repos exist and are declared leaves
- `prefix` agrees with work-ref prefixes
- leaf `domain.*` agrees with the registry
- local symlink target agrees with the descriptor when local Slock checks are
  enabled
- `MEMORY.md` exactly matches the generated pointer-only form when local checks
  are enabled

## Generated Memory Shape

Use a clear generated marker:

```markdown
# <display-name>

<!-- slock-generated:pointer-only -->

Slock handle: `<current-handle>`
Desired handle: `<desired-handle>` # only when different
Display name: `<display-name>`

Owning repo: `<repo-target>`
Repo symlink: `repo`
Epic protocol: `<epic-target>`
Work prefix: `<prefix>`
Home channel: `<channel>`

Role: <one-line boundary>.

Commands:

- `task verify`
- `task work -- view ready`
- `task work -- show W-NNNN`

Slock is notification only. Canonical work state lives in the repo `.work/`.
The canonical Slock-to-repo mapping lives in `<descriptor>`.
```

If a project needs additional startup guidance, put it in the repo's
`AGENTS.md` or a repo-local skill reference, not in generated memory.

## Commit Boundaries

Commit epic protocol changes in the epic repo. Commit leaf docs, skill roots,
and `.work` type changes in the leaf repo that owns them. Do not commit local
Slock runtime projections as canonical repo data.
