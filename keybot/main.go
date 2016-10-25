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
	"github.com/keybase/slackbot/launchd"
	"gopkg.in/alecthomas/kingpin.v2"
)

func setDarwinEnv(name string, val string) error {
	_, err := slackbot.NewExecCommand("/bin/launchctl", []string{"setenv", name, val}, false, "Set the env").Run("", []string{})
	return err
}

func jobKeybotHandler(channel string, args []string) (string, error) {
	app := kingpin.New("keybot", "Job command parser for keybot")
	app.Terminate(nil)
	stringBuffer := new(bytes.Buffer)
	app.Writer(stringBuffer)

	build := app.Command("build", "Build things")

	buildDarwin := build.Command("darwin", "Start a darwin build")
	clientCommit := buildDarwin.Flag("client-commit", "Build a specific client commit hash").String()
	kbfsCommit := buildDarwin.Flag("kbfs-commit", "Build a specific kbfs commit hash").String()

	cancel := app.Command("cancel", "Cancel")
	cancelCommandArgs := cancel.Flag("command", "Command name").Required().String()

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

	dumplogCmd := app.Command("dumplog", "Dump log for viewing")
	dumplogCommandArgs := dumplogCmd.Flag("command", "Command name").Required().String()

	cmd, usage, cmdErr := cli.Parse(app, args, stringBuffer)
	if usage != "" || cmdErr != nil {
		return usage, cmdErr
	}

	switch cmd {
	case cancel.FullCommand():
		label := labelForCommand(*cancelCommandArgs)
		return launchd.Stop(label)

	case buildDarwin.FullCommand():
		script := launchd.Script{
			Label:      "keybase.build.darwin",
			Path:       "github.com/keybase/client/packaging/prerelease/pull_build.sh",
			Command:    "build darwin",
			BucketName: "prerelease.keybase.io",
			Platform:   "darwin",
			EnvVars: []launchd.EnvVar{
				launchd.EnvVar{Key: "CLIENT_COMMIT", Value: *clientCommit},
				launchd.EnvVar{Key: "KBFS_COMMIT", Value: *kbfsCommit},
			},
		}
		return runScript(launchd.NewEnv(), script)

	case buildAndroid.FullCommand():

		script := launchd.Script{
			Label:      "keybase.build.android",
			Path:       "github.com/keybase/client/packaging/android/build_and_publish.sh",
			Command:    "build android",
			BucketName: "prerelease.keybase.io",
		}
		env := launchd.NewEnv()
		env.GoPath = env.PathFromHome("go-android") // Custom go path for Android so we don't conflict
		return runScript(env, script)

	case buildIOS.FullCommand():
		script := launchd.Script{
			Label:      "keybase.build.ios",
			Path:       "github.com/keybase/client/packaging/ios/build_and_publish.sh",
			Command:    "build ios",
			BucketName: "prerelease.keybase.io",
		}
		env := launchd.NewEnv()
		env.GoPath = env.PathFromHome("go-ios") // Custom go path for iOS so we don't conflict
		return runScript(env, script)

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
		return runScript(launchd.NewEnv(), script)

	case dumplogCmd.FullCommand():
		env := launchd.NewEnv()
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
		return runScript(launchd.NewEnv(), script)

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
		return slackbot.NewExecCommand("/bin/launchctl", []string{"start", "keybase.prerelease.smoketest"}, false, "Start or stop smoketesting a given build").Run("", nil)
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
	bot.AddCommand("restart", slackbot.NewExecCommand("/bin/launchctl", []string{"stop", "keybase.keybot"}, false, "Restart the bot"))

	bot.SetDefault(slackbot.FuncCommand{
		Fn: jobKeybotHandler,
	})
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
