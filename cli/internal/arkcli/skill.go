package arkcli

// SkillCmd groups the `ark skill` subcommands. Each leaf lives in its own
// file (skill_init.go, skill_audit.go, skill_sync.go, skill_check.go).
type SkillCmd struct {
	Init  SkillInitCmd  `cmd:"" help:"scaffold a skill router"`
	Audit SkillAuditCmd `cmd:"" help:"audit a skill router, references, and local CLI layout"`
	Sync  SkillSyncCmd  `cmd:"" help:"render per-adapter SKILL files from canonical skill sources"`
	Check SkillCheckCmd `cmd:"" help:"verify per-adapter SKILL files are in sync with canonical sources"`
}
