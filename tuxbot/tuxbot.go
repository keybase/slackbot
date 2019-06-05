// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"github.com/keybase/slackbot"
	"github.com/keybase/slackbot/cli"
	"github.com/nlopes/slack"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func (t *tuxbot) linuxBuildFunc(channel string, args []string, skipCI bool, nightly bool) (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	t.bot.SendMessage("building linux!!!", channel)
	prereleaseScriptPath := filepath.Join(currentUser.HomeDir, "slackbot/systemd/prerelease.sh")
	prereleaseCmd := exec.Command(prereleaseScriptPath)
	prereleaseCmd.Stdout = os.Stdout
	prereleaseCmd.Stderr = os.Stderr
	prereleaseCmd.Env = os.Environ()
	if skipCI {
		prereleaseCmd.Env = append(prereleaseCmd.Env, "NOWAIT=1")
		t.bot.SendMessage("--- with NOWAIT=1", channel)
	}
	if nightly {
		prereleaseCmd.Env = append(prereleaseCmd.Env, "KEYBASE_NIGHTLY=1")
		t.bot.SendMessage("--- with KEYBASE_NIGHTLY=1", channel)
	}
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
	bot *slackbot.Bot
}

func (t *tuxbot) Run(bot *slackbot.Bot, channel string, args []string) (string, error) {
	app := kingpin.New("tuxbot", "Command parser for tuxbot")
	app.Terminate(nil)
	stringBuffer := new(bytes.Buffer)
	app.Writer(stringBuffer)

	build := app.Command("build", "Build things")
	buildLinux := build.Command("linux", "Start a linux build")
	buildLinuxSkipCI := buildLinux.Flag("skip-ci", "Whether to skip CI").Bool()
	buildLinuxNightly := buildLinux.Flag("nightly", "Trigger a nightly build instead of main channel").Bool()

	cmd, usage, err := cli.Parse(app, args, stringBuffer)
	if usage != "" || err != nil {
		return usage, err
	}

	switch cmd {
	case buildLinux.FullCommand():
		if bot.Config().DryRun() {
			if *buildLinuxSkipCI {
				return "Dry Run: Doing that would run `prerelease.sh` with NOWAIT=1 set", nil
			}
			return "Dry Run: Doing that would run `prerelease.sh`", nil
		}
		if bot.Config().Paused() {
			return "I'm paused so I can't do that, but I would have run `prerelease.sh`", nil
		}

		ret, err := t.linuxBuildFunc(channel, args, *buildLinuxSkipCI, *buildLinuxNightly)

		var stathatErr error
		if err == nil {
			stathatErr = postStathat("tuxbot - nightly - success", "1")
		} else {
			stathatErr = postStathat("tuxbot - nightly - failure", "1")
		}
		if stathatErr != nil {
			return fmt.Sprintf("stathat error. original message: %s", ret),
				fmt.Errorf("stathat error: %s. original error: %s", stathatErr, err)
		}

		return ret, err
	}

	return cmd, nil
}

func postStathat(key string, count string) error {
	ezkey := os.Getenv("STATHAT_EZKEY")
	if ezkey == "" {
		return fmt.Errorf("no stathat key")
	}
	vals := url.Values{
		"ezkey": {ezkey},
		"stat":  {key},
		"count": {count},
	}
	_, err := http.PostForm("https://api.stathat.com/ez", vals)
	return err
}

func (t *tuxbot) Help(bot *slackbot.Bot) string {
	out, err := t.Run(bot, "", nil)
	if err != nil {
		return fmt.Sprintf("Error getting help: %s", err)
	}
	return out
}
