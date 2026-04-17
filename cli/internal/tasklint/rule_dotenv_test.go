package tasklint

import "testing"

func TestRuleDotenvFilesGitignored(t *testing.T) {
	t.Run("pass no dotenv", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: "version: '3'\ntasks: {}\n"}
		assertRuleAbsent(t, fx.run(), "dotenv-files-gitignored")
	})
	t.Run("pass top-level dotenv gitignored", func(t *testing.T) {
		fx := &testFixture{t: t,
			taskfile: `version: '3'
dotenv: ['.env']
tasks: {}
`,
			gitignore: ".env\n",
		}
		assertRuleAbsent(t, fx.run(), "dotenv-files-gitignored")
	})
	t.Run("pass example file skipped", func(t *testing.T) {
		fx := &testFixture{t: t,
			taskfile: `version: '3'
dotenv: ['.env.example']
tasks: {}
`,
		}
		assertRuleAbsent(t, fx.run(), "dotenv-files-gitignored")
	})
	t.Run("pass sample file skipped", func(t *testing.T) {
		fx := &testFixture{t: t,
			taskfile: `version: '3'
dotenv: ['config.sample']
tasks: {}
`,
		}
		assertRuleAbsent(t, fx.run(), "dotenv-files-gitignored")
	})
	t.Run("fail top-level not gitignored", func(t *testing.T) {
		fx := &testFixture{t: t,
			taskfile: `version: '3'
dotenv: ['.env']
tasks: {}
`,
			gitignore: "node_modules/\n",
		}
		got := assertHasRule(t, fx.run(), "dotenv-files-gitignored")
		if got[0].Line == 0 {
			t.Errorf("expected line info, got Line=0")
		}
	})
	t.Run("fail task-level not gitignored", func(t *testing.T) {
		fx := &testFixture{t: t,
			taskfile: `version: '3'
tasks:
  build:
    dotenv: ['.env.prod']
    cmds: [echo hi]
`,
			gitignore: ".env\n",
		}
		assertHasRule(t, fx.run(), "dotenv-files-gitignored")
	})
	t.Run("fail no gitignore present", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: `version: '3'
dotenv: ['.env']
tasks: {}
`}
		assertHasRule(t, fx.run(), "dotenv-files-gitignored")
	})
}
