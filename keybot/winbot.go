// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/keybase/slackbot"
	"github.com/keybase/slackbot/cli"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type winbot struct{}

func (d *winbot) Run(bot slackbot.Bot, channel string, args []string) (string, error) {
	app := kingpin.New("winbot", "Job command parser for winbot")
	app.Terminate(nil)
	stringBuffer := new(bytes.Buffer)
	app.Writer(stringBuffer)

	build := app.Command("build", "Build things")

	buildWindows := build.Command("windows", "Start a windows build")
	buildWindowsTest := buildWindows.Flag("test", "Whether build is for testing").Bool()
	buildWindowsCientCommit := buildWindows.Flag("client-commit", "Build a specific client commit").String()
	buildWindowsKbfsCommit := buildWindows.Flag("kbfs-commit", "Build a specific kbfs commit").String()
	buildWindowsSkipCI := buildWindows.Flag("skip-ci", "Whether to skip CI").Bool()

	cancel := app.Command("cancel", "Cancel current")

	dumplogCmd := app.Command("dumplog", "Show the last log file")

	logFileName := path.Join(os.TempDir(), "keybase.build.windows.log")

	cmd, usage, cmdErr := cli.Parse(app, args, stringBuffer)
	if usage != "" || cmdErr != nil {
		return usage, cmdErr
	}

	if bot.Config().DryRun() {
		return fmt.Sprintf("I would have run: `%#v`", cmd), nil
	}

	var currentCmd *exec.Cmd

	switch cmd {
	case cancel.FullCommand():
		tempCmd := currentCmd
		if tempCmd == nil {
			return "No build running", nil
		}
		if err := tempCmd.Process.Kill(); err != nil {
			return "failed to cancel build", err
		}

	case buildWindows.FullCommand():
		smokeTest := true
		skipCI := *buildWindowsSkipCI
		testBuild := *buildWindowsTest
		// Don't smoke, wait for CI or promote test if custom build
		if *buildWindowsCientCommit != "" || *buildWindowsKbfsCommit != "" {
			smokeTest = false
			skipCI = true
			testBuild = true
		}

		updateChannel := "None"
		if testBuild {
			updateChannel = "Test"
		} else if smokeTest {
			updateChannel = "Smoke"
			if !skipCI {
				updateChannel = "SmokeCI"
			}
		}

		// TODO: use SMOKE_TEST and TEST like other scripts
		cmd := exec.Command(
			"cmd", "/c",
			path.Join(os.Getenv("GOPATH"), "src/github.com/keybase/client/packaging/windows/dorelease.cmd"),
			">",
			logFileName,
			"2>&1")
		cmd.Env = append(os.Environ(),
			"ClientRevision="+*buildWindowsCientCommit,
			"KbfsRevision="+*buildWindowsKbfsCommit,
			"SKIP_CI="+boolToEnvString(skipCI),
			"UpdateChannel="+updateChannel,
			"SlackBot=1",
		)
		msg := fmt.Sprintf("I'm starting the job `windows build`. To cancel run `!%s cancel`", bot.Name())
		bot.SendMessage(msg, channel)
		currentCmd = cmd
		err := cmd.Run()
		currentCmd = nil

		bucketName := os.Getenv("BUCKET_NAME")
		if bucketName == "" {
			bucketName = "prerelease.keybase.io"
		}
		sendLogCmd := exec.Command(
			path.Join(os.Getenv("GOPATH"), "src/github.com/keybase/release/release.exe"),
			"save-log",
			"--bucket-name="+bucketName,
			"--path="+logFileName,
		)
		urlBytes, err2 := sendLogCmd.Output()
		if err2 != nil {
			msg := fmt.Sprintf("Finished the job `windows build`, log upload error %s", err2.Error())
			bot.SendMessage(msg, channel)
		} else {
			msg := fmt.Sprintf("Finished the job `windows build`, view log at %s", string(urlBytes[:]))
			bot.SendMessage(msg, channel)
		}

		if err != nil {
			index := 0
			logContents, err := ioutil.ReadFile(logFileName)
			if err != nil {
				return "Error reading " + logFileName, err
			}
			if len(logContents) > 160 {
				index = len(logContents) - 160
			}
			return string(logContents[index:]), err
		}

	case dumplogCmd.FullCommand():
		logContents, err := ioutil.ReadFile(logFileName)
		if err != nil {
			return "Error reading " + logFileName, err
		}
		return string(logContents[:]), nil
	}
	return cmd, nil
}

func (d *winbot) Help(bot slackbot.Bot) string {
	out, err := d.Run(bot, "", nil)
	if err != nil {
		return fmt.Sprintf("Error getting help: %s", err)
	}
	return out
}
