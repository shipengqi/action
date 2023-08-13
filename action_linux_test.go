package action

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActionFind(t *testing.T) {
	var err error
	t.Run("cannot find target without sub actions", func(t *testing.T) {
		var buf bytes.Buffer
		act := getRootAction(&buf)
		err = act.Execute()
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "PersistentPreRun action name: root")
		assert.Contains(t, buf.String(), "PreRun action name: root")
		assert.Contains(t, buf.String(), "Run action name: root")
		assert.Contains(t, buf.String(), "PostRun action name: root")
		assert.Contains(t, buf.String(), "PersistentPostRun action name: root")
	})

	t.Run("execute the target action", func(t *testing.T) {
		var buf bytes.Buffer
		act := getRootAction(&buf)
		acts := getSubActions(&buf, _rootActionName, 10)
		acts[9].Executable = getTargetFunc(true)
		_ = act.AddAction(acts...)
		err = act.Execute()
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "PersistentPreRun action name: root-sub-action-10")
		assert.Contains(t, buf.String(), "PreRun action name: root-sub-action-10")
		assert.Contains(t, buf.String(), "Run action name: root-sub-action-10")
		assert.Contains(t, buf.String(), "PostRun action name: root-sub-action-10")
		assert.Contains(t, buf.String(), "PersistentPostRun action name: root-sub-action-10")
	})

	t.Run("execute the multi layers target action", func(t *testing.T) {
		var buf bytes.Buffer
		act := getRootAction(&buf)
		acts := getSubActions(&buf, _rootActionName, 5)
		subsubs := getSubActions(&buf, "root-sub-action-5", 10)
		subsubs[9].Executable = getTargetFunc(true)
		_ = acts[4].AddAction(subsubs...)
		_ = act.AddAction(acts...)
		err = act.Execute()
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "PersistentPreRun action name: root-sub-action-5-sub-action-10")
		assert.Contains(t, buf.String(), "PersistentPreRun action name: root-sub-action-5-sub-action-10")
		assert.Contains(t, buf.String(), "PreRun action name: root-sub-action-5-sub-action-10")
		assert.Contains(t, buf.String(), "Run action name: root-sub-action-5-sub-action-10")
		assert.Contains(t, buf.String(), "PostRun action name: root-sub-action-5-sub-action-10")
		assert.Contains(t, buf.String(), "PersistentPostRun action name: root-sub-action-5-sub-action-10")
	})
}
