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
	if bot.Config().DryRun() {
		return fmt.Sprintf("I would have run a launchd job (%s)", script.Label), nil
	}

	if bot.Config().Paused() {
		return fmt.Sprintf("I'm paused so I can't do that, but I would have run a launchd job (%s)", script.Label), nil
	}

	// Write job plist
	path, err := env.WritePlist(script)
	if err != nil {
		return "", err
	}

	// Remove previous log
	if err := launchd.CleanupLog(env, script.Label); err != nil {
		return "", err
	}

	msg := fmt.Sprintf("Starting job `%s`. To cancel run `!%s cancel %s`", script.Label, bot.Name(), script.Label)
	bot.SendMessage(msg, channel)
	return launchd.NewStartCommand(path, script.Label).Run("", nil)
}

func addBasicCommands(bot slackbot.Bot) {
	bot.AddCommand("date", slackbot.NewExecCommand("/bin/date", nil, true, "Show the current date", bot.Config()))
	bot.AddCommand("pause", slackbot.NewPauseCommand(bot.Config()))
	bot.AddCommand("resume", slackbot.NewResumeCommand(bot.Config()))
	bot.AddCommand("config", slackbot.NewShowConfigCommand(bot.Config()))
	bot.AddCommand("toggle-dryrun", slackbot.NewToggleDryRunCommand(bot.Config()))
	bot.AddCommand("restart", slackbot.NewExecCommand("/bin/launchctl", []string{"stop", bot.Label()}, false, "Restart the bot", bot.Config()))
}

type extension interface {
	Run(b slackbot.Bot, channel string, args []string) (string, error)
	Help(bot slackbot.Bot) string
}

func main() {
	name := os.Getenv("BOT_NAME")
	var label string
	var ext extension
	switch name {
	case "keybot":
		ext = &keybot{}
		label = "keybase.keybot"
	case "darwinbot":
		ext = &darwinbot{}
		label = "keybase.darwinbot"
	default:
		log.Fatal("Invalid BOT_NAME")
	}

	bot, err := slackbot.NewBot(slackbot.GetTokenFromEnv(), name, label, slackbot.ReadConfigOrDefault())
	if err != nil {
		log.Fatal(err)
	}

	addBasicCommands(bot)

	// Extension
	runFn := func(channel string, args []string) (string, error) {
		return ext.Run(bot, channel, args)
	}
	bot.SetDefault(slackbot.NewFuncCommand(runFn, "Extension", bot.Config()))
	bot.SetHelp(bot.HelpMessage() + "\n\n" + ext.Help(bot))

	bot.SendMessage("I'm running.", os.Getenv("SLACK_CHANNEL"))
	bot.Listen()
}
