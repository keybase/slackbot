// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"bytes"
	"log"
	"os"

	"github.com/keybase/slackbot"
	"gopkg.in/alecthomas/kingpin.v2"
)

func setEnv(name string, val string) error {
	_, err := slackbot.NewExecCommand("/bin/launchctl", []string{"setenv", name, val}, false, "Set the env").Run([]string{})
	return err
}

func kingpinHandler(args []string) (string, error) {
	app := kingpin.New("slackbot", "Command parser for slackbot")
	app.Terminate(nil)
	stringBuffer := new(bytes.Buffer)
	app.Writer(stringBuffer)

	build := app.Command("build", "Build things")

	clientCommit := build.Flag("client-commit", "Build a specific client commit hash").String()
	kbfsCommit := build.Flag("kbfs-commit", "Build a specific kbfs commit hash").String()

	buildPlease := build.Command("please", "Start a build")
	buildTest := build.Command("test", "Start a test build")
	cancel := build.Command("cancel", "Cancel any existing builds. commit flags don't affect this.")

	cmd, err := app.Parse(args)

	if stringBuffer.Len() > 0 {
		return stringBuffer.String(), nil
	}

	if err != nil {
		return err.Error(), nil
	}

	buildStart := slackbot.NewExecCommand("/bin/launchctl", []string{"start", "keybase.prerelease"}, false, "Perform a build")
	buildStop := slackbot.NewExecCommand("/bin/launchctl", []string{"stop", "keybase.prerelease"}, false, "Cancel a running build")
	buildStartTest := slackbot.NewExecCommand("/bin/launchctl", []string{"start", "keybase.prerelease.test"}, false, "Test the build")

	emptyArgs := []string{}

	switch cmd {
	case buildPlease.FullCommand():
		err = setEnv("CLIENT_COMMIT", *clientCommit)

		if err != nil {
			return "", err
		}

		err = setEnv("KBFS_COMMIT", *kbfsCommit)

		if err != nil {
			return "", err
		}

		return buildStart.Run(emptyArgs)
	case cancel.FullCommand():
		return buildStop.Run(emptyArgs)
	case buildTest.FullCommand():
		return buildStartTest.Run(emptyArgs)
	}

	return cmd, nil
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

	bot.AddCommand("pause", slackbot.ConfigCommand{
		"Pause any future builds",
		func(c slackbot.Config) (slackbot.Config, error) {
			c.Paused = true
			return c, nil
		},
	})

	bot.AddCommand("start", slackbot.ConfigCommand{
		"Continue any future builds",
		func(c slackbot.Config) (slackbot.Config, error) {
			c.Paused = false
			return c, nil
		},
	})

	bot.AddCommand("ls-config", slackbot.ConfigCommand{
		"List current config",
		func(c slackbot.Config) (slackbot.Config, error) {
			return c, nil
		},
	})

	bot.AddCommand("toggle-dryrun", slackbot.ToggleDryRunCommand{})

	bot.AddCommand("build", slackbot.FuncCommand{"Build all the things!", kingpinHandler})

	bot.AddCommand("restart", slackbot.NewExecCommand("/bin/launchctl", []string{"stop", "keybase.keybot"}, false, "Restart the bot"))

	bot.AddCommand("date", slackbot.NewExecCommand("/bin/date", nil, true, "Show the current date"))

	log.Println("Started keybot")
	bot.Listen()
}
