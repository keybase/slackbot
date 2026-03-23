// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package slackbot

import (
	"testing"

	"github.com/keybase/go-keybase-chat-bot/kbchat/types/chat1"
	"github.com/stretchr/testify/require"
)

func TestHelp(t *testing.T) {
	bot, err := NewTestBot()
	require.NoError(t, err)
	bot.AddCommand("date", NewExecCommand("/bin/date", nil, true, "Show the current date", &config{}))
	bot.AddCommand("utc", NewExecCommand("/bin/date", []string{"-u"}, true, "Show the current date (utc)", &config{}))
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

func TestAdvertisedCommands(t *testing.T) {
	bot, err := NewTestBot()
	require.NoError(t, err)
	bot.AddCommand("date", NewExecCommand("/bin/date", nil, true, "Show the current date", &config{}))
	bot.SetHelp("help body")
	bot.AddAdvertisements(chat1.UserBotCommandInput{
		Name:        "build",
		Description: "Build things",
		Usage:       "!testbot build <target>",
	})

	commands := bot.AdvertisedCommands()
	if len(commands) != 3 {
		t.Fatalf("expected 3 advertised commands, got %d", len(commands))
	}
	if commands[0].Name != "help" {
		t.Fatalf("expected help command first, got %q", commands[0].Name)
	}
	if commands[0].ExtendedDescription == nil || commands[0].ExtendedDescription.DesktopBody != "help body" {
		t.Fatalf("unexpected help extended description: %+v", commands[0].ExtendedDescription)
	}
	if commands[1] != (chat1.UserBotCommandInput{
		Name:        "date",
		Description: "Show the current date",
		Usage:       "!testbot date",
	}) {
		t.Fatalf("unexpected builtin command: %+v", commands[1])
	}
	if commands[2] != (chat1.UserBotCommandInput{
		Name:        "build",
		Description: "Build things",
		Usage:       "!testbot build <target>",
	}) {
		t.Fatalf("unexpected extra command: %+v", commands[2])
	}
}
