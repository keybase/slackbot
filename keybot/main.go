// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"bytes"
	"log"
	"strconv"
	"strings"

	"github.com/keybase/slackbot"
	"github.com/keybase/slackbot/cli"
	"github.com/keybase/slackbot/jenkins"
	"github.com/keybase/slackbot/launchd"
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
	testClientCommit := test.Flag("client-commit", "Test a specific client commit hash").String()

	buildDarwin := build.Command("darwin", "Start a darwin build")
	testDarwin := test.Command("darwin", "Start a darwin test build")
	cancelDarwin := cancel.Command("darwin", "Cancel the darwin build")

	buildAndroid := build.Command("android", "Start an android build")
	buildIOS := build.Command("ios", "Start an ios build")

	release := app.Command("release", "Release things")
	releasePromote := release.Command("promote", "Promote a release to public")
	releaseToPromote := releasePromote.Arg("release-to-promote", "Promote a specific release to public immediately").String()
	releaseBroken := release.Command("broken", "Mark a release as broken")
	releaseBrokenVersion := releaseBroken.Arg("version", "Mark a release as broken").Required().String()

	smoketestBuild := app.Command("smoketest", "Set the smoke testing status of a build")
	smoketestBuildA := smoketestBuild.Flag("build-a", "The first of the two IDs comprising the new build").Required().String()
	smoketestBuildPlatform := smoketestBuild.Flag("platform", "The build's platform (darwin, linux, windows)").Required().String()
	smoketestBuildEnable := smoketestBuild.Flag("enable", "Whether smoketesting should be enabled").Required().Bool()
	smoketestBuildMaxTesters := smoketestBuild.Flag("max-testers", "Max number of testers for this build").Required().Int()

	buildWindows := build.Command("windows", "Start a windows build")
	testWindows := test.Command("windows", "Start a windows test build")
	cancelWindows := cancel.Command("windows", "Cancel last windows build")
	cancelWindowsQueueID := cancelWindows.Arg("quid", "Queue id of build to stop").Required().String()

	dumplogCmd := app.Command("dumplog", "Dump log for viewing")
	dumplogCommandArgs := dumplogCmd.Flag("command", "Command name").Required().String()

	cmd, usage, cmdErr := cli.Parse(app, args, stringBuffer)
	if usage != "" || cmdErr != nil {
		return usage, cmdErr
	}

	env := launchd.NewEnv()

	if setErr := setDarwinEnv("CLIENT_COMMIT", *clientCommit); setErr != nil {
		return "", setErr
	}
	if setErr := setDarwinEnv("KBFS_COMMIT", *kbfsCommit); setErr != nil {
		return "", setErr
	}
	if *testClientCommit != "" {
		if setErr := setDarwinEnv("CLIENT_COMMIT", *testClientCommit); setErr != nil {
			return "", setErr
		}
	}

	emptyArgs := []string{}
	switch cmd {
	// Darwin
	case buildDarwin.FullCommand():
		return slackbot.NewExecCommand("/bin/launchctl", []string{"start", "keybase.darwin"}, false, "Perform a darwin build").Run("", emptyArgs)
	case testDarwin.FullCommand():
		return slackbot.NewExecCommand("/bin/launchctl", []string{"start", "keybase.darwin.test"}, false, "Test the darwin build").Run("", emptyArgs)
	case cancelDarwin.FullCommand():
		return slackbot.NewExecCommand("/bin/launchctl", []string{"stop", "keybase.darwin"}, false, "Cancel a running darwin build").Run("", emptyArgs)

	// Windows
	case buildWindows.FullCommand():
		return jenkins.StartBuild(*clientCommit, *kbfsCommit, "")
	case testWindows.FullCommand():
		return jenkins.StartBuild(*clientCommit, *kbfsCommit, "update-windows-prod-test-v2.json")
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
		script := launchd.Script{
			Label:      "keybase.release.promote",
			Path:       "github.com/keybase/slackbot/scripts/release.promote.sh",
			Command:    "release promote",
			BucketName: "prerelease.keybase.io",
			Platform:   "darwin",
			EnvVars: []launchd.EnvVar{
				launchd.EnvVar{Key: "RELEASE_TO_PROMOTE", Value: *releaseToPromote},
			},
		}
		return runScript(env, script)

	case dumplogCmd.FullCommand():
		readPath, err := env.LogPath(labelForCommand(*dumplogCommandArgs))
		if err != nil {
			return "", err
		}
		script := launchd.Script{
			Label:      "keybase.dumplog",
			Path:       "github.com/keybase/slackbot/scripts/dumplog.sh",
			Command:    "dumplog",
			BucketName: "prerelease.keybase.io",
			EnvVars: []launchd.EnvVar{
				launchd.EnvVar{Key: "READ_PATH", Value: readPath},
			},
		}
		return runScript(env, script)

	case releaseBroken.FullCommand():
		script := launchd.Script{
			Label:      "keybase.release.broken",
			Path:       "github.com/keybase/slackbot/scripts/release.broken.sh",
			Command:    "release broken",
			BucketName: "prerelease.keybase.io",
			Platform:   "darwin",
			EnvVars: []launchd.EnvVar{
				launchd.EnvVar{Key: "BROKEN_RELEASE", Value: *releaseBrokenVersion},
			},
		}
		return runScript(env, script)

	case smoketestBuild.FullCommand():
		if err := setDarwinEnv("SMOKETEST_BUILD_A", *smoketestBuildA); err != nil {
			return "", err
		}
		if err := setDarwinEnv("SMOKETEST_PLATFORM", *smoketestBuildPlatform); err != nil {
			return "", err
		}
		if err := setDarwinEnv("SMOKETEST_MAX_TESTERS", strconv.Itoa(*smoketestBuildMaxTesters)); err != nil {
			return "", err
		}
		buildEnable := "true"
		if !*smoketestBuildEnable {
			buildEnable = "false"
		}
		if err := setDarwinEnv("SMOKETEST_BUILD_ENABLE", buildEnable); err != nil {
			return "", err
		}
		return slackbot.NewExecCommand("/bin/launchctl", []string{"start", "keybase.prerelease.smoketest"}, false, "Start or stop smoketesting a given build").Run("", emptyArgs)
	}
	return cmd, nil
}

func labelForCommand(cmd string) string {
	return "keybase." + strings.Replace(cmd, " ", ".", -1)
}

func runScript(env launchd.Env, script launchd.Script) (string, error) {
	path, err := env.WritePlist(script)
	if err != nil {
		return "", err
	}
	defer env.Cleanup(script)
	return launchd.NewStartCommand(path, script.Label).Run("", nil)
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

	bot.AddCommand("smoketest", slackbot.FuncCommand{
		Desc: "Smoketest all the things!",
		Fn:   kingpinKeybotHandler,
	})

	bot.AddCommand("dumplog", slackbot.FuncCommand{
		Desc: "Access logs",
		Fn:   kingpinKeybotHandler,
	})

	bot.AddCommand("restart", slackbot.NewExecCommand("/bin/launchctl", []string{"stop", "keybase.keybot"}, false, "Restart the bot"))
}

func main() {
	bot, err := slackbot.NewBot(slackbot.GetTokenFromEnv())
	if err != nil {
		log.Fatal(err)
	}

	addCommands(bot)

	log.Println("Started keybot")
	bot.Listen()
}
