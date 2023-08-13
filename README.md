# action

Inspired by [Cobra](https://github.com/spf13/cobra). Cobra is a library for creating powerful modern CLI applications.
"action" can be used to create a more fine-grained behavior of a command.

[![Test Status](https://github.com/shipengqi/action/actions/workflows/test.yml/badge.svg)](https://github.com/shipengqi/action/actions/workflows/test.yml)
[![Codecov](https://codecov.io/gh/shipengqi/action/branch/main/graph/badge.svg?token=SMU4SI304O)](https://codecov.io/gh/shipengqi/action)
[![Release](https://img.shields.io/github/release/shipengqi/action.svg)](https://github.com/shipengqi/action/releases)
[![License](https://img.shields.io/github/license/shipengqi/action)](https://github.com/shipengqi/action/blob/main/LICENSE)

## Quickstart

```go
cmd := &cobra.Command{
    Use:   "example-cmd",
    Short: "A sample command.",
    RunE: func(cmd *cobra.Command, args []string) error {
        act := &action.Action{
            Name: "example-action",
            Run:  func(act *action.Action) error { return nil },
        }
        
        _ = act.AddAction(
            newSubAction1(),
            newSubAction2(),
        )

        act.Execute()
    },
}

func newSubAction1() *action.Action {
    return &action.Action{
        Name: "sub-action1",
		Executable: func(act *action.Action) bool {
			// do something
			return true
        },
        Run:  func(act *action.Action) error { return nil },
    }
}

func newSubAction2() *action.Action {
    return &action.Action{
        Name: "sub-action2",
        Executable: func(act *action.Action) bool {
            // do something
            return false
        },
        Run:  func(act *action.Action) error { return nil },
    }
}
```

- `Executable`: whether is an executable action.
- `Execute()` will find the first executable action of the root action and execute it.

## Documentation

You can find the docs at [go docs](https://pkg.go.dev/github.com/shipengqi/action).
