// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package launchd

import (
	"fmt"
	"os/exec"
)

// StartCommand loads and starts a launchd job
type StartCommand struct {
	plistPath string
	label     string
}

// NewStartCommand creates a StartCommand
func NewStartCommand(plistPath string, label string) StartCommand {
	return StartCommand{
		plistPath: plistPath,
		label:     label,
	}
}

// Run runs the exec command
func (c StartCommand) Run(_ string, _ []string) (string, error) {
	if _, err := exec.Command("/bin/launchctl", "unload", c.plistPath).CombinedOutput(); err != nil {
		return "", fmt.Errorf("Error in launchctl unload: %s", err)
	}

	if _, err := exec.Command("/bin/launchctl", "load", c.plistPath).CombinedOutput(); err != nil {
		return "", fmt.Errorf("Error in launchctl load: %s", err)
	}

	if _, err := exec.Command("/bin/launchctl", "start", c.label).CombinedOutput(); err != nil {
		return "", fmt.Errorf("Error in launchctl start: %s", err)
	}

	return "", nil
}

// Stop a launchd job
func Stop(label string) (string, error) {
	if _, err := exec.Command("/bin/launchctl", "stop", label).CombinedOutput(); err != nil {
		return "", fmt.Errorf("Error in launchctl stop: %s", err)
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

// Label returns job label
func (c StartCommand) Label() string {
	return c.label
}
