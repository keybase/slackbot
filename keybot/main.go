// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/keybase/go-keybase-chat-bot/kbchat"

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

func runScript(bot *slackbot.Bot, channel string, env launchd.Env, script launchd.Script) (string, error) {
	if bot.Config().DryRun() {
		return fmt.Sprintf("I would have run a launchd job (%s)\nPath: %#v\nEnvVars: %#v", script.Label, script.Path, script.EnvVars), nil
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
}

func main() {
	name := os.Getenv("BOT_NAME")
	var err error
	var label string
	var ext extension
	var backend slackbot.BotBackend
	var hybrids []slackbot.HybridBackendMember
	var channel string

	// Set up Slack
	slackChannel := os.Getenv("SLACK_CHANNEL")
	slackBackend, err := slackbot.NewSlackBotBackend(slackbot.GetTokenFromEnv())
	if err != nil {
		log.Printf("failed to initialize Slack backend: %s", err)
	} else {
		hybrids = append(hybrids, slackbot.HybridBackendMember{
			Backend: slackBackend,
			Channel: slackChannel,
		})
	}

	// Set up Keybase
	var opts kbchat.RunOptions
	keybaseChannel := os.Getenv("KEYBASE_CHAT_CONVID")
	opts.KeybaseLocation = os.Getenv("KEYBASE_LOCATION")
	opts.HomeDir = os.Getenv("KEYBASE_HOME")
	oneshotUsername := os.Getenv("KEYBASE_ONESHOT_USERNAME")
	oneshotPaperkey := os.Getenv("KEYBASE_ONESHOT_PAPERKEY")
	if len(oneshotPaperkey) > 0 && len(oneshotUsername) > 0 {
		opts.Oneshot = &kbchat.OneshotOptions{
			Username: oneshotUsername,
			PaperKey: oneshotPaperkey,
		}
	}
	keybaseBackend, err := slackbot.NewKeybaseChatBotBackend(name, keybaseChannel, opts)
	if err != nil {
		log.Printf("failed to initialize Keybase backend: %s", err)
	} else {
		hybrids = append(hybrids, slackbot.HybridBackendMember{
			Backend: keybaseBackend,
			Channel: keybaseChannel,
		})
	}

	// Set up hybrid backend
	hybridChannel := ""
	hybridBackend := slackbot.NewHybridBackend(hybrids...)

	switch name {
	case "keybot":
		ext = &keybot{}
		label = "keybase.keybot"
		backend = slackBackend
		channel = slackChannel
	case "darwinbot":
		ext = &darwinbot{}
		label = "keybase.darwinbot"
		backend = hybridBackend
		channel = hybridChannel
	case "winbot":
		ext = &winbot{}
		label = "keybase.winbot"
		channel = slackChannel
		backend = slackBackend
	default:
		log.Fatal("Invalid BOT_NAME")
	}

	bot := slackbot.NewBot(slackbot.ReadConfigOrDefault(), name, label, backend)
	addBasicCommands(bot)

	// Extension
	runFn := func(channel string, args []string) (string, error) {
		return ext.Run(bot, channel, args)
	}
	bot.SetDefault(slackbot.NewFuncCommand(runFn, "Extension", bot.Config()))
	bot.SetHelp(bot.HelpMessage() + "\n\n" + ext.Help(bot))

	bot.SendMessage("I'm running.", channel)

	bot.Listen()
}
