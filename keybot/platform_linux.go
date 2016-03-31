// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"github.com/keybase/slackbot"
	"log"
)

func setEnvCommand(name string, val string) slackbot.ExecCommand {
	log.Printf("WARNING: setEnvCommand(%q, %q) is a NO-OP on Linux", name, val)
	return slackbot.NewExecCommand("true", []string{}, false, "Set the env")
}

func buildStartCommand() slackbot.ExecCommand {
	return slackbot.NewExecCommand("bash", []string{"-c", "echo not implemented; false"}, false, "Perform a build")
}

func buildStopCommand() slackbot.ExecCommand {
	return slackbot.NewExecCommand("bash", []string{"-c", "echo not implemented; false"}, false, "Cancel a running build")
}

func buildStartTestCommand() slackbot.ExecCommand {
	return slackbot.NewExecCommand("bash", []string{"-c", "echo not implemented; false"}, false, "Test the build")
}

func buildAndroidCommand() slackbot.ExecCommand {
	return slackbot.NewExecCommand("bash", []string{"-c", "echo not implemented; false"}, false, "Perform an alpha build")
}

func restartCommand() slackbot.ExecCommand {
	return slackbot.NewExecCommand("bash", []string{"-c", "echo not implemented; false"}, false, "Restart the bot")
}
