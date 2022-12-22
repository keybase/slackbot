// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/keybase/slackbot"
	"github.com/keybase/slackbot/cli"
	"github.com/keybase/slackbot/launchd"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type darwinbot struct{}

func (d *darwinbot) Run(bot *slackbot.Bot, channel string, args []string) (string, error) {
	app := kingpin.New("darwinbot", "Job command parser for darwinbot")
	app.Terminate(nil)
	stringBuffer := new(bytes.Buffer)
	app.Writer(stringBuffer)

	build := app.Command("build", "Build things")

	buildDarwin := build.Command("darwin", "Start a darwin build")
	buildDarwinTest := buildDarwin.Flag("test", "Whether build is for testing").Bool()
	buildDarwinClientCommit := buildDarwin.Flag("client-commit", "Build a specific client commit").String()
	buildDarwinKbfsCommit := buildDarwin.Flag("kbfs-commit", "Build a specific kbfs commit").String()
	buildDarwinNoPull := buildDarwin.Flag("skip-pull", "Don't pull before building the app").Bool()
	buildDarwinSkipCI := buildDarwin.Flag("skip-ci", "Whether to skip CI").Bool()
	buildDarwinSmoke := buildDarwin.Flag("smoke", "Whether to make a pair of builds for smoketesting when on a branch").Bool()
	buildDarwinNoS3 := buildDarwin.Flag("skip-s3", "Don't push to S3 after building the app").Bool()
	buildDarwinNoNotarize := buildDarwin.Flag("skip-notarize", "Don't notarize the app").Bool()

	cancel := app.Command("cancel", "Cancel")
	cancelLabel := cancel.Arg("label", "Launchd job label").Required().String()

	dumplogCmd := app.Command("dumplog", "Show the log file")
	dumplogCommandLabel := dumplogCmd.Arg("label", "Launchd job label").Required().String()
	gitDiffCmd := app.Command("gdiff", "Show the git diff")
	gitDiffRepo := gitDiffCmd.Arg("repo", "Repo path relative to $GOPATH/src").Required().String()

	gitCleanCmd := app.Command("gclean", "Clean the repo")

	upgrade := app.Command("upgrade", "Upgrade package")
	upgradePackageName := upgrade.Arg("name", "Package name (yarn, go, fastlane, etc)").Required().String()

	cmd, usage, cmdErr := cli.Parse(app, args, stringBuffer)
	if usage != "" || cmdErr != nil {
		return usage, cmdErr
	}

	if bot.Config().DryRun() {
		return fmt.Sprintf("I would have run: `%#v`", cmd), nil
	}

	home := os.Getenv("HOME")
	path := "/sbin:/usr/sbin:/bin:/usr/local/bin:/usr/bin"
	env := launchd.NewEnv(home, path)
	switch cmd {
	case cancel.FullCommand():
		return launchd.Stop(*cancelLabel)

	case buildDarwin.FullCommand():
		smokeTest := true
		skipCI := *buildDarwinSkipCI
		testBuild := *buildDarwinTest
		// If it's a custom build, make it a test build unless --smoke is passed.
		if *buildDarwinClientCommit != "" || *buildDarwinKbfsCommit != "" {
			smokeTest = *buildDarwinSmoke
			testBuild = !*buildDarwinSmoke
		}
		script := launchd.Script{
			Label:      "keybase.build.darwin",
			Path:       "github.com/keybase/client/packaging/build_darwin.sh",
			BucketName: "prerelease.keybase.io",
			Platform:   "darwin",
			EnvVars: []launchd.EnvVar{
				{Key: "SMOKE_TEST", Value: boolToEnvString(smokeTest)},
				{Key: "TEST", Value: boolToEnvString(testBuild)},
				{Key: "CLIENT_COMMIT", Value: *buildDarwinClientCommit},
				{Key: "KBFS_COMMIT", Value: *buildDarwinKbfsCommit},
				// TODO: Rename to SKIP_CI in packaging scripts
				{Key: "NOWAIT", Value: boolToEnvString(skipCI)},
				{Key: "NOPULL", Value: boolToEnvString(*buildDarwinNoPull)},
				{Key: "NOS3", Value: boolToEnvString(*buildDarwinNoS3)},
				{Key: "NONOTARIZE", Value: boolToEnvString(*buildDarwinNoNotarize)},
			},
		}
		return runScript(bot, channel, env, script)
	case dumplogCmd.FullCommand():
		readPath, err := env.LogPathForLaunchdLabel(*dumplogCommandLabel)
		if err != nil {
			return "", err
		}
		script := launchd.Script{
			Label:      "keybase.dumplog",
			Path:       "github.com/keybase/slackbot/scripts/dumplog.sh",
			BucketName: "prerelease.keybase.io",
			EnvVars: []launchd.EnvVar{
				{Key: "READ_PATH", Value: readPath},
				{Key: "NOLOG", Value: boolToEnvString(true)},
			},
		}
		return runScript(bot, channel, env, script)
	case gitDiffCmd.FullCommand():
		rawRepoText := *gitDiffRepo
		repoParsed := strings.Split(strings.Trim(rawRepoText, "`<>"), "|")[1]

		script := launchd.Script{
			Label:      "keybase.gitdiff",
			Path:       "github.com/keybase/slackbot/scripts/run_and_send_stdout.sh",
			BucketName: "prerelease.keybase.io",
			EnvVars: []launchd.EnvVar{
				{Key: "REPO", Value: repoParsed},
				{Key: "PREFIX_GOPATH", Value: boolToEnvString(true)},
				{Key: "SCRIPT_TO_RUN", Value: "./git_diff.sh"},
			},
		}
		return runScript(bot, channel, env, script)
	case gitCleanCmd.FullCommand():
		script := launchd.Script{
			Label:      "keybase.gitclean",
			Path:       "github.com/keybase/slackbot/scripts/run_and_send_stdout.sh",
			BucketName: "prerelease.keybase.io",
			EnvVars: []launchd.EnvVar{
				{Key: "SCRIPT_TO_RUN", Value: "./git_clean.sh"},
			},
		}
		return runScript(bot, channel, env, script)

	case upgrade.FullCommand():
		script := launchd.Script{
			Label: "keybase.upgrade",
			Path:  "github.com/keybase/slackbot/scripts/upgrade.sh",
			EnvVars: []launchd.EnvVar{
				{Key: "NAME", Value: *upgradePackageName},
			},
		}
		return runScript(bot, channel, env, script)
	}
	return cmd, nil
}

func (d *darwinbot) Help(bot *slackbot.Bot) string {
	out, err := d.Run(bot, "", nil)
	if err != nil {
		return fmt.Sprintf("Error getting help: %s", err)
	}
	return out
}
