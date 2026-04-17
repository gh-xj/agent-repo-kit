package tasklint

import (
	"strings"
	"testing"
)

func TestRuleSchemaError(t *testing.T) {
	t.Run("pass valid taskfile", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: "version: '3'\ntasks:\n  build:\n    cmds: [echo hi]\n"}
		assertRuleAbsent(t, fx.run(), "schema-error")
	})

	t.Run("dotenv scalar at root triggers schema-error", func(t *testing.T) {
		// Bug 2 regression: upstream AST expects `dotenv` as []string.
		// A scalar value fails `yaml.Unmarshal` into `ast.Taskfile`
		// and our linter must surface that as a `schema-error`.
		fx := &testFixture{t: t, taskfile: "version: '3'\ndotenv: \".env\"\ntasks:\n  build:\n    cmds: [echo hi]\n"}
		got := assertHasRule(t, fx.run(), "schema-error")
		if !strings.Contains(got[0].Detail, "dotenv") && !strings.Contains(got[0].Detail, "[]string") {
			t.Errorf("expected detail to mention dotenv or []string, got %q", got[0].Detail)
		}
	})

	t.Run("silent scalar-string at root triggers schema-error", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: "version: '3'\nsilent: \"nope\"\ntasks:\n  build:\n    cmds: [echo hi]\n"}
		assertHasRule(t, fx.run(), "schema-error")
	})

	t.Run("schema-error does not short-circuit other rules", func(t *testing.T) {
		// Invalid `dotenv` AND an unknown top-level key → both rules fire.
		fx := &testFixture{t: t, taskfile: "version: '3'\ndotenv: \".env\"\nvariables:\n  foo: bar\ntasks: {}\n"}
		findings := fx.run()
		schema := findingsByRule(findings, "schema-error")
		unknown := findingsByRule(findings, "unknown-top-level-keys")
		if len(schema) == 0 {
			t.Fatalf("expected schema-error finding, got none; all findings:\n%s", dumpFindings(findings))
		}
		if len(unknown) == 0 {
			t.Fatalf("expected unknown-top-level-keys finding, got none; all findings:\n%s", dumpFindings(findings))
		}
	})
}
