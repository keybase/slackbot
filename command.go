// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package lib

type Command struct {
	trigger    string   // Trigger without the ! (e.g. "build")
	execute    string   // Command to execute
	args       []string // Args for command
	showResult bool     // Whether to output result back to channel
}

func NewCommand(trigger string, execute string, args []string, showResult bool) Command {
	return Command{
		trigger:    trigger,
		execute:    execute,
		args:       args,
		showResult: showResult,
	}
}
