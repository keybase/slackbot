// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/keybase/slackbot"
	"github.com/keybase/slackbot/cli"
	"github.com/keybase/slackbot/launchd"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type darwinbot struct{}

func (d *darwinbot) Run(bot slackbot.Bot, channel string, args []string) (string, error) {
	app := kingpin.New("darwinbot", "Job command parser for darwinbot")
	app.Terminate(nil)
	stringBuffer := new(bytes.Buffer)
	app.Writer(stringBuffer)

	build := app.Command("build", "Build things")

	buildDarwin := build.Command("darwin", "Start a darwin build")
	buildDarwinTest := buildDarwin.Flag("test", "Whether build is for testing").Bool()
	buildDarwinCientCommit := buildDarwin.Flag("client-commit", "Build a specific client commit").String()
	buildDarwinKbfsCommit := buildDarwin.Flag("kbfs-commit", "Build a specific kbfs commit").String()
	buildDarwinSkipCI := buildDarwin.Flag("skip-ci", "Whether to skip CI").Bool()

	cancel := app.Command("cancel", "Cancel")
	cancelLabel := cancel.Arg("label", "Launchd job label").Required().String()

	cmd, usage, cmdErr := cli.Parse(app, args, stringBuffer)
	if usage != "" || cmdErr != nil {
		return usage, cmdErr
	}

	if bot.Config().DryRun() {
		return fmt.Sprintf("I would have run: `%#v`", cmd), nil
	}

	home := os.Getenv("HOME")
	shims := filepath.Join(home, ".rbenv/shims")
	path := "/sbin:/usr/sbin:/bin:/usr/bin:/usr/local/bin:" + shims
	env := launchd.NewEnv(home, path)
	switch cmd {
	case cancel.FullCommand():
		return launchd.Stop(*cancelLabel)

	case buildDarwin.FullCommand():
		script := launchd.Script{
			Label:      "keybase.build.darwin",
			Path:       "github.com/keybase/client/packaging/prerelease/pull_build.sh",
			Command:    "build darwin",
			BucketName: "prerelease.keybase.io",
			Platform:   "darwin",
			EnvVars: []launchd.EnvVar{
				launchd.EnvVar{Key: "SMOKE_TEST", Value: boolToEnvString(true)},
				launchd.EnvVar{Key: "TEST", Value: boolToEnvString(*buildDarwinTest)},
				launchd.EnvVar{Key: "CLIENT_COMMIT", Value: *buildDarwinCientCommit},
				launchd.EnvVar{Key: "KBFS_COMMIT", Value: *buildDarwinKbfsCommit},
				// TODO: Rename to SKIP_CI in packaging scripts
				launchd.EnvVar{Key: "NOWAIT", Value: boolToEnvString(*buildDarwinSkipCI)},
			},
		}
		return runScript(bot, channel, env, script)
	}
	return cmd, nil
}

func (d *darwinbot) Help(bot slackbot.Bot) string {
	out, err := d.Run(bot, "", nil)
	if err != nil {
		return fmt.Sprintf("Error getting help: %s", err)
	}
	return out
}
