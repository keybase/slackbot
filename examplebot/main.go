// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"log"

	"github.com/keybase/slackbot"
)

type runner struct{}

func main() {
	bot, err := slackbot.NewBot(slackbot.GetTokenFromEnv(), "examplebot", "", nil)
	if err != nil {
		log.Fatal(err)
	}

	bot.AddCommand("date", slackbot.NewExecCommand("/bin/date", nil, true, "Show the current date"))
	bot.AddCommand("pause", slackbot.NewPauseCommand())
	bot.AddCommand("resume", slackbot.NewResumeCommand())

	jobHelp, _ := bot.Run("", nil)
	helpMessage := bot.HelpMessage()
	helpMessage = helpMessage + "\n\n" + jobHelp
	bot.SetHelp(helpMessage)

	bot.SetDefault(slackbot.FuncCommand{
		Fn: bot.Run,
	})

	bot.Listen()
}
