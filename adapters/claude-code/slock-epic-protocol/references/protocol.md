# Protocol

This protocol connects Slock agents to repo-owned work without making Slock a
second database.

## Surface Ownership

| Surface | Owner | Rule |
| --- | --- | --- |
| Epic descriptor | Repo | Canonical Slock-to-repo map for multi-repo products. |
| Leaf descriptor | Leaf repo | Local domain claim: prefix, owner handle, home channel, epic pointer. |
| Leaf `.work/` | Leaf repo | Canonical work lifecycle and workspaces. Tracked by default. |
| Leaf docs | Leaf repo | Durable rules, quality bars, operating docs, strategy records. |
| Leaf skills | Leaf repo | Domain procedure and judgment for the agent's work. |
| Slock `MEMORY.md` | Generated local projection | Pointer-only startup router. |
| Slock `repo` symlink | Generated local projection | Convenience entrypoint to owning leaf. |
| Slock `epic` symlink | Generated local projection (optional) | Entry into the epic wrapper for cross-leaf docs/work/skills. Off unless `epic_symlink_name` is set on the registry. |
| Slock `.claude/skills` symlink | Generated local projection (when epic enabled) | Feeds Claude Code's `$CWD/.claude/skills/` auto-scan from the agent's home dir. Points at epic's project skills dir. |
| Slock `.agents/skills` symlink | Generated local projection (when epic enabled) | Same for Codex agents (Codex scans `$CWD/.agents/skills/`). Points at epic's `.agents/skills` (itself a mirror of `.claude/skills`). |
| Slock notes | Local runtime history | Historical scratch unless migrated into repo-owned surfaces. |

## Descriptor Model

For an epic product, the canonical map lives in the epic descriptor:

```yaml
slock_agent_registry:
  canonical_surface: epic_descriptor
  relation_mode: descriptor_only
  memory_mode: pointer_only
  symlink_mode: generated_local
  symlink_name: repo
  epic_symlink_name: epic   # optional; default omitted = off. Set when the
                            # wrapper hosts cross-leaf skills agents need to
                            # discover (e.g. slock-change-protocol).
  agents:
    - key: scout
      agent_id: <uuid>
      repo: <leaf-repo>
      prefix: <work-prefix>
      mention_handle: "@current-handle"
      desired_handle: "@desired-handle"
      display_name: <display-name>
      home_channel: "#channel"
      role: <one-line boundary>
```

When `epic_symlink_name` is set (default `epic` when present), the install
step creates **three** symlinks per agent home:

1. `<agent_home>/<epic_symlink_name>` → wrapper repo root. Gives agents
   path access to epic docs (`epic/.docs/`), epic work (`epic/.work/`), and
   epic skills (`epic/.claude/skills/`).
2. `<agent_home>/.claude/skills` → `<epic>/.claude/skills`. Feeds Claude
   Code's `$CWD/.claude/skills/` auto-scan. **This is what makes epic
   project skills DISCOVERABLE by Claude Code agents** — the `epic`
   symlink alone gives path access but not skill discovery.
3. `<agent_home>/.agents/skills` → `<epic>/.agents/skills`. Same for
   Codex agents.

The two `skills` symlinks are non-optional companions to `epic_symlink_name`
because Claude Code / Codex don't follow ad-hoc symlinks during their
default skill scan; only their canonical `<runtime>/skills/` paths are
auto-discovered. The `check` step verifies all three symlinks.

Leaf-direct operations remain unaffected — these projections only matter
when the agent is invoked with cwd = its Slock home dir.

For a single repo, use `canonical_surface: repo_descriptor` and keep the same
agent fields.

Leaf descriptors should claim only local ownership:

```yaml
domain:
  prefix: <work-prefix>
  owner_handle: "@desired-handle"
  home_channel: "#channel"
  epic: ../<epic-repo>
```

Do not duplicate the whole registry in every leaf.

## Work Model

Default to `per_leaf_tracked`.

- Each leaf repo owns its own `.work/` and workspaces.
- Work IDs are repo-local.
- Across repos, all refs are qualified.
- Handoffs create target repo inbox/work state. Work items are not moved
  between repos.

Example:

```text
scout#IN-0016 -> analysis#IN-0039 -> analysis#W-0039 -> content#IN-0021
```

The target inbox/work item is canonical because it gives the receiving repo a
local lifecycle. The Slock message is only the notification.

## Slock Memory

Generated `MEMORY.md` should contain only:

- display name and current/desired handle
- owning repo path or repo-relative target
- repo symlink name
- epic protocol pointer when applicable
- work prefix
- home channel
- one-line role boundary
- minimal repo commands
- pointer-only boundary statement

No daily logs, quality bars, pipeline rules, source lists, or strategy notes.

## Skill Discovery

Slock does not own durable skills. Slock memory points to the repo; the runtime
discovers skills from repo-local or global skill roots.

Use repo-local roots when the agent's domain procedure is versioned with the
repo:

```text
.claude/skills/<skill>/
.agents/skills/<skill>/
```

If both Claude and Codex may operate the Slock agent, expose the same portable
skill intent through both roots, preferably with a symlink or generated mirror.

### Session restart for new skills

Slock-mediated agents run as long-lived processes (daemon-spawned Claude
Code or Codex sessions). Skill discovery happens **once at process bootstrap**
— after that, the agent's skill set is cached for the lifetime of the
session.

Consequences:

- Adding a new skill (e.g. via `epic_symlink_name` + `.claude/skills`
  projection install) does NOT take effect for an already-running agent.
  The agent continues to operate with its bootstrap-time skill list.
- A new wake-up triggered by an inbound Slock message starts a fresh
  session, which re-scans skills. After that the new skill becomes
  available to that agent.

When you've just installed new skills or refreshed the projection symlinks,
either:

1. Send a benign Slock message to the affected agent(s) to force a
   wake-up + fresh session bootstrap, or
2. Wait for the next natural inbound that wakes them.

Do not assume long-running sessions pick up new skills mid-conversation.
This was empirically confirmed in the 2026-05-20 PVR work: a stale-session
qa-bot continued operating on the prior skill set for 5+ hours of active
chat in a single thread, while a freshly-bootstrapped session in a
parallel thread immediately discovered the new project-level skill.

## Boundary

Repos may cite lightweight Slock refs such as `#channel:msgid` for provenance.
Repos must not store tokens, runtime sessions, raw channel exports, or copied
local memory as canonical docs.
