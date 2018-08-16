// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"strings"
	"testing"

	"github.com/keybase/slackbot"
)

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
