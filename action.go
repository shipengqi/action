package action

import (
	"context"
	"errors"
)

// Action is just that, an action for your application.
type Action struct {
	// Name action's name
	Name string

	// The *Run functions are executed in the following order:
	//   * PersistentPreRun()
	//   * PreRun()
	//   * Run()
	//   * PostRun()
	//   * PersistentPostRun()
	// All functions get the same args.
	//
	// PersistentPreRun: children of this action will inherit and execute.
	PersistentPreRun func(act *Action) error
	// PreRun: children of this action will not inherit.
	PreRun func(act *Action) error
	// Run: Typically the actual work function. Most actions will only implement this.
	Run func(act *Action) error
	// PostRun: run after the Run action.
	PostRun func(act *Action) error
	// PersistentPostRun: children of this action will inherit and execute after PostRun.
	PersistentPostRun func(act *Action) error

	// Executable: whether is an executable action.
	Executable func(act *Action) bool

	// actions is the list of actions supported by this action.
	actions []*Action
	// parent is a parent action for this action.
	parent *Action

	ctx context.Context
}

// Context returns underlying action context. If action wasn't
// executed with ExecuteContext Context returns Background context.
func (a *Action) Context() context.Context {
	return a.ctx
}

// HasParent determines if the action is a child action.
func (a *Action) HasParent() bool {
	return a.parent != nil
}

// Runnable determines if the action is itself runnable.
func (a *Action) Runnable() bool {
	return a.Run != nil
}

// Parent returns a actions parent action.
func (a *Action) Parent() *Action {
	return a.parent
}

// Root finds root Action.
func (a *Action) Root() *Action {
	if a.HasParent() {
		return a.Parent().Root()
	}
	return a
}

// HasSubActions determines if the Action has children actions.
func (a *Action) HasSubActions() bool {
	return len(a.actions) > 0
}

// Actions returns a slice of child actions.
func (a *Action) Actions() []*Action {
	return a.actions
}

// AddAction adds one or more actions to this parent action.
func (a *Action) AddAction(actions ...*Action) error {
	for i, x := range actions {
		if actions[i] == a {
			return errors.New("action can't be a child of itself")
		}
		actions[i].parent = a

		a.actions = append(a.actions, x)
	}
	return nil
}

// RemoveAction removes one or more actions from a parent action.
func (a *Action) RemoveAction(actions ...*Action) {
	var acts []*Action
main:
	for _, action := range a.actions {
		for _, act := range actions {
			if action == act {
				action.parent = nil
				continue main
			}
		}
		acts = append(acts, action)
	}
	a.actions = acts
}

// ExecuteContext is the same as Execute(), but sets the ctx on the action.
func (a *Action) ExecuteContext(ctx context.Context) (err error) {
	a.ctx = ctx
	return a.Execute()
}

// Execute executes the action.
func (a *Action) Execute() (err error) {
	if a.ctx == nil {
		a.ctx = context.Background()
	}

	// Regardless of what action execute is called on, run on Root only
	if a.HasParent() {
		return a.Root().Execute()
	}

	var target *Action
	act := a.Find()

	if act != nil {
		target = act
	} else {
		target = a
	}

	// We have to pass global context to children action
	// if context is present on the parent action.
	if target.ctx == nil {
		target.ctx = a.ctx
	}

	err = target.execute()
	if err != nil {
		return err
	}
	return err
}

// Find the first executable action.
func (a *Action) Find() *Action {
	if a.Executable != nil && a.Executable(a) {
		return a
	}
	if !a.HasSubActions() {
		return nil
	}
	for _, v := range a.Actions() {
		if target := v.Find(); target != nil {
			return target
		}
	}
	return nil
}

func (a *Action) execute() error {
	if a == nil {
		return nil
	}

	if !a.Runnable() {
		return nil
	}

	for p := a; p != nil; p = p.Parent() {
		if p.PersistentPreRun != nil {
			if err := p.PersistentPreRun(a); err != nil {
				return err
			}
			break
		}
	}
	if a.PreRun != nil {
		if err := a.PreRun(a); err != nil {
			return err
		}
	}

	if a.Run != nil {
		if err := a.Run(a); err != nil {
			return err
		}
	}
	if a.PostRun != nil {
		if err := a.PostRun(a); err != nil {
			return err
		}
	}
	for p := a; p != nil; p = p.Parent() {
		if p.PersistentPostRun != nil {
			if err := p.PersistentPostRun(a); err != nil {
				return err
			}
			break
		}
	}

	return nil
}
