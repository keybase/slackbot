// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"strings"
	"testing"

	"github.com/keybase/slackbot"
)

func TestAddCommands(t *testing.T) {
	bot, err := slackbot.NewTestBot()
	if err != nil {
		t.Fatal(err)
	}
	addCommands(bot)
}

func TestBuildDarwin(t *testing.T) {
	out, err := kingpinKeybotHandler("", []string{"build", "darwin"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "Dry Run: Doing that would run `/bin/launchctl` with args: [start keybase.darwin]" {
		t.Errorf("Unexpected output: %s", out)
	}
}

func TestPromoteRelease(t *testing.T) {
	out, err := kingpinKeybotHandler("", []string{"release", "promote", "1.2.3"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "I would have run a launchd job (keybase.release.promote)" {
		t.Errorf("Unexpected output: %s", out)
	}
}

func TestInvalidUsage(t *testing.T) {
	out, err := kingpinKeybotHandler("", []string{"release", "oops"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "```\nI don't know what you mean by") {
		t.Errorf("Unexpected output: %s", out)
	}
}
