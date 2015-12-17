// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"log"
	"os"

	"github.com/keybase/slackbot"
)

func main() {
	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		log.Fatal("SLACK_TOKEN is not set")
	}

	bot, err := slackbot.NewBot(token)
	if err != nil {
		log.Fatal(err)
	}

	// For debugging
	bot.AddCommand(slackbot.NewCommand("date", "/bin/date", nil, true))

	bot.AddCommand(slackbot.NewCommand("build", "/bin/launchctl", []string{"start", "keybase.prerelease"}, false))

	bot.Listen()
}
