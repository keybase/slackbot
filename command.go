// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package slackbot

import (
	"fmt"
	"os/exec"
)

// Command is the interface the bot uses to run things
type Command interface {
	Run([]string) (string, error)
	ShowResult() bool // Whether to output result back to channel
	Description() string
}

// ExecCommand is a Command that does an exec.Command(...) on the system
type ExecCommand struct {
	exec        string   // Command to execute
	args        []string // Args for exec.Command
	showResult  bool
	description string
}

// ToggleDryRunCommand is a special command that toggles dry run state
type ToggleDryRunCommand struct{}

// FuncCommand runs an arbitrary function on trigger
type FuncCommand struct {
	Desc string
	Fn   func(args []string) (string, error)
}

// NewExecCommand creates an ExecCommand
func NewExecCommand(exec string, args []string, showResult bool, description string) ExecCommand {
	return ExecCommand{
		exec:        exec,
		args:        args,
		showResult:  showResult,
		description: description,
	}
}

// Run runs the exec command
func (c ExecCommand) Run(_ []string) (string, error) {
	config := readConfigOrDefault()

	if config.DryRun {
		return fmt.Sprintf("Dry Run: Doing that would run `%s` with args: %s", c.exec, c.args), nil
	}

	if config.Paused {
		return fmt.Sprintf("I'm paused so I can't do that, but I would have ran `%s` with args: %s", c.exec, c.args), nil
	}

	out, err := exec.Command(c.exec, c.args...).CombinedOutput()
	outAsString := fmt.Sprintf("%s", out)
	return outAsString, err
}

// ShowResult decides whether to show the results from the exec
func (c ExecCommand) ShowResult() bool {
	config := readConfigOrDefault()
	return config.DryRun || config.Paused || c.showResult
}

// Description describes the command
func (c ExecCommand) Description() string {
	return c.description
}

// Run toggles the dry run state. (Itself is never run under dry run mode)
func (c ToggleDryRunCommand) Run(_ []string) (string, error) {
	config, err := updateConfig(func(c Config) (Config, error) {
		c.DryRun = !c.DryRun
		return c, nil
	})

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Dry Run Value is now: %t", config.DryRun), nil
}

// ShowResult always shows results for toggling dry run
func (c ToggleDryRunCommand) ShowResult() bool {
	return true
}

// Description describes what it does
func (c ToggleDryRunCommand) Description() string {
	return "Toggles the Dry Run value"
}

// Run runs the Fn func
func (c FuncCommand) Run(args []string) (string, error) {
	return c.Fn(args)
}

// ShowResult always shows results
func (c FuncCommand) ShowResult() bool {
	return true
}

// Description describes what it does
func (c FuncCommand) Description() string {
	return c.Desc
}
