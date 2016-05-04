// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"bytes"
	"log"
	"os"

	"github.com/keybase/slackbot"
	"github.com/keybase/slackbot/cli"
	"github.com/keybase/slackbot/jenkins"
	"gopkg.in/alecthomas/kingpin.v2"
)

func setDarwinEnv(name string, val string) error {
	_, err := slackbot.NewExecCommand("/bin/launchctl", []string{"setenv", name, val}, false, "Set the env").Run("", []string{})
	return err
}

func kingpinKeybotHandler(channel string, args []string) (string, error) {
	app := kingpin.New("keybot", "Command parser for keybot")
	app.Terminate(nil)
	stringBuffer := new(bytes.Buffer)
	app.Writer(stringBuffer)

	build := app.Command("build", "Build things")
	test := app.Command("test", "Test")
	cancel := app.Command("cancel", "Cancel")

	clientCommit := build.Flag("client-commit", "Build a specific client commit hash").String()
	kbfsCommit := build.Flag("kbfs-commit", "Build a specific kbfs commit hash").String()

	buildDarwin := build.Command("darwin", "Start a darwin build")
	testDarwin := test.Command("darwin", "Start a darwin test build")
	cancelDarwin := cancel.Command("darwin", "Cancel the darwin build")

	buildAndroid := build.Command("android", "Start an android build")
	buildIOS := build.Command("ios", "Start an ios build")

	release := app.Command("release", "Release things")
	releasePromote := release.Command("promote", "Promote a release to public")
	releaseToPromote := releasePromote.Arg("release-to-promote", "Promote a specific release to public immediately").String()

	buildWindows := build.Command("windows", "start a windows build")
	testWindows := test.Command("windows", "Start a windows test build")
	cancelWindows := cancel.Command("windows", "Cancel last windows build")
	cancelWindowsQueueID := cancelWindows.Arg("quid", "Queue id of build to stop").Required().String()

	cmd, usage, err := cli.Parse(app, args, stringBuffer)
	if usage != "" || err != nil {
		return usage, err
	}

	if err := setDarwinEnv("CLIENT_COMMIT", *clientCommit); err != nil {
		return "", err
	}
	if err := setDarwinEnv("KBFS_COMMIT", *kbfsCommit); err != nil {
		return "", err
	}

	emptyArgs := []string{}
	switch cmd {
	// Darwin
	case buildDarwin.FullCommand():
		return slackbot.NewExecCommand("/bin/launchctl", []string{"start", "keybase.prerelease"}, false, "Perform a build").Run("", emptyArgs)
	case testDarwin.FullCommand():
		return slackbot.NewExecCommand("/bin/launchctl", []string{"start", "keybase.prerelease.test"}, false, "Test the build").Run("", emptyArgs)
	case cancelDarwin.FullCommand():
		return slackbot.NewExecCommand("/bin/launchctl", []string{"stop", "keybase.prerelease"}, false, "Cancel a running build").Run("", emptyArgs)

	// Windows
	case buildWindows.FullCommand():
		return jenkins.StartBuild(*clientCommit, *kbfsCommit, "")
	case testWindows.FullCommand():
		return jenkins.StartBuild(*clientCommit, *kbfsCommit, "update-windows-prod-test.json")
	case cancelWindows.FullCommand():
		jenkins.StopBuild(*cancelWindowsQueueID)
		out := "Issued stop for " + *cancelWindowsQueueID
		return out, nil

	// Android
	case buildAndroid.FullCommand():
		return slackbot.NewExecCommand("/bin/launchctl", []string{"start", "keybase.android"}, false, "Perform an android build").Run("", emptyArgs)
	case buildIOS.FullCommand():
		return slackbot.NewExecCommand("/bin/launchctl", []string{"start", "keybase.ios"}, false, "Perform an ios build").Run("", emptyArgs)

	case releasePromote.FullCommand():
		err = setDarwinEnv("RELEASE_TO_PROMOTE", *releaseToPromote)
		if err != nil {
			return "", err
		}
		return slackbot.NewExecCommand("/bin/launchctl", []string{"start", "keybase.prerelease.promotearelease"}, false, "Promote a release to public, takes an optional specific release").Run("", emptyArgs)
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
		Fn:   kingpinKeybotHandler,
	})

	bot.AddCommand("release", slackbot.FuncCommand{
		Desc: "Release all the things!",
		Fn:   kingpinKeybotHandler,
	})

	bot.AddCommand("test", slackbot.FuncCommand{
		Desc: "Test all the things!",
		Fn:   kingpinKeybotHandler,
	})

	bot.AddCommand("cancel", slackbot.FuncCommand{
		Desc: "Cancel all the things!",
		Fn:   kingpinKeybotHandler,
	})

	bot.AddCommand("restart", slackbot.NewExecCommand("/bin/launchctl", []string{"stop", "keybase.keybot"}, false, "Restart the bot"))
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
