// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"log"

	"github.com/keybase/slackbot"
)

func main() {
	backend, err := slackbot.NewSlackBotBackend(slackbot.GetTokenFromEnv())
	if err != nil {
		log.Fatal(err)
	}
	bot := slackbot.NewBot(slackbot.ReadConfigOrDefault(), "tuxbot", "", backend)

	bot.AddCommand("date", slackbot.NewExecCommand("/bin/date", nil, true, "Show the current date", bot.Config()))
	bot.AddCommand("pause", slackbot.NewPauseCommand(bot.Config()))
	bot.AddCommand("resume", slackbot.NewResumeCommand(bot.Config()))
	bot.AddCommand("config", slackbot.NewShowConfigCommand(bot.Config()))
	bot.AddCommand("toggle-dryrun", slackbot.NewToggleDryRunCommand(bot.Config()))

	// Extension
	ext := &tuxbot{bot: bot}
	runFn := func(channel string, args []string) (string, error) {
		return ext.Run(bot, channel, args)
	}
	bot.SetDefault(slackbot.NewFuncCommand(runFn, "Extension", bot.Config()))
	bot.SetHelp(bot.HelpMessage() + "\n\n" + ext.Help(bot))

	log.Println("Started tuxbot")
	bot.Listen()
}
