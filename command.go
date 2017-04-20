// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package slackbot

import (
	"fmt"
	"os/exec"
)

// Command is the interface the bot uses to run things
type Command interface {
	Run(channel string, args []string) (string, error)
	ShowResult() bool // Whether to output result back to channel
	Description() string
}

// execCommand is a Command that does an exec.Command(...) on the system
type execCommand struct {
	exec        string   // Command to execute
	args        []string // Args for exec.Command
	showResult  bool
	description string
	config      Config
}

// NewExecCommand creates an ExecCommand
func NewExecCommand(exec string, args []string, showResult bool, description string, config Config) Command {
	return execCommand{
		exec:        exec,
		args:        args,
		showResult:  showResult,
		description: description,
		config:      config,
	}
}

// Run runs the exec command
func (c execCommand) Run(_ string, _ []string) (string, error) {
	if c.config.DryRun() {
		return fmt.Sprintf("I'm in dry run mode. I would have run `%s` with args: %s", c.exec, c.args), nil
	}

	out, err := exec.Command(c.exec, c.args...).CombinedOutput()
	outAsString := fmt.Sprintf("%s", out)
	return outAsString, err
}

// ShowResult decides whether to show the results from the exec
func (c execCommand) ShowResult() bool {
	return c.config.DryRun() || c.showResult
}

// Description describes the command
func (c execCommand) Description() string {
	return c.description
}

// CommandFn is the function that is run for this command
type CommandFn func(channel string, args []string) (string, error)

// NewFuncCommand creates a new function command
func NewFuncCommand(fn CommandFn, desc string, config Config) Command {
	return funcCommand{
		fn:     fn,
		desc:   desc,
		config: config,
	}
}

type funcCommand struct {
	desc   string
	fn     CommandFn
	config Config
}

func (c funcCommand) Run(channel string, args []string) (string, error) {
	return c.fn(channel, args)
}

func (c funcCommand) ShowResult() bool {
	return true
}

func (c funcCommand) Description() string {
	return c.desc
}
