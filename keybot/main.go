// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/keybase/slackbot"
	"github.com/keybase/slackbot/launchd"
)

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func boolToEnvString(b bool) string {
	if b {
		return "1"
	}
	return ""
}

func runScript(bot slackbot.Bot, channel string, env launchd.Env, script launchd.Script) (string, error) {
	path, err := env.WritePlist(script)
	if err != nil {
		return "", err
	}
	msg := fmt.Sprintf("Starting job `%s`. To cancel run `!%s cancel %s`", script.Label, bot.Name(), script.Label)
	bot.SendMessage(msg, channel)
	return launchd.NewStartCommand(path, script.Label).Run("", nil)
}

func addCommands(bot slackbot.Bot) {
	helpMessage := bot.HelpMessage()

	bot.AddCommand("date", slackbot.NewExecCommand("/bin/date", nil, true, "Show the current date"))
	bot.AddCommand("pause", slackbot.NewPauseCommand())
	bot.AddCommand("resume", slackbot.NewResumeCommand())
	bot.AddCommand("config", slackbot.NewListConfigCommand())
	bot.AddCommand("toggle-dryrun", slackbot.ToggleDryRunCommand{})
	bot.AddCommand("restart", slackbot.NewExecCommand("/bin/launchctl", []string{"stop", bot.Label()}, false, "Restart the bot"))

	jobHelp, _ := bot.Runner().Run(bot, "", nil)
	helpMessage = helpMessage + "\n\n" + jobHelp
	bot.SetHelp(helpMessage)

	bot.SetDefault(slackbot.FuncCommand{
		Fn: bot.Runner().Run,
	})
}

func main() {
	name := os.Getenv("BOT_NAME")
	var label string
	var runner slackbot.Runner
	switch name {
	case "keybot":
		runner = &keybot{}
		label = "keybase.keybot"
	case "darwinbot":
		runner = &darwinbot{}
		label = "keybase.darwinbot"
	default:
		log.Fatal("Invalid BOT_NAME")
	}

	bot, err := slackbot.NewBot(slackbot.GetTokenFromEnv(), name, label, runner)
	if err != nil {
		log.Fatal(err)
	}

	addCommands(bot)

	log.Printf("Started %s\n", name)
	bot.Listen()
}
