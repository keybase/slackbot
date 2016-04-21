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

func setDarwinEnv(name string, val string) error {
	_, err := slackbot.NewExecCommand("/bin/launchctl", []string{"setenv", name, val}, false, "Set the env").Run([]string{})
	return err
}

func isParseContextValid(app *kingpin.Application, args []string) error {
	if pcontext, perr := app.ParseContext(args); pcontext == nil {
		return perr
	}
	return nil
}

func parse(app *kingpin.Application, args []string, stringBuffer *bytes.Buffer) (string, string, error) {
	// Make sure context is valid otherwise showing Usage on error will fail later.
	// This is a workaround for a kingpin bug.
	if err := isParseContextValid(app, args); err != nil {
		return "", "", err
	}

	cmd, err := app.Parse(args)

	if err != nil && stringBuffer.Len() == 0 {
		log.Printf("Error in parsing command: %s. got %s", args, err)
		io.WriteString(stringBuffer, fmt.Sprintf("I don't know what you mean by `%s`.\nError: `%s`\nHere's my usage:\n\n", strings.Join(args, " "), err.Error()))
		// Print out help page if there was an error parsing command
		app.Usage([]string{})
	}

	if stringBuffer.Len() > 0 {
		return "", slackbot.SlackBlockQuote(stringBuffer.String()), nil
	}

	return cmd, "", err
}

func kingpinKeybotHandler(args []string) (string, error) {
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
	releaseToPromote := releasePromote.Arg("release-to-promote", "Promote a specific release to public immediately").Required().String()

	buildWindows := build.Command("windows", "start a windows build")
	testWindows := test.Command("windows", "Start a windows test build")
	cancelWindows := cancel.Command("windows", "Cancel last windows build")
	cancelWindowsQueueID := cancelWindows.Arg("quid", "Queue id of build to stop").Required().String()

	cmd, usage, err := parse(app, args, stringBuffer)
	if usage != "" || err != nil {
		return usage, err
	}

	emptyArgs := []string{}

	if err := setDarwinEnv("CLIENT_COMMIT", *clientCommit); err != nil {
		return "", err
	}
	if err := setDarwinEnv("KBFS_COMMIT", *kbfsCommit); err != nil {
		return "", err
	}

	switch cmd {
	// Darwin
	case buildDarwin.FullCommand():
		return slackbot.NewExecCommand("/bin/launchctl", []string{"start", "keybase.prerelease"}, false, "Perform a build").Run(emptyArgs)
	case testDarwin.FullCommand():
		return slackbot.NewExecCommand("/bin/launchctl", []string{"start", "keybase.prerelease.test"}, false, "Test the build").Run(emptyArgs)
	case cancelDarwin.FullCommand():
		return slackbot.NewExecCommand("/bin/launchctl", []string{"stop", "keybase.prerelease"}, false, "Cancel a running build").Run(emptyArgs)

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
		return slackbot.NewExecCommand("/bin/launchctl", []string{"start", "keybase.android"}, false, "Perform an android build").Run(emptyArgs)
	case buildIOS.FullCommand():
		return slackbot.NewExecCommand("/bin/launchctl", []string{"start", "keybase.ios"}, false, "Perform an ios build").Run(emptyArgs)

	case releasePromote.FullCommand():
		err = setDarwinEnv("RELEASE_TO_PROMOTE", *releaseToPromote)
		if err != nil {
			return "", err
		}
		return slackbot.NewExecCommand("/bin/launchctl", []string{"start", "keybase.prerelease.promotearelease"}, false, "Promote a specific release to public").Run(emptyArgs)
	}
	return cmd, nil
}

func kingpinTuxbotHandler(args []string) (string, error) {
	app := kingpin.New("tuxbot", "Command parser for tuxbot")
	app.Terminate(nil)
	stringBuffer := new(bytes.Buffer)
	app.Writer(stringBuffer)

	build := app.Command("build", "Build things")
	buildLinux := build.Command("linux", "Start a linux build")

	cmd, usage, err := parse(app, args, stringBuffer)
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
	// Add commands for all bots here
	bot.AddCommand("date", slackbot.NewExecCommand("/bin/date", nil, true, "Show the current date"))

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

	if bot.User() == "tuxbot" {
		// Tuxbot only knows how to build linux
		bot.AddCommand("build", slackbot.FuncCommand{
			Desc: "Build all the things!",
			Fn:   kingpinTuxbotHandler,
		})
	} else {
		// Keybot can do more
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
