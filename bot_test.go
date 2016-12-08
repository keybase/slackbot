// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package slackbot

import (
	"testing"
)

func TestHelp(t *testing.T) {
	bot, err := NewTestBot()
	if err != nil {
		t.Fatal(err)
	}
	bot.AddCommand("date", NewExecCommand("/bin/date", nil, true, "Show the current date"))
	bot.AddCommand("utc", NewExecCommand("/bin/date", []string{"-u"}, true, "Show the current date (utc)"))
	msg := bot.HelpMessage()
	if msg == "" {
		t.Fatal("No help message")
	}
	t.Logf("Help:\n%s", msg)
}

func TestParseInput(t *testing.T) {
	args := parseInput(`!keybot dumplog "release promote"`)
	if args[0] != "!keybot" || args[1] != "dumplog" || args[2] != `release promote` {
		t.Fatal("Invalid parse")
	}
}
