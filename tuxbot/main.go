// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"bytes"
	"log"
	"os"

	"github.com/keybase/slackbot"
	"github.com/keybase/slackbot/cli"
	"gopkg.in/alecthomas/kingpin.v2"
)

func kingpinTuxbotHandler(args []string) (string, error) {
	app := kingpin.New("tuxbot", "Command parser for tuxbot")
	app.Terminate(nil)
	stringBuffer := new(bytes.Buffer)
	app.Writer(stringBuffer)

	build := app.Command("build", "Build things")
	buildLinux := build.Command("linux", "Start a linux build")

	cmd, usage, err := cli.Parse(app, args, stringBuffer)
	if usage != "" || err != nil {
		return usage, err
	}

	switch cmd {
	case buildLinux.FullCommand():
		return slackbot.NewExecCommand("bash", []string{"-c", "systemctl --user start keybase.prerelease.service && echo 'SUCCESS'"}, true, "Perform a linux build").Run([]string{})
	}

	return cmd, nil
}

func addCommands(bot *slackbot.Bot) {
	bot.AddCommand("date", slackbot.NewExecCommand("/bin/date", nil, true, "Show the current date"))
	bot.AddCommand("pause", slackbot.NewPauseCommand())
	bot.AddCommand("resume", slackbot.NewResumeCommand())
	bot.AddCommand("config", slackbot.NewListConfigCommand())
	bot.AddCommand("toggle-dryrun", slackbot.ToggleDryRunCommand{})

	bot.AddCommand("build", slackbot.FuncCommand{
		Desc: "Build all the things!",
		Fn:   kingpinTuxbotHandler,
	})
}

func main() {
	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		log.Fatal("SLACK_TOKEN is not set")
	}

	bot, err := slackbot.NewBot(token)
	if err != nil {
		log.Fatal(err)
	}

	addCommands(bot)

	log.Println("Started keybot")
	bot.Listen()
}
