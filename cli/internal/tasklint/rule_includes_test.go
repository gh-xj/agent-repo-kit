package tasklint

import "testing"

func TestRuleIncludesPathsResolvable(t *testing.T) {
	t.Run("pass existing include", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: `version: '3'
includes:
  sub: ./sub/Taskfile.yml
tasks: {}
`, includeFiles: map[string]string{
			"sub/Taskfile.yml": "version: '3'\ntasks: {}\n",
		}}
		assertRuleAbsent(t, fx.run(), "includes-paths-resolvable")
	})
	t.Run("fail missing include", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: `version: '3'
includes:
  sub: ./no-such/Taskfile.yml
tasks: {}
`}
		got := assertHasRule(t, fx.run(), "includes-paths-resolvable")
		if got[0].Line == 0 {
			t.Errorf("expected line info, got Line=0")
		}
	})
	t.Run("pass optional missing include", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: `version: '3'
includes:
  sub:
    taskfile: ./no-such/Taskfile.yml
    optional: true
tasks: {}
`}
		assertRuleAbsent(t, fx.run(), "includes-paths-resolvable")
	})
	t.Run("pass remote scheme skipped", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: `version: '3'
includes:
  remote: https://example.com/Taskfile.yml
tasks: {}
`}
		assertRuleAbsent(t, fx.run(), "includes-paths-resolvable")
	})
	t.Run("pass directory include", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: `version: '3'
includes:
  sub: ./sub
tasks: {}
`, includeFiles: map[string]string{
			"sub/Taskfile.yml": "version: '3'\ntasks: {}\n",
		}}
		assertRuleAbsent(t, fx.run(), "includes-paths-resolvable")
	})
}

func TestRuleFlattenNoNameCollision(t *testing.T) {
	t.Run("pass flatten without collision", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: `version: '3'
includes:
  sub:
    taskfile: ./sub/Taskfile.yml
    flatten: true
tasks:
  only-in-root:
    cmds: [echo root]
`, includeFiles: map[string]string{
			"sub/Taskfile.yml": `version: '3'
tasks:
  only-in-sub:
    cmds: [echo sub]
`,
		}}
		assertRuleAbsent(t, fx.run(), "flatten-no-name-collision")
	})
	t.Run("fail collide with root task", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: `version: '3'
includes:
  sub:
    taskfile: ./sub/Taskfile.yml
    flatten: true
tasks:
  build:
    cmds: [echo root]
`, includeFiles: map[string]string{
			"sub/Taskfile.yml": `version: '3'
tasks:
  build:
    cmds: [echo sub]
`,
		}}
		assertHasRule(t, fx.run(), "flatten-no-name-collision")
	})
	t.Run("fail collide between two flattens", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: `version: '3'
includes:
  a:
    taskfile: ./a/Taskfile.yml
    flatten: true
  b:
    taskfile: ./b/Taskfile.yml
    flatten: true
tasks: {}
`, includeFiles: map[string]string{
			"a/Taskfile.yml": "version: '3'\ntasks:\n  shared:\n    cmds: [echo a]\n",
			"b/Taskfile.yml": "version: '3'\ntasks:\n  shared:\n    cmds: [echo b]\n",
		}}
		assertHasRule(t, fx.run(), "flatten-no-name-collision")
	})
	t.Run("pass collision excluded", func(t *testing.T) {
		fx := &testFixture{t: t, taskfile: `version: '3'
includes:
  sub:
    taskfile: ./sub/Taskfile.yml
    flatten: true
    excludes: [build]
tasks:
  build:
    cmds: [echo root]
`, includeFiles: map[string]string{
			"sub/Taskfile.yml": `version: '3'
tasks:
  build:
    cmds: [echo sub]
  other:
    cmds: [echo other]
`,
		}}
		assertRuleAbsent(t, fx.run(), "flatten-no-name-collision")
	})
}
