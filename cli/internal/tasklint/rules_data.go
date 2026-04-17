package tasklint

// Rule describes a single lint rule for catalog/listing purposes.
type Rule struct {
	ID          string
	Title       string
	Description string
	Severity    Severity
}

// V1 rule catalog — order mirrors ruleFuncs.
var rulesCatalog = []Rule{
	{
		ID:          "version-required",
		Title:       "Taskfile must declare a version",
		Description: "The top-level `version:` key is required so Task can validate the schema.",
		Severity:    SeverityError,
	},
	{
		ID:          "version-is-three",
		Title:       "Taskfile version must be 3 (or 3.x)",
		Description: "Only version 3 of the Taskfile schema is supported.",
		Severity:    SeverityError,
	},
	{
		ID:          "unknown-top-level-keys",
		Title:       "Unknown top-level keys are rejected",
		Description: "Only the documented set of top-level keys is allowed.",
		Severity:    SeverityError,
	},
	{
		ID:          "unknown-task-keys",
		Title:       "Unknown task-level keys are rejected",
		Description: "Only the documented set of task keys is allowed inside a task.",
		Severity:    SeverityError,
	},
	{
		ID:          "cmd-and-cmds-mutex",
		Title:       "A task cannot set both cmd and cmds",
		Description: "Pick one — `cmd:` for a single command, `cmds:` for a list.",
		Severity:    SeverityError,
	},
	{
		ID:          "includes-paths-resolvable",
		Title:       "Non-optional include paths must resolve on disk",
		Description: "Every include entry (unless optional or using a remote scheme) must point at an existing file.",
		Severity:    SeverityError,
	},
	{
		ID:          "flatten-no-name-collision",
		Title:       "Flattened includes must not collide with existing task names",
		Description: "When `flatten: true`, the included file's task names must not already exist at the root or in another flattened include.",
		Severity:    SeverityError,
	},
	{
		ID:          "method-valid-enum",
		Title:       "method: must be checksum, timestamp, or none",
		Description: "Applies to the top-level default and per-task overrides.",
		Severity:    SeverityError,
	},
	{
		ID:          "fingerprint-dir-gitignored",
		Title:       ".task/ must be gitignored when any task uses sources:",
		Description: "Task writes fingerprints under `.task/`. Check it into .gitignore to avoid committing generated artifacts.",
		Severity:    SeverityError,
	},
	{
		ID:          "dotenv-files-gitignored",
		Title:       "dotenv files must be gitignored (unless *.example / *.sample)",
		Description: "Dotenv files typically hold secrets. Commit an example template instead and keep the real file out of git.",
		Severity:    SeverityError,
	},
}

// Rules returns the V1 rule set. Callers may use this for docs-sync
// tests or for surfacing a `--list-rules` output.
func Rules() []Rule {
	out := make([]Rule, len(rulesCatalog))
	copy(out, rulesCatalog)
	return out
}
