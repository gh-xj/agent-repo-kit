package tasklint

import "testing"

func TestRuleCmdAndCmdsMutex(t *testing.T) {
	t.Run("pass cmds only", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: `version: '3'
tasks:
  build:
    cmds: [echo hi]
`}
		assertRuleAbsent(t, fx.run(), "cmd-and-cmds-mutex")
	})
	t.Run("pass cmd only", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: `version: '3'
tasks:
  build:
    cmd: echo hi
`}
		assertRuleAbsent(t, fx.run(), "cmd-and-cmds-mutex")
	})
	t.Run("fail both", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: `version: '3'
tasks:
  build:
    cmd: echo a
    cmds: [echo b]
`}
		got := assertHasRule(t, fx.run(), "cmd-and-cmds-mutex")
		if got[0].Line == 0 {
			t.Errorf("expected location info, got Line=0")
		}
	})
	t.Run("fail multiple tasks", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: `version: '3'
tasks:
  build:
    cmd: echo a
    cmds: [echo b]
  test:
    cmds: [echo c]
    cmd: echo d
`}
		got := assertHasRule(t, fx.run(), "cmd-and-cmds-mutex")
		if len(got) != 2 {
			t.Fatalf("expected 2 findings, got %d", len(got))
		}
	})
}
