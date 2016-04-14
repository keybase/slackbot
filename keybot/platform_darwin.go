// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"github.com/keybase/slackbot"
)

func setEnvCommand(name string, val string) slackbot.ExecCommand {
	return slackbot.NewExecCommand("/bin/launchctl", []string{"setenv", name, val}, false, "Set the env")
}

func buildDarwinCommand() slackbot.ExecCommand {
	return slackbot.NewExecCommand("/bin/launchctl", []string{"start", "keybase.prerelease"}, false, "Perform a build")
}

func buildDarwinCancelCommand() slackbot.ExecCommand {
	return slackbot.NewExecCommand("/bin/launchctl", []string{"stop", "keybase.prerelease"}, false, "Cancel a running build")
}

func buildDarwinTestCommand() slackbot.ExecCommand {
	return slackbot.NewExecCommand("/bin/launchctl", []string{"start", "keybase.prerelease.test"}, false, "Test the build")
}

func buildAndroidCommand() slackbot.ExecCommand {
	return slackbot.NewExecCommand("/bin/launchctl", []string{"start", "keybase.android.release"}, false, "Perform an alpha build")
}

func buildIOSCommand() slackbot.ExecCommand {
	return slackbot.NewExecCommand("/bin/launchctl", []string{"start", "keybase.ios.release"}, false, "Perform an testflight build")
}

func restartCommand() slackbot.ExecCommand {
	return slackbot.NewExecCommand("/bin/launchctl", []string{"stop", "keybase.keybot"}, false, "Restart the bot")
}

func releasePromoteCommand() slackbot.ExecCommand {
	return slackbot.NewExecCommand("/bin/launchctl", []string{"start", "keybase.prerelease.promotearelease"}, false, "Promote a specific release to public")
}
