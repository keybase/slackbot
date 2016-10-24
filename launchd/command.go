// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package launchd

import (
	"fmt"
	"os/exec"
)

// StartCommand loads and starts a launchd job
type StartCommand struct {
	label string
}

// NewStartCommand creates a StartCommand
func NewStartCommand(label string) StartCommand {
	return StartCommand{
		label: label,
	}
}

// Run runs the exec command
func (c StartCommand) Run(_ string, _ []string) (string, error) {
	// config := ReadConfigOrDefault()
	//
	// if config.DryRun {
	// 	return fmt.Sprintf("I would have run a launchd job (%s)", c.label), nil
	// }
	//
	// if config.Paused {
	// 	return fmt.Sprintf("I'm paused so I can't do that, but I would have run a launchd job (%s)", c.label), nil
	// }

	if _, err := exec.Command("/bin/launchctl", "load", c.label).CombinedOutput(); err != nil {
		return "", err
	}

	if _, err := exec.Command("/bin/launchctl", "start", c.label).CombinedOutput(); err != nil {
		return "", err
	}

	return "", nil
}

// ShowResult decides whether to show the results from the exec
func (c StartCommand) ShowResult() bool {
	return false
}

// Description describes the command
func (c StartCommand) Description() string {
	return fmt.Sprintf("Run launchd job (%s)", c.label)
}
