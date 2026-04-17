package tasklint

import (
	"strings"
	"testing"
)

func TestRuleUnknownTopLevelKeys(t *testing.T) {
	t.Run("pass", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: "version: '3'\ntasks:\n  build:\n    cmds: [echo hi]\n"}
		assertRuleAbsent(t, fx.run(), "unknown-top-level-keys")
	})
	t.Run("typo variables", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: "version: '3'\nvariables:\n  foo: bar\ntasks: {}\n"}
		got := assertHasRule(t, fx.run(), "unknown-top-level-keys")
		if !strings.Contains(got[0].Detail, "vars") {
			t.Errorf("expected typo hint for `vars` in Detail, got %q", got[0].Detail)
		}
	})
	t.Run("nonsense key", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: "version: '3'\nhello: world\ntasks: {}\n"}
		got := assertHasRule(t, fx.run(), "unknown-top-level-keys")
		if got[0].Line < 2 {
			t.Errorf("expected line >= 2, got %d", got[0].Line)
		}
	})
	t.Run("rejects experimental (not in upstream schema)", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: "version: '3'\nexperimental: {}\ntasks: {}\n"}
		got := assertHasRule(t, fx.run(), "unknown-top-level-keys")
		if !strings.Contains(got[0].Message, "experimental") {
			t.Errorf("expected message to flag `experimental`, got %q", got[0].Message)
		}
	})
	t.Run("rejects requires at root (cmd-only field)", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: "version: '3'\nrequires:\n  vars: [X]\ntasks: {}\n"}
		assertHasRule(t, fx.run(), "unknown-top-level-keys")
	})
	t.Run("accepts yaml merge key at root", func(t *testing.T) {
		// Bug 6 regression: `<<:` is a YAML merge directive, not a schema key.
		fx := &testFixture{t: t, taskfile: "x-defs: &base\n  tasks: {}\nversion: '3'\n<<: *base\ntasks: {}\n"}
		findings := fx.run()
		for _, f := range findings {
			if f.RuleID == "unknown-top-level-keys" && strings.Contains(f.Message, "<<") {
				t.Errorf("merge key `<<` flagged as unknown top-level: %+v", f)
			}
		}
	})
}

func TestRuleUnknownTaskKeys(t *testing.T) {
	t.Run("pass", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: `version: '3'
tasks:
  build:
    desc: Build it
    cmds: [echo hi]
    sources: ['**/*.go']
`}
		assertRuleAbsent(t, fx.run(), "unknown-task-keys")
	})
	t.Run("unknown key 'command'", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: `version: '3'
tasks:
  build:
    command: echo hi
`}
		got := assertHasRule(t, fx.run(), "unknown-task-keys")
		if !strings.Contains(got[0].Message, "command") {
			t.Errorf("expected message to mention the bad key, got %q", got[0].Message)
		}
	})
	t.Run("unknown key 'script'", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: `version: '3'
tasks:
  build:
    cmds: [echo hi]
    script: ./x.sh
`}
		got := assertHasRule(t, fx.run(), "unknown-task-keys")
		if got[0].Line == 0 {
			t.Errorf("expected line info, got Line=0")
		}
	})
	t.Run("shortcut task skipped", func(t *testing.T) {
		// Scalar/sequence shortcut tasks have no keys; rule should not trigger.
		fx := &testFixture{t: t, taskfile: `version: '3'
tasks:
  build: echo hi
  test:
    - echo a
    - echo b
`}
		assertRuleAbsent(t, fx.run(), "unknown-task-keys")
	})
	t.Run("accepts dotenv at task level", func(t *testing.T) {
		// Bug 3 regression: `dotenv:` is a valid task-level key.
		fx := &testFixture{t: t, taskfile: `version: '3'
tasks:
  build:
    dotenv: ['.env']
    cmds: [echo hi]
`}
		assertRuleAbsent(t, fx.run(), "unknown-task-keys")
	})
	t.Run("accepts failfast at task level", func(t *testing.T) {
		// Bug 3 regression: `failfast:` is a valid task-level key.
		fx := &testFixture{t: t, taskfile: `version: '3'
tasks:
  build:
    failfast: true
    cmds: [echo hi]
`}
		assertRuleAbsent(t, fx.run(), "unknown-task-keys")
	})
	t.Run("rejects for at task level (cmd-only)", func(t *testing.T) {
		// Bug 3 regression: `for:` belongs in cmd objects, not tasks.
		fx := &testFixture{t: t, taskfile: `version: '3'
tasks:
  build:
    for: [1, 2]
    cmds: [echo hi]
`}
		got := assertHasRule(t, fx.run(), "unknown-task-keys")
		if !strings.Contains(got[0].Message, "for") {
			t.Errorf("expected `for` flagged, got %q", got[0].Message)
		}
	})
	t.Run("rejects defer at task level (cmd-only)", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: `version: '3'
tasks:
  build:
    defer: true
    cmds: [echo hi]
`}
		assertHasRule(t, fx.run(), "unknown-task-keys")
	})
	t.Run("rejects output at task level (top-level-only)", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: `version: '3'
tasks:
  build:
    output: prefixed
    cmds: [echo hi]
`}
		assertHasRule(t, fx.run(), "unknown-task-keys")
	})
	t.Run("accepts yaml merge key at task level", func(t *testing.T) {
		// Bug 6 regression: `<<:` is a YAML merge directive, not a schema key.
		fx := &testFixture{t: t, taskfile: `version: '3'
x-defs: &base
  desc: common
tasks:
  build:
    <<: *base
    cmds: [go build .]
`}
		findings := fx.run()
		for _, f := range findings {
			if f.RuleID == "unknown-task-keys" && strings.Contains(f.Message, "<<") {
				t.Errorf("merge key `<<` flagged as unknown task key: %+v", f)
			}
		}
	})
}
