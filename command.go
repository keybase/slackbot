// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package slackbot

import (
	"fmt"
	"os/exec"
)

type Command interface {
	Run() (string, error)
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

func NewExecCommand(exec string, args []string, showResult bool, description string) ExecCommand {
	return ExecCommand{
		exec:        exec,
		args:        args,
		showResult:  showResult,
		description: description,
	}
}

func (c ExecCommand) Run() (string, error) {
	out, err := exec.Command(c.exec, c.args...).Output()
	outAsString := fmt.Sprintf("%s", out)
	return outAsString, err
}

func (c ExecCommand) ShowResult() bool {
	return c.showResult
}

func (c ExecCommand) Description() string {
	return c.description
}
