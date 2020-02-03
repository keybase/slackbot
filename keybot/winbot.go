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
	"strings"
	"sync"
	"time"

	"github.com/keybase/slackbot"
	"github.com/keybase/slackbot/cli"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type winbot struct {
	testAuto chan struct{}
	stopAuto chan struct{}
}

const numLogLines = 10

// Keep track of the current build process, protected with a mutex,
// to support cancellation
var buildProcessMutex sync.Mutex
var buildProcess *os.Process

func (d *winbot) Run(bot *slackbot.Bot, channel string, args []string) (string, error) {
	app := kingpin.New("winbot", "Job command parser for winbot")
	app.Terminate(nil)
	stringBuffer := new(bytes.Buffer)
	app.Writer(stringBuffer)

	buildWindows := app.Command("build", "Start a windows build")
	buildWindowsTest := buildWindows.Flag("test", "Test build, skips admin/test channel").Bool()
	buildWindowsCientCommit := buildWindows.Flag("client-commit", "Build a specific client commit").String()
	buildWindowsKbfsCommit := buildWindows.Flag("kbfs-commit", "Build a specific kbfs commit").String()
	buildWindowsUpdaterCommit := buildWindows.Flag("updater-commit", "Build a specific updater commit").String()
	buildWindowsSkipCI := buildWindows.Flag("skip-ci", "Whether to skip CI").Bool()
	buildWindowsSmoke := buildWindows.Flag("smoke", "Build a smoke pair").Bool()
	buildWindowsAuto := buildWindows.Flag("automated", "Specify build was triggered automatically").Hidden().Bool()

	cancel := app.Command("cancel", "Cancel current")

	dumplogCmd := app.Command("dumplog", "Show the last log file")
	gitDiffCmd := app.Command("gdiff", "Show the git diff")
	gitDiffRepo := gitDiffCmd.Arg("repo", "Repo path relative to $GOPATH/src").Required().String()

	gitCleanCmd := app.Command("gclean", "Clean the repo")
	gitCleanRepo := gitCleanCmd.Arg("repo", "Repo path relative to $GOPATH/src").Required().String()

	logFileName := path.Join(os.TempDir(), "keybase.build.windows.log")

	testAutoBuild := app.Command("testauto", "Simulate an automated daily build").Hidden()
	startAutoTimer := app.Command("startAutoTimer", "Start the auto build timer")
	startAutoTimerInterval := startAutoTimer.Flag("interval", "Number of hours between auto builds, 0  to stop").Default("24").Int()
	startAutoTimerStartHour := startAutoTimer.Flag("startHour", "Number of hours after midnight to build, local time").Default("7").Int()
	startAutoTimerDelay := startAutoTimer.Flag("delay", "Number of hours to wait before starting auto timer").Default("0").Int()

	restartCmd := app.Command("restart", "Quit and let calling script invoke bot again")

	cmd, usage, cmdErr := cli.Parse(app, args, stringBuffer)
	if usage != "" || cmdErr != nil {
		return usage, cmdErr
	}

	// do these regardless of dry run status
	if cmd == testAutoBuild.FullCommand() {
		d.testAuto <- struct{}{}
		return "Sent test signal", nil
	}

	if cmd == startAutoTimer.FullCommand() {
		if d.stopAuto != nil {
			d.stopAuto <- struct{}{}
		}
		if *startAutoTimerInterval > 0 {
			go d.winAutoBuild(bot, channel, *startAutoTimerInterval, *startAutoTimerDelay, *startAutoTimerStartHour)
		}
		return "", nil
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
		skipTestChannel := *buildWindowsTest
		var autoBuild string

		if bot.Config().DryRun() {
			return fmt.Sprintf("I would have done a build"), nil
		}

		if bot.Config().Paused() {
			return fmt.Sprintf("I'm paused so I can't do that, but I would have done a build"), nil
		}

		if *buildWindowsAuto {
			autoBuild = "Automatic Build: "
		}

		// Test channel tells the scripts this is an admin build
		updateChannel := "Test"
		if skipTestChannel {
			if smokeTest {
				return "Test and Smoke are exclusive options", nil
			}
			updateChannel = "None"
		} else if smokeTest {
			updateChannel = "Smoke"
			if !skipCI {
				updateChannel = "SmokeCI"
			}
		}

		msg := fmt.Sprintf(autoBuild+"I'm starting the job `windows build`. To cancel run `!%s cancel`. ", bot.Name())
		msg = fmt.Sprintf(msg+"updateChannel is %s, smokeTest is %v", updateChannel, smokeTest)
		bot.SendMessage(msg, channel)

		os.Remove(logFileName)
		logf, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return "Unable to open logfile", err
		}

		gitCmd := exec.Command(
			"git.exe",
			"checkout",
			"master",
		)
		gitCmd.Dir = os.ExpandEnv("$GOPATH/src/github.com/keybase/client")
		stdoutStderr, err := gitCmd.CombinedOutput()
		logf.Write(stdoutStderr)
		if err != nil {
			logf.WriteString(gitCmd.Dir)
			logf.Close()
			return string(stdoutStderr), err
		}

		gitCmd = exec.Command(
			"git.exe",
			"pull",
		)
		gitCmd.Dir = os.ExpandEnv("$GOPATH/src/github.com/keybase/client")
		stdoutStderr, err = gitCmd.CombinedOutput()
		logf.Write(stdoutStderr)
		if err != nil {
			logf.WriteString(gitCmd.Dir)
			logf.Close()
			return string(stdoutStderr), err
		}

		if buildWindowsCientCommit != nil && *buildWindowsCientCommit != "" && *buildWindowsCientCommit != "master" {
			msg := fmt.Sprintf(autoBuild+"I'm trying to use commit %s", *buildWindowsCientCommit)
			bot.SendMessage(msg, channel)

			gitCmd = exec.Command(
				"git.exe",
				"checkout",
				*buildWindowsCientCommit,
			)
			gitCmd.Dir = os.ExpandEnv("$GOPATH/src/github.com/keybase/client")
			stdoutStderr, err = gitCmd.CombinedOutput()
			logf.Write(stdoutStderr)

			if err != nil {
				logf.WriteString(fmt.Sprintf("error doing git pull in %s\n", gitCmd.Dir))
				logf.Close()
				return string(stdoutStderr), err
			}

			// Test if we're on a branch. If so, do git pull once more.
			gitCmd = exec.Command(
				"git.exe",
				"rev-parse",
				"--abbrev-ref",
				"HEAD",
			)
			gitCmd.Dir = os.ExpandEnv("$GOPATH/src/github.com/keybase/client")
			stdoutStderr, err = gitCmd.CombinedOutput()
			if err != nil {
				logf.WriteString(fmt.Sprintf("error going git rev-parse dir: %s\n", gitCmd.Dir))
				logf.Close()
				return string(stdoutStderr), err
			}
			commit := strings.TrimSpace(string(stdoutStderr[:]))
			if commit != "HEAD" {
				gitCmd = exec.Command(
					"git.exe",
					"pull",
				)
				gitCmd.Dir = os.ExpandEnv("$GOPATH/src/github.com/keybase/client")
				stdoutStderr, err = gitCmd.CombinedOutput()
				logf.Write(stdoutStderr)
				if err != nil {
					logf.WriteString(fmt.Sprintf("error doing git pull on %s in %s\n", commit, gitCmd.Dir))
					logf.Close()
					return string(stdoutStderr), err
				}
			}
		}

		gitCmd = exec.Command(
			"git.exe",
			"rev-parse",
			"HEAD",
		)
		gitCmd.Dir = os.ExpandEnv("$GOPATH/src/github.com/keybase/client")
		stdoutStderr, err = gitCmd.CombinedOutput()
		if err != nil {
			logf.WriteString(fmt.Sprintf("error getting current commit for logs: %s", gitCmd.Dir))
			logf.Close()
			return string(stdoutStderr), err
		}
		logf.WriteString(fmt.Sprintf("HEAD is currently at %s\n", string(stdoutStderr)))

		cmd := exec.Command(
			"cmd", "/c",
			path.Join(os.Getenv("GOPATH"), "src/github.com/keybase/client/packaging/windows/dorelease.cmd"),
			">>",
			logFileName,
			"2>&1")
		cmd.Env = append(os.Environ(),
			"ClientRevision="+*buildWindowsCientCommit,
			"KbfsRevision="+*buildWindowsKbfsCommit,
			"UpdaterRevision="+*buildWindowsUpdaterCommit,
			"UpdateChannel="+updateChannel,
			"SlackBot=1",
		)
		logf.WriteString(fmt.Sprintf("cmd: %+v\n", cmd))
		logf.Close()

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
				"--maxsize=5000000",
				"--bucket-name="+bucketName,
				"--path="+logFileName,
			)
			resultMsg := autoBuild + "Finished the job `windows build`"
			if err != nil {
				resultMsg = autoBuild + "Error in job `windows build`"
				var lines [numLogLines]string
				// Send a log snippet too
				index := 0
				lineCount := 0

				f, err := os.Open(logFileName)
				if err != nil {
					bot.SendMessage(autoBuild+"Error reading "+logFileName+": "+err.Error(), channel)
				}

				scanner := bufio.NewScanner(f)
				for scanner.Scan() {
					lines[lineCount%numLogLines] = scanner.Text()
					lineCount += 1
				}
				if err := scanner.Err(); err != nil {
					bot.SendMessage(autoBuild+"Error scanning "+logFileName+": "+err.Error(), channel)
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

	case gitDiffCmd.FullCommand():
		rawRepoText := *gitDiffRepo
		repoParsed := strings.Split(strings.Trim(rawRepoText, "`<>"), "|")[1]

		gitDiffCmd := exec.Command(
			"git.exe",
			"diff",
		)
		gitDiffCmd.Dir = os.ExpandEnv(path.Join("$GOPATH/src", repoParsed))

		if exists, err := Exists(path.Join(gitDiffCmd.Dir, ".git")); !exists {
			return "Not a git repo", err
		}

		stdoutStderr, err := gitDiffCmd.CombinedOutput()
		if err != nil {
			return "Error", err
		}
		bot.SendMessage(string(stdoutStderr), channel)

	case gitCleanCmd.FullCommand():
		rawRepoText := *gitCleanRepo
		repoParsed := strings.Split(strings.Trim(rawRepoText, "`<>"), "|")[1]

		gitCleanCmd := exec.Command(
			"git.exe",
			"clean",
			"-f",
		)
		gitCleanCmd.Dir = os.ExpandEnv(path.Join("$GOPATH/src", repoParsed))

		if exists, err := Exists(path.Join(gitCleanCmd.Dir, ".git")); !exists {
			return "Not a git repo", err
		}

		stdoutStderr, err := gitCleanCmd.CombinedOutput()
		if err != nil {
			return "Error", err
		}

		bot.SendMessage(string(stdoutStderr), channel)

	case restartCmd.FullCommand():
		os.Exit(0)
	}
	return cmd, nil
}

func (d *winbot) Help(bot *slackbot.Bot) string {
	out, err := d.Run(bot, "", nil)
	if err != nil {
		return fmt.Sprintf("Error getting help: %s", err)
	}
	return out
}

func Exists(name string) (bool, error) {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err != nil, err
}

func (d *winbot) winAutoBuild(bot *slackbot.Bot, channel string, interval int, delay int, startHour int) {
	d.testAuto = make(chan struct{})
	d.stopAuto = make(chan struct{})
	for {
		hour := time.Now().Hour() + delay
		if delay > 0 {
			delay = 0
		} else {
			hour = ((interval - hour) + startHour)
		}
		next := time.Now().Add(time.Hour * time.Duration(hour))
		for next.Weekday() == time.Saturday || next.Weekday() == time.Sunday {
			hour += interval
			next = time.Now().Add(time.Hour * time.Duration(hour))
		}

		msg := fmt.Sprintf("Next automatic build at %s", next.Format(time.RFC822))
		bot.SendMessage(msg, channel)

		args := []string{"build", "--automated"}

		select {
		case <-d.testAuto:
		case <-time.After(time.Duration(hour) * time.Hour):
			args = append(args, "--smoke")
		case <-d.stopAuto:
			return
		}
		message, err := d.Run(bot, channel, args)
		if err != nil {
			msg := fmt.Sprintf("AutoBuild ERROR -- %s: %s", message, err.Error())
			bot.SendMessage(msg, channel)
		}
	}
}
