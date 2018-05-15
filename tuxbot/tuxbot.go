// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"github.com/keybase/slackbot"
	"github.com/keybase/slackbot/cli"
	"github.com/nlopes/slack"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func (t *tuxbot) linuxBuildFunc(channel string, args []string) (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	t.bot.SendMessage("building linux!!!", channel)
	prereleaseScriptPath := filepath.Join(currentUser.HomeDir, "slackbot/systemd/prerelease.sh")
	prereleaseCmd := exec.Command(prereleaseScriptPath)
	prereleaseCmd.Stdout = os.Stdout
	prereleaseCmd.Stderr = os.Stderr
	err = prereleaseCmd.Run()
	if err != nil {
		journal, _ := exec.Command("journalctl", "--since=today", "--user-unit", "keybase.keybot.service").CombinedOutput()
		api := slack.New(slackbot.GetTokenFromEnv())
		snippetFile := slack.FileUploadParameters{
			Channels: []string{channel},
			Title:    "failed build output",
			Content:  string(journal),
		}
		_, _ = api.UploadFile(snippetFile) // ignore errors here for now
		return "FAILURE", err
	}
	return "SUCCESS", nil
}

type tuxbot struct {
	bot slackbot.Bot
}

func (t *tuxbot) Run(bot slackbot.Bot, channel string, args []string) (string, error) {
	app := kingpin.New("tuxbot", "Command parser for tuxbot")
	app.Terminate(nil)
	stringBuffer := new(bytes.Buffer)
	app.Writer(stringBuffer)

	build := app.Command("build", "Build things")
	buildLinux := build.Command("linux", "Start a linux build")

	cmd, usage, err := cli.Parse(app, args, stringBuffer)
	if usage != "" || err != nil {
		return usage, err
	}

	switch cmd {
	case buildLinux.FullCommand():
		if bot.Config().DryRun() {
			return "Dry Run: Doing that would run `prerelease.sh`", nil
		}
		if bot.Config().Paused() {
			return "I'm paused so I can't do that, but I would have run `prerelease.sh`", nil
		}

		return t.linuxBuildFunc(channel, args)
	}

	return cmd, nil
}

func (t *tuxbot) Help(bot slackbot.Bot) string {
	out, err := t.Run(bot, "", nil)
	if err != nil {
		return fmt.Sprintf("Error getting help: %s", err)
	}
	return out
}
