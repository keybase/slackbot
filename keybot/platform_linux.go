// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"log"

	"github.com/keybase/slackbot"
)

func setEnvCommand(name string, val string) slackbot.ExecCommand {
	log.Printf("WARNING: setEnvCommand(%q, %q) is a NO-OP on Linux", name, val)
	return slackbot.NewExecCommand("true", []string{}, true, "Set the env")
}

func buildDarwinCommand() slackbot.ExecCommand {
	return slackbot.NewExecCommand("bash", []string{"-c", "systemctl --user start keybase.prerelease.service && echo 'SUCCESS'"}, true, "Perform a build")
}

func buildDarwinCancelCommand() slackbot.ExecCommand {
	return slackbot.NewExecCommand("bash", []string{"-c", "echo not implemented; false"}, true, "Cancel a running build")
}

func buildDarwinTestCommand() slackbot.ExecCommand {
	return slackbot.NewExecCommand("bash", []string{"-c", "echo not implemented; false"}, true, "Test the build")
}

func buildAndroidCommand() slackbot.ExecCommand {
	return slackbot.NewExecCommand("bash", []string{"-c", "echo not implemented; false"}, true, "Perform an alpha build")
}

func restartCommand() slackbot.ExecCommand {
	return slackbot.NewExecCommand("bash", []string{"-c", "echo not implemented; false"}, true, "Restart the bot")
}

func releasePromoteCommand() slackbot.ExecCommand {
	return slackbot.NewExecCommand("bash", []string{"-c", "echo not implemented; false"}, true, "Promote a specific release to public")
}
