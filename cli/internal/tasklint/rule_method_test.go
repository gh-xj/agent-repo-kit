package tasklint

import "testing"

func TestRuleMethodValidEnum(t *testing.T) {
	t.Run("pass no method set", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: "version: '3'\ntasks:\n  build:\n    cmds: [echo hi]\n"}
		assertRuleAbsent(t, fx.run(), "method-valid-enum")
	})
	t.Run("pass valid top-level", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: `version: '3'
method: checksum
tasks:
  build:
    cmds: [echo hi]
`}
		assertRuleAbsent(t, fx.run(), "method-valid-enum")
	})
	t.Run("pass valid per-task", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: `version: '3'
tasks:
  build:
    method: timestamp
    cmds: [echo hi]
`}
		assertRuleAbsent(t, fx.run(), "method-valid-enum")
	})
	t.Run("fail bogus top-level", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: `version: '3'
method: banana
tasks:
  build:
    cmds: [echo hi]
`}
		got := assertHasRule(t, fx.run(), "method-valid-enum")
		if got[0].Line == 0 {
			t.Errorf("expected line info, got Line=0")
		}
	})
	t.Run("fail bogus per-task", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: `version: '3'
tasks:
  build:
    method: apple
    cmds: [echo hi]
`}
		assertHasRule(t, fx.run(), "method-valid-enum")
	})
}
