// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/keybase/slackbot"
	"github.com/keybase/slackbot/jenkins"
	"gopkg.in/alecthomas/kingpin.v2"
)

func setEnv(name string, val string) error {
	_, err := setEnvCommand(name, val).Run([]string{})
	return err
}

func kingpinHandler(args []string) (string, error) {
	app := kingpin.New("slackbot", "Command parser for slackbot")
	app.Terminate(nil)
	stringBuffer := new(bytes.Buffer)
	app.Writer(stringBuffer)

	build := app.Command("build", "Build things")
	buildPlease := build.Command("please", "Start a build")
	buildTest := build.Command("test", "Start a test build")

	buildAndroid := build.Command("android", "Start an android build")

	clientCommit := buildPlease.Flag("client-commit", "Build a specific client commit hash").String()
	kbfsCommit := buildPlease.Flag("kbfs-commit", "Build a specific kbfs commit hash").String()

	testClientCommit := buildTest.Flag("client-commit", "Build a specific client commit hash").String()
	testKbfsCommit := buildTest.Flag("kbfs-commit", "Build a specific kbfs commit hash").String()

	release := app.Command("release", "Release things")
	releasePromote := release.Command("promote", "Promote a release to public")
	releaseToPromote := releasePromote.Arg("release-to-promote", "Promote a specific release to public immediately").Required().String()

	cancel := build.Command("cancel", "Cancel any existing builds")
	buildWindows := build.Command("windows", "start a windows build")
	testWindows := buildTest.Arg("windows", "Start a windows test build").String()
	cancelWindows := cancel.Arg("windows", "Cancel last windows build").String()
	// Make sure context parses otherwise showing Usage on error will fail later
	if _, perr := app.ParseContext(args); perr != nil {
		return "", perr
	}

	cmd, err := app.Parse(args)

	if err != nil && stringBuffer.Len() == 0 {
		log.Printf("Error in parsing command: %s. got %s", args, err)
		io.WriteString(stringBuffer, fmt.Sprintf("I don't know what you mean by `%s`.\nError: `%s`\nHere's my usage:\n\n", strings.Join(args, " "), err.Error()))
		// Print out help page if there was an error parsing command
		app.Usage([]string{})
	}

	if stringBuffer.Len() > 0 {
		return slackbot.SlackBlockQuote(stringBuffer.String()), nil
	}

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

		return buildStartCommand().Run(emptyArgs)

	case buildAndroid.FullCommand():
		return buildAndroidCommand().Run(emptyArgs)

	case cancel.FullCommand():
		return buildStopCommand().Run(emptyArgs)

	case releasePromote.FullCommand():
		err = setEnv("RELEASE_TO_PROMOTE", *releaseToPromote)
		if err != nil {
			return "", err
		}
		return releasePromoteCommand().Run(emptyArgs)

	case buildTest.FullCommand():
		err = setEnv("CLIENT_COMMIT", *testClientCommit)

		if err != nil {
			return "", err
		}

		err = setEnv("KBFS_COMMIT", *testKbfsCommit)

		if err != nil {
			return "", err
		}

		return buildStartTestCommand().Run(emptyArgs)
	return cmd, nil
}

func addCommands(bot *slackbot.Bot) {
	bot.AddCommand("pause", slackbot.ConfigCommand{
		Desc: "Pause any future builds",
		Updater: func(c slackbot.Config) (slackbot.Config, error) {
			c.Paused = true
			return c, nil
		},
	})

	bot.AddCommand("resume", slackbot.ConfigCommand{
		Desc: "Continue any future builds",
		Updater: func(c slackbot.Config) (slackbot.Config, error) {
			c.Paused = false
			return c, nil
		},
	})

	bot.AddCommand("config", slackbot.ConfigCommand{
		Desc: "List current config",
		Updater: func(c slackbot.Config) (slackbot.Config, error) {
			return c, nil
		},
	})

	bot.AddCommand("toggle-dryrun", slackbot.ToggleDryRunCommand{})

	bot.AddCommand("build", slackbot.FuncCommand{
		Desc: "Build all the things!",
		Fn:   kingpinHandler,
	})

	bot.AddCommand("restart", restartCommand())

	bot.AddCommand("date", slackbot.NewExecCommand("/bin/date", nil, true, "Show the current date"))
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
