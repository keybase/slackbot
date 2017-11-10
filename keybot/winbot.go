// Copyright 2017 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"sync"

	"github.com/keybase/slackbot"
	"github.com/keybase/slackbot/cli"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type winbot struct{}

const numLogLines = 10

// Keep track of the current build process, protected with a mutex,
// to support cancellation
var buildProcessMutex sync.Mutex
var buildProcess *os.Process

func (d *winbot) Run(bot slackbot.Bot, channel string, args []string) (string, error) {
	app := kingpin.New("winbot", "Job command parser for winbot")
	app.Terminate(nil)
	stringBuffer := new(bytes.Buffer)
	app.Writer(stringBuffer)

	buildWindows := app.Command("build", "Start a windows build")
	buildWindowsTest := buildWindows.Flag("test", "Whether build is for testing (skips CI and smoke)").Bool()
	buildWindowsCientCommit := buildWindows.Flag("client-commit", "Build a specific client commit").String()
	buildWindowsKbfsCommit := buildWindows.Flag("kbfs-commit", "Build a specific kbfs commit").String()
	buildWindowsSkipCI := buildWindows.Flag("skip-ci", "Whether to skip CI").Bool()
	buildWindowsSmoke := buildWindows.Flag("smoke", "Build a smoke pair").Bool()

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

	switch cmd {
	case cancel.FullCommand():
		buildProcessMutex.Lock()
		defer buildProcessMutex.Unlock()
		if buildProcess == nil {
			return "No build running", nil
		}
		if err := buildProcess.Kill(); err != nil {
			return "failed to cancel build", err
		}

	case buildWindows.FullCommand():
		smokeTest := *buildWindowsSmoke
		skipCI := *buildWindowsSkipCI
		testBuild := *buildWindowsTest

		updateChannel := "None"
		if testBuild {
			if smokeTest {
				return "Test and Smoke are exclusive options", nil
			}
			updateChannel = "Test"
		} else if smokeTest {
			updateChannel = "Smoke"
			if !skipCI {
				updateChannel = "SmokeCI"
			}
		}

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
			"SMOKE_TEST="+boolToEnvString(smokeTest),
			"TEST="+boolToEnvString(testBuild),
		)
		msg := fmt.Sprintf("I'm starting the job `windows build`. To cancel run `!%s cancel`", bot.Name())
		bot.SendMessage(msg, channel)

		go func() {
			err := cmd.Start()
			buildProcessMutex.Lock()
			buildProcess = cmd.Process
			buildProcessMutex.Unlock()
			err = cmd.Wait()

			bucketName := os.Getenv("BUCKET_NAME")
			if bucketName == "" {
				bucketName = "prerelease.keybase.io"
			}
			sendLogCmd := exec.Command(
				path.Join(os.Getenv("GOPATH"), "src/github.com/keybase/release/release.exe"),
				"save-log",
				"--maxsize=2000000",
				"--bucket-name="+bucketName,
				"--path="+logFileName,
			)
			resultMsg := "Finished the job `windows build`"
			if err != nil {
				resultMsg = "Error in job `windows build`"
				var lines [numLogLines]string
				// Send a log snippet too
				index := 0
				lineCount := 0

				f, err := os.Open(logFileName)
				if err != nil {
					bot.SendMessage("Error reading "+logFileName+": "+err.Error(), channel)
				}

				scanner := bufio.NewScanner(f)
				for scanner.Scan() {
					lines[lineCount%numLogLines] = scanner.Text()
					lineCount += 1
				}
				if err := scanner.Err(); err != nil {
					bot.SendMessage("Error scanning "+logFileName+": "+err.Error(), channel)
				}
				if lineCount > numLogLines {
					index = lineCount % numLogLines
					lineCount = numLogLines
				}
				snippet := "```\n"
				for i := 0; i < lineCount; i++ {
					snippet += lines[(i+index)%numLogLines] + "\n"
				}
				snippet += "```"
				bot.SendMessage(snippet, channel)
			}
			urlBytes, err2 := sendLogCmd.Output()
			if err2 != nil {
				msg := fmt.Sprintf("%s, log upload error %s", resultMsg, err2.Error())
				bot.SendMessage(msg, channel)
			} else {
				msg := fmt.Sprintf("%s, view log at %s", resultMsg, string(urlBytes[:]))
				bot.SendMessage(msg, channel)
			}
		}()
		return "", nil
	case dumplogCmd.FullCommand():
		logContents, err := ioutil.ReadFile(logFileName)
		if err != nil {
			return "Error reading " + logFileName, err
		}
		index := 0
		if len(logContents) > 1000 {
			index = len(logContents) - 1000
		}
		bot.SendMessage(string(logContents[index:]), channel)
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
