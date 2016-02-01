// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package slackbot

import (
	"testing"
)

func TestHelp(t *testing.T) {
	bot := Bot{}
	bot.commands = make(map[string]Command)

	bot.AddCommand("date", NewExecCommand("/bin/date", nil, true, "Show the current date"))
	bot.AddCommand("utc", NewExecCommand("/bin/date", []string{"-u"}, true, "Show the current date (utc)"))
	msg := bot.helpMessage()
	if msg == "" {
		t.Fatal("No help message")
	}
	t.Logf("Help:\n%s", msg)
}
