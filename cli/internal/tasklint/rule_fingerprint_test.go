package tasklint

import "testing"

func TestRuleFingerprintDirGitignored(t *testing.T) {
	t.Run("pass no sources", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: `version: '3'
tasks:
  build:
    cmds: [echo hi]
`}
		assertRuleAbsent(t, fx.run(), "fingerprint-dir-gitignored")
	})
	t.Run("pass sources plus .task/ in gitignore", func(t *testing.T) {
		fx := &testFixture{t: t,
			taskfile: `version: '3'
tasks:
  build:
    sources: ['**/*.go']
    cmds: [echo hi]
`,
			gitignore: ".task/\n",
		}
		assertRuleAbsent(t, fx.run(), "fingerprint-dir-gitignored")
	})
	t.Run("pass sources plus .task in gitignore without slash", func(t *testing.T) {
		fx := &testFixture{t: t,
			taskfile: `version: '3'
tasks:
  build:
    sources: ['**/*.go']
    cmds: [echo hi]
`,
			gitignore: ".task\n",
		}
		assertRuleAbsent(t, fx.run(), "fingerprint-dir-gitignored")
	})
	t.Run("fail sources without gitignore", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: `version: '3'
tasks:
  build:
    sources: ['**/*.go']
    cmds: [echo hi]
`}
		got := assertHasRule(t, fx.run(), "fingerprint-dir-gitignored")
		if got[0].Line == 0 {
			t.Errorf("expected line info, got Line=0")
		}
	})
	t.Run("fail sources with unrelated gitignore", func(t *testing.T) {
		fx := &testFixture{t: t,
			taskfile: `version: '3'
tasks:
  build:
    sources: ['**/*.go']
    cmds: [echo hi]
`,
			gitignore: "node_modules/\n",
		}
		assertHasRule(t, fx.run(), "fingerprint-dir-gitignored")
	})
}
