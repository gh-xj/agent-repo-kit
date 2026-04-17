package tasklint

import "testing"

func TestRuleVersionRequired(t *testing.T) {
	t.Run("pass", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: "version: '3'\ntasks: {}\n"}
		assertRuleAbsent(t, fx.run(), "version-required")
	})
	t.Run("missing version", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: "tasks:\n  build:\n    cmds: [echo hi]\n"}
		got := assertHasRule(t, fx.run(), "version-required")
		if got[0].Line < 1 {
			t.Fatalf("expected line >= 1, got %d", got[0].Line)
		}
	})
	t.Run("empty document", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: "# just a comment\n"}
		// Empty doc yields a parse-error (root is not a mapping) —
		// which satisfies the contract that broken inputs don't
		// cascade into misleading rule findings.
		findings := fx.run()
		if len(findings) != 1 || findings[0].RuleID != "parse-error" {
			t.Fatalf("expected single parse-error finding, got:\n%s", dumpFindings(findings))
		}
	})
}

func TestRuleVersionIsThree(t *testing.T) {
	t.Run("pass version 3", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: "version: '3'\ntasks: {}\n"}
		assertRuleAbsent(t, fx.run(), "version-is-three")
	})
	t.Run("pass version 3.0", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: "version: '3.0'\ntasks: {}\n"}
		assertRuleAbsent(t, fx.run(), "version-is-three")
	})
	t.Run("fail version 2", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: "version: '2'\ntasks: {}\n"}
		got := assertHasRule(t, fx.run(), "version-is-three")
		if got[0].Line == 0 {
			t.Fatalf("expected location info, got Line=0")
		}
	})
	t.Run("fail version 4", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: "version: '4'\ntasks: {}\n"}
		assertHasRule(t, fx.run(), "version-is-three")
	})
	t.Run("fail numeric without quotes", func(t *testing.T) {
		// `version: 3` (unquoted int) parses to "3" in yaml.Node.Value,
		// so this case actually passes — documented as rule-friendly.
		fx := &testFixture{t: t, taskfile: "version: 3\ntasks: {}\n"}
		assertRuleAbsent(t, fx.run(), "version-is-three")
	})
}
