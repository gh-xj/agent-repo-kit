# Open-Source Git Exclude Workflow

Use this workflow when you want convention-engineering in an open-source repo without committing personal workflow files.

This is the variant triggered by prompts like:

- `open source project git exclude`
- `oss git exclude`
- `another repo git exclude setup`

## Goal

Apply local convention scaffolding through `.git/info/exclude` and keep planning docs in `.docs/` (not `docs/`).

## Step 1: Add Local Ignore Rules

Edit `<repo>/.git/info/exclude` and add:

```gitignore
AGENTS.override.md
CLAUDE.local.md
.agent-local/
.agents/skills/_local/
Taskfile.yml
taskfiles/
.worktrees/
.wt-*/
.docs
.docs/
.claude
.claude/
.session
.session/
```

Verify:

```bash
git check-ignore -v .docs CLAUDE.local.md AGENTS.override.md .claude .session Taskfile.yml
```

## Step 2: Create `.docs` Workspace

Create local scratch docs under `.docs/`:

```bash
mkdir -p .docs/{requests,planning,plans,implementation,taxonomy}
```

Optional symlink to a shared workspace:

```bash
ln -sfn "$HOME/<your-state-dir>/<repo-name>/docs" .docs
```

## Step 3: Add Local Convention Files (Untracked)

Create local versions of your convention files (for example `CLAUDE.local.md`, `AGENTS.override.md`, `Taskfile.yml`, `taskfiles/`) and keep them untracked via git exclude.

Do not replace tracked upstream docs unless you intend to contribute those changes upstream.

## Step 4: Add Checker Config for OSS Overlay

Create `.docs/convention-engineering.overlay.json`:

```json
{
  "contract_version": 1,
  "mode": "overlay",
  "profiles": ["go"],
  "docs_root": ".docs",
  "ownership_policy": {
    "portable_skill_authoring_owner": "skill-author",
    "domain_knowledge_owner": "domain-skills",
    "repo_local_skills": {
      "allowed": false,
      "placement_roots": [".claude/skills", ".codex/skills"],
      "authoring_owner": "skill-author",
      "requires_justification": true
    }
  },
  "mirror_policy": {
    "mode": "mirrored",
    "files": ["CLAUDE.local.md", "AGENTS.override.md"]
  },
  "evaluation_inputs": {
    "repo_risk": "standard"
  },
  "chunk_plan": {
    "enabled": false,
    "chunks": []
  },
  "required_files": ["Taskfile.yml"],
  "git_exclude_checks": [
    {
      "name": "oss-local-overlay",
      "file": ".git/info/exclude",
      "required_patterns": [
        ".docs",
        ".docs/",
        "CLAUDE.local.md",
        "AGENTS.override.md",
        "Taskfile.yml"
      ]
    }
  ]
}
```

## Step 5: Run Contract Checker

```bash
SKILL_DIR="$HOME/.claude/skills/convention-engineering"
GO111MODULE=off go run "$SKILL_DIR/scripts" \
  --repo-root . \
  --config .docs/convention-engineering.overlay.json \
  --json
```

For another repo, point `--repo-root` to that path.

## Step 6: Report

Report:

- Which patterns passed/failed in `.git/info/exclude`
- Whether `.docs/` exists and is ignored
- Whether `.docs` taxonomy folders exist (`requests/`, `planning/`, `plans/`)
- Whether local convention files are untracked as intended
