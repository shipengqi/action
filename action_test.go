package action

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const _rootActionName = "root"

func getTestHookFunc(buf *bytes.Buffer, name string) func(act *Action) error {
	return func(act *Action) error {
		_, _ = fmt.Fprintf(buf, "%s action name: %s\n", name, act.Name)
		return nil
	}
}

func getTargetFunc(istarget bool) func(act *Action) bool {
	return func(act *Action) bool {
		return istarget
	}
}

func getRootAction(buf *bytes.Buffer) *Action {
	return &Action{
		Name:              "root",
		PersistentPreRun:  getTestHookFunc(buf, "PersistentPreRun"),
		PreRun:            getTestHookFunc(buf, "PreRun"),
		Run:               getTestHookFunc(buf, "Run"),
		PostRun:           getTestHookFunc(buf, "PostRun"),
		PersistentPostRun: getTestHookFunc(buf, "PersistentPostRun"),
		Executable:        getTargetFunc(false),
	}
}

func getSubActions(buf *bytes.Buffer, parent string, count int) []*Action {
	if count < 1 {
		return nil
	}
	var actions []*Action
	for i := 0; i < count; i++ {
		actions = append(actions, &Action{
			Name:              fmt.Sprintf("%s-sub-action-%d", parent, i+1),
			PersistentPreRun:  getTestHookFunc(buf, "PersistentPreRun"),
			PreRun:            getTestHookFunc(buf, "PreRun"),
			Run:               getTestHookFunc(buf, "Run"),
			PostRun:           getTestHookFunc(buf, "PostRun"),
			PersistentPostRun: getTestHookFunc(buf, "PersistentPostRun"),
			Executable:        getTargetFunc(false),
		})
	}
	return actions
}

func TestActionRunnable(t *testing.T) {
	var buf bytes.Buffer
	act := getRootAction(&buf)
	assert.True(t, act.Runnable())
	act.Run = nil
	assert.False(t, act.Runnable())
}

func TestActionAddAction(t *testing.T) {
	t.Run("successfully add 5 actions", func(t *testing.T) {
		var buf bytes.Buffer
		act := getRootAction(&buf)
		acts := getSubActions(&buf, _rootActionName, 5)
		_ = act.AddAction(acts...)
		assert.Equal(t, 5, len(act.Actions()))
	})
	t.Run("doesn't accept self as sub action", func(t *testing.T) {
		var buf bytes.Buffer
		act := getRootAction(&buf)
		err := act.AddAction(act)
		assert.EqualError(t, err, "action can't be a child of itself")
	})
}

func TestActionParent(t *testing.T) {
	t.Run("should returns parent action", func(t *testing.T) {
		var buf bytes.Buffer
		act := getRootAction(&buf)
		acts := getSubActions(&buf, _rootActionName, 1)
		_ = act.AddAction(acts...)
		assert.False(t, act.HasParent())
		assert.True(t, act.HasSubActions())
		assert.True(t, acts[0].HasParent())
		assert.Equal(t, act, acts[0].Parent())
	})
}

func TestActionRemove(t *testing.T) {
	var buf bytes.Buffer
	act := getRootAction(&buf)
	acts := getSubActions(&buf, _rootActionName, 10)
	_ = act.AddAction(acts...)
	acts[9].Executable = getTargetFunc(true)
	assert.Equal(t, acts[9], act.Find())
	act.RemoveAction(acts[9])
	assert.Nil(t, act.Find())
}

func TestActionRoot(t *testing.T) {
	var buf bytes.Buffer
	act := getRootAction(&buf)
	acts := getSubActions(&buf, _rootActionName, 10)
	_ = act.AddAction(acts...)
	assert.Equal(t, act, act)
	assert.Equal(t, act, acts[9].Root())
}

func TestActionExecuteContext(t *testing.T) {
	var buf bytes.Buffer
	act := getRootAction(&buf)
	act.Run = func(act *Action) error {
		for {
			select {
			case <-act.Context().Done():
				return errors.New("done")
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(500 * time.Millisecond)
		cancel()
	}()
	err := act.ExecuteContext(ctx)
	assert.EqualError(t, err, "done")
}

func TestActionExecute(t *testing.T) {
	t.Run("ignore action without Run", func(t *testing.T) {
		var buf bytes.Buffer
		act := getRootAction(&buf)
		assert.Nil(t, act.Execute())
	})

	t.Run("returns error in PreRun", func(t *testing.T) {
		var buf bytes.Buffer
		act := getRootAction(&buf)
		act.PreRun = func(act *Action) error {
			return errors.New("PreRun error")
		}
		err := act.Execute()
		assert.EqualError(t, err, "PreRun error")
	})

	t.Run("returns error in PostRun", func(t *testing.T) {
		var buf bytes.Buffer
		act := getRootAction(&buf)
		act.PostRun = func(act *Action) error {
			return errors.New("PostRun error")
		}
		err := act.Execute()
		assert.EqualError(t, err, "PostRun error")
	})

	t.Run("returns error in PersistentPreRun", func(t *testing.T) {
		var buf bytes.Buffer
		act := getRootAction(&buf)
		act.PersistentPreRun = func(act *Action) error {
			return errors.New("PersistentPreRun error")
		}
		err := act.Execute()
		assert.EqualError(t, err, "PersistentPreRun error")
	})

	t.Run("returns error in PersistentPostRun", func(t *testing.T) {
		var buf bytes.Buffer
		act := getRootAction(&buf)
		act.PersistentPostRun = func(act *Action) error {
			return errors.New("PersistentPostRun error")
		}
		err := act.Execute()
		assert.EqualError(t, err, "PersistentPostRun error")
	})
}
