package action

import "github.com/Netflix/go-expect"

type expectConsole interface {
	ExpectString(string)
	ExpectEOF()
	SendLine(string)
	Send(string)
}

type consoleWithErrorHandling struct {
	console *expect.Console
	t       *testing.T
}

func (c *consoleWithErrorHandling) ExpectString(s string) {
	if _, err := c.console.ExpectString(s); err != nil {
		c.t.Helper()
		c.t.Fatalf("ExpectString(%q) = %v", s, err)
	}
}

func (c *consoleWithErrorHandling) SendLine(s string) {
	if _, err := c.console.SendLine(s); err != nil {
		c.t.Helper()
		c.t.Fatalf("SendLine(%q) = %v", s, err)
	}
}

func (c *consoleWithErrorHandling) Send(s string) {
	if _, err := c.console.Send(s); err != nil {
		c.t.Helper()
		c.t.Fatalf("Send(%q) = %v", s, err)
	}
}

func (c *consoleWithErrorHandling) ExpectEOF() {
	if _, err := c.console.ExpectEOF(); err != nil {
		c.t.Helper()
		c.t.Fatalf("ExpectEOF() = %v", err)
	}
}

func TestActionFind(t *testing.T) {
	c, err := expect.NewConsole(expect.WithStdin(os.Stdin), expect.WithStdout(os.Stdout))
	if err != nil {
		t.Fatalf("failed to create console: %v", err)
	}
	defer func() { _ = c.Close() }()
	console := &consoleWithErrorHandling{console: c, t: t}

	t.Run("cannot find target without sub actions", func(t *testing.T) {
		act := getRootAction(t)
		err := act.Execute()
		assert.NoError(t, err)
		console.ExpectString("PersistentPreRun action name: root")
		console.ExpectString("PreRun action name: root")
		console.ExpectString("Run action name: root")
		console.ExpectString("PostRun action name: root")
		console.ExpectString("PersistentPostRun action name: root")
		console.ExpectEOF()
	})

	t.Run("execute the target action", func(t *testing.T) {
		act := getRootAction(t)
		acts := getSubActions(t, _rootActionName, 10)
		acts[9].Executable = getTargetFunc(true)
		_ = act.AddAction(acts...)
		err := act.Execute()
		assert.NoError(t, err)
		console.ExpectString("PersistentPreRun action name: root-sub-action-10")
		console.ExpectString("PreRun action name: root-sub-action-10")
		console.ExpectString("Run action name: root-sub-action-10")
		console.ExpectString("PostRun action name: root-sub-action-10")
		console.ExpectString("PersistentPostRun action name: root-sub-action-10")
		console.ExpectEOF()
	})

	t.Run("execute the multi layers target action", func(t *testing.T) {
		act := getRootAction(t)
		acts := getSubActions(t, _rootActionName, 10)
		subsubs := getSubActions(t, "root-sub-action-10", 10)
		subsubs[9].Executable = getTargetFunc(true)
		_ = acts[9].AddAction(subsubs...)
		_ = act.AddAction(acts...)
		err := act.Execute()
		assert.NoError(t, err)
		console.ExpectString("PersistentPreRun action name: root-sub-action-10")
		console.ExpectString("PreRun action name: root-sub-action-10")
		console.ExpectString("Run action name: root-sub-action-10")
		console.ExpectString("PostRun action name: root-sub-action-10")
		console.ExpectString("PersistentPostRun action name: root-sub-action-10")
		console.ExpectEOF()
	})
}
