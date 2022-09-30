package action

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const _rootActionName = "root"

func getTestHookFunc(t *testing.T, name string) func(act *Action) error {
	return func(act *Action) error {
		t.Logf("%s action name: %s", name, act.Name)
		return nil
	}
}

func getTargetFunc(istarget bool) func(act *Action) bool {
	return func(act *Action) bool {
		return istarget
	}
}

func getRootAction(t *testing.T) *Action {
	return &Action{
		Name:              "root",
		PersistentPreRun:  getTestHookFunc(t, "PersistentPreRun"),
		PreRun:            getTestHookFunc(t, "PreRun"),
		Run:               getTestHookFunc(t, "Run"),
		PostRun:           getTestHookFunc(t, "PostRun"),
		PersistentPostRun: getTestHookFunc(t, "PersistentPostRun"),
		Executable:        getTargetFunc(false),
	}
}

func getSubActions(t *testing.T, parent string, count int) []*Action {
	if count < 1 {
		return nil
	}
	var actions []*Action
	for i := 0; i < count; i++ {
		actions = append(actions, &Action{
			Name:              fmt.Sprintf("%s-sub-action-%d", parent, i+1),
			PersistentPreRun:  getTestHookFunc(t, "PersistentPreRun"),
			PreRun:            getTestHookFunc(t, "PreRun"),
			Run:               getTestHookFunc(t, "Run"),
			PostRun:           getTestHookFunc(t, "PostRun"),
			PersistentPostRun: getTestHookFunc(t, "PersistentPostRun"),
			Executable:        getTargetFunc(false),
		})
	}
	return actions
}

func TestActionRunnable(t *testing.T) {
	act := getRootAction(t)
	assert.True(t, act.Runnable())
	act.Run = nil
	assert.False(t, act.Runnable())
}

func TestActionAddAction(t *testing.T) {
	t.Run("successfully add 5 actions", func(t *testing.T) {
		act := getRootAction(t)
		acts := getSubActions(t, _rootActionName, 5)
		_ = act.AddAction(acts...)
		assert.Equal(t, 5, len(act.Actions()))
	})
	t.Run("doesn't accept self as sub action", func(t *testing.T) {
		act := getRootAction(t)
		err := act.AddAction(act)
		assert.EqualError(t, err, "action can't be a child of itself")
	})
}

func TestActionParent(t *testing.T) {
	t.Run("should returns parent action", func(t *testing.T) {
		act := getRootAction(t)
		acts := getSubActions(t, _rootActionName, 1)
		_ = act.AddAction(acts...)
		assert.False(t, act.HasParent())
		assert.True(t, act.HasSubActions())
		assert.True(t, acts[0].HasParent())
		assert.Equal(t, act, acts[0].Parent())
	})
}

func TestActionRemove(t *testing.T) {
	act := getRootAction(t)
	acts := getSubActions(t, _rootActionName, 10)
	_ = act.AddAction(acts...)
	acts[9].Executable = getTargetFunc(true)
	assert.Equal(t, acts[9], act.Find())
	act.RemoveAction(acts[9])
	assert.Nil(t, act.Find())
}

func TestActionRoot(t *testing.T) {
	act := getRootAction(t)
	acts := getSubActions(t, _rootActionName, 10)
	_ = act.AddAction(acts...)
	assert.Equal(t, act, act)
	assert.Equal(t, act, acts[9].Root())
}

func TestActionExecuteContext(t *testing.T) {
	act := getRootAction(t)
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
		act := getRootAction(t)
		assert.Nil(t, act.Execute())
	})

	t.Run("returns error in PreRun", func(t *testing.T) {
		act := getRootAction(t)
		act.PreRun = func(act *Action) error {
			return errors.New("PreRun error")
		}
		err := act.Execute()
		assert.EqualError(t, err, "PreRun error")
	})

	t.Run("returns error in PostRun", func(t *testing.T) {
		act := getRootAction(t)
		act.PostRun = func(act *Action) error {
			return errors.New("PostRun error")
		}
		err := act.Execute()
		assert.EqualError(t, err, "PostRun error")
	})

	t.Run("returns error in PersistentPreRun", func(t *testing.T) {
		act := getRootAction(t)
		act.PersistentPreRun = func(act *Action) error {
			return errors.New("PersistentPreRun error")
		}
		err := act.Execute()
		assert.EqualError(t, err, "PersistentPreRun error")
	})

	t.Run("returns error in PersistentPostRun", func(t *testing.T) {
		act := getRootAction(t)
		act.PersistentPostRun = func(act *Action) error {
			return errors.New("PersistentPostRun error")
		}
		err := act.Execute()
		assert.EqualError(t, err, "PersistentPostRun error")
	})
}
