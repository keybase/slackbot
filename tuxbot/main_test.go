// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"os"
	"strings"
	"testing"

	"github.com/keybase/slackbot"
	"github.com/stretchr/testify/require"
)

type attachmentBackend struct {
	filename string
	title    string
	channel  string
	content  []byte
}

func (*attachmentBackend) SendMessage(string, string) {}

func (b *attachmentBackend) SendAttachment(filename, title, channel string) error {
	//nolint:gosec // The test reads the exact temporary path produced by the code under test.
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	b.filename = filename
	b.title = title
	b.channel = channel
	b.content = content
	return nil
}

func (*attachmentBackend) Listen(slackbot.BotCommandRunner) {}

func TestBuildLinux(t *testing.T) {
	bot, err := slackbot.NewTestBot()
	if err != nil {
		t.Fatal(err)
	}
	ext := &tuxbot{}
	out, err := ext.Run(bot, "", []string{"build", "linux"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "Dry Run: Doing that would run `prerelease.sh`" {
		t.Errorf("Unexpected output: %s", out)
	}
}

func TestInvalidUsage(t *testing.T) {
	bot, err := slackbot.NewTestBot()
	if err != nil {
		t.Fatal(err)
	}
	ext := &tuxbot{}
	out, err := ext.Run(bot, "", []string{"build", "oops"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "```\nI don't know what you mean by") {
		t.Errorf("Unexpected output: %s", out)
	}
}

func TestBuildLinuxSkipCI(t *testing.T) {
	bot, err := slackbot.NewTestBot()
	if err != nil {
		t.Fatal(err)
	}
	ext := &tuxbot{}
	out, err := ext.Run(bot, "", []string{"build", "linux", "--skip-ci"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "Dry Run: Doing that would run `prerelease.sh` with NOWAIT=1 set" {
		t.Errorf("Unexpected output: %s", out)
	}
}

func TestSendFailedBuildOutputAttachment(t *testing.T) {
	backend := &attachmentBackend{}
	bot := slackbot.NewBot(slackbot.NewConfig(false, false), "tuxbot", "", backend)
	ext := &tuxbot{bot: bot}
	journal := []byte(strings.Repeat("full journal output\n", 500))

	require.NoError(t, ext.sendFailedBuildOutput(journal, "conv"))
	require.Equal(t, "failed build output", backend.title)
	require.Equal(t, "conv", backend.channel)
	require.Equal(t, journal, backend.content)
	_, err := os.Stat(backend.filename)
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
}
