// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"strings"
	"testing"

	"github.com/keybase/slackbot"
)

func TestAddBasicCommands(t *testing.T) {
	bot, err := slackbot.NewTestBot()
	if err != nil {
		t.Fatal(err)
	}
	addBasicCommands(bot)
}

func TestBuildDarwin(t *testing.T) {
	bot, err := slackbot.NewTestBot()
	if err != nil {
		t.Fatal(err)
	}
	ext := &darwinbot{}
	out, err := ext.Run(bot, "", []string{"build", "darwin"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "I would have run a launchd job (keybase.build.darwin)" {
		t.Errorf("Unexpected output: %s", out)
	}
}

func TestPromoteRelease(t *testing.T) {
	bot, err := slackbot.NewTestBot()
	if err != nil {
		t.Fatal(err)
	}
	ext := &keybot{}
	out, err := ext.Run(bot, "", []string{"release", "promote", "1.2.3"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "I would have run a launchd job (keybase.release.promote)" {
		t.Errorf("Unexpected output: %s", out)
	}
}

func TestInvalidUsage(t *testing.T) {
	bot, err := slackbot.NewTestBot()
	if err != nil {
		t.Fatal(err)
	}
	ext := &keybot{}
	out, err := ext.Run(bot, "", []string{"release", "oops"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "```\nI don't know what you mean by") {
		t.Errorf("Unexpected output: %s", out)
	}
}
