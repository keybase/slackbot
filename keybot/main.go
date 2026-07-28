// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/keybase/go-keybase-chat-bot/kbchat/types/chat1"

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
	return "0"
}

func runScript(bot *slackbot.Bot, channel string, env launchd.Env, script launchd.Script, ignorePause bool) (string, error) {
	if bot.Config().DryRun() {
		return fmt.Sprintf("I would have run a launchd job (%s)\nPath: %#v\nEnvVars: %#v", script.Label, script.Path, script.EnvVars), nil
	}

	if bot.Config().Paused() && !ignorePause {
		return fmt.Sprintf("I'm paused so I can't do that, but I would have run a launchd job (%s)", script.Label), nil
	}

	path, err := env.WritePlist(script)
	if err != nil {
		return "", err
	}

	if err := launchd.CleanupLog(env, script.Label); err != nil {
		return "", err
	}

	msg := fmt.Sprintf("I'm starting the job `%s`. To cancel run `!%s cancel %s`", script.Label, bot.Name(), script.Label)
	bot.SendMessage(msg, channel)
	return launchd.NewStartCommand(path, script.Label).Run("", nil)
}

func addBasicCommands(bot *slackbot.Bot) {
	bot.AddCommand("date", slackbot.NewExecCommand("/bin/date", nil, true, "Show the current date", bot.Config()))
	bot.AddCommand("pause", slackbot.NewPauseCommand(bot.Config()))
	bot.AddCommand("resume", slackbot.NewResumeCommand(bot.Config()))
	bot.AddCommand("config", slackbot.NewShowConfigCommand(bot.Config()))
	bot.AddCommand("toggle-dryrun", slackbot.NewToggleDryRunCommand(bot.Config()))
	if runtime.GOOS != "windows" {
		bot.AddCommand("restart", slackbot.NewExecCommand("/bin/launchctl", []string{"stop", bot.Label()}, false, "Restart the bot", bot.Config()))
	}
}

type extension interface {
	Run(b *slackbot.Bot, channel string, args []string) (string, error)
	Help(bot *slackbot.Bot) string
	Advertisements(bot *slackbot.Bot) []chat1.UserBotCommandInput
}

func main() {
	name := os.Getenv("BOT_NAME")
	var label string
	var ext extension

	channel := os.Getenv("KEYBASE_CHAT_CONVID")
	backend, err := slackbot.NewKeybaseChatBotBackend(name, channel, slackbot.KeybaseRunOptionsFromEnv(name))
	if err != nil {
		log.Fatalf("failed to initialize Keybase backend: %s", err)
	}

	switch name {
	case "keybot":
		ext = &keybot{}
		label = "keybase.keybot"
	case "winbot":
		ext = &winbot{}
		label = "keybase.winbot"
	default:
		log.Fatal("Invalid BOT_NAME")
	}

	bot := slackbot.NewBot(slackbot.ReadConfigOrDefault(), name, label, backend)
	addBasicCommands(bot)

	runFn := func(channel string, args []string) (string, error) {
		return ext.Run(bot, channel, args)
	}
	bot.SetDefault(slackbot.NewFuncCommand(runFn, "Extension", bot.Config()))
	bot.SetHelp(bot.HelpMessage() + "\n\n" + ext.Help(bot))
	bot.AddAdvertisements(ext.Advertisements(bot)...)

	bot.SendMessage("I'm running.", channel)
	if err := bot.Listen(); err != nil {
		log.Fatalf("bot listener failed: %s", err)
	}
}
