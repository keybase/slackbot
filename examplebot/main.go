// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"log"

	"github.com/keybase/slackbot"
)

type runner struct{}

func main() {
	config := slackbot.NewConfig(false, false)
	bot, err := slackbot.NewBot(slackbot.GetTokenFromEnv(), "examplebot", "", config)
	if err != nil {
		log.Fatal(err)
	}

	// Command that runs and shows date
	bot.AddCommand("date", slackbot.NewExecCommand("/bin/date", nil, true, "Show the current date", config))

	// Commands for config, pausing and doing dry runs
	bot.AddCommand("pause", slackbot.NewPauseCommand(config))
	bot.AddCommand("resume", slackbot.NewResumeCommand(config))
	bot.AddCommand("config", slackbot.NewShowConfigCommand(config))
	bot.AddCommand("toggle-dryrun", slackbot.NewToggleDryRunCommand(bot.Config()))

	// Extension as default command with help
	ext := &extension{}
	runFn := func(channel string, args []string) (string, error) {
		return ext.Run(bot, channel, args)
	}
	bot.SetDefault(slackbot.NewFuncCommand(runFn, "Extension", bot.Config()))
	bot.SetHelp(bot.HelpMessage() + "\n\n" + ext.Help(bot))

	// Connect to slack and listen
	bot.Listen()
}
