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
| Slock `repo` symlink | Generated local projection | Convenience entrypoint to owning repo. |
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

## Boundary

Repos may cite lightweight Slock refs such as `#channel:msgid` for provenance.
Repos must not store tokens, runtime sessions, raw channel exports, or copied
local memory as canonical docs.
