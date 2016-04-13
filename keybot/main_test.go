// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
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

func TestKinpinHandler(t *testing.T) {
	out, err := kingpinHandler([]string{"build", "darwin"})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Output: %s", out)

	out2, err2 := kingpinHandler([]string{"build"})
	if err2 != nil {
		t.Fatal(err2)
	}
	t.Logf("Usage: %s", out2)
}
