// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package slackbot

import (
	"errors"
	"testing"
	"time"

	"github.com/keybase/go-keybase-chat-bot/kbchat/types/chat1"
	"github.com/stretchr/testify/require"
)

type fakeBackend struct {
	messages    []string
	attachments []string
	listenErr   error
}

func (b *fakeBackend) SendMessage(text string, _ string) {
	b.messages = append(b.messages, text)
}

func (b *fakeBackend) SendAttachment(filename, _ string, _ string) error {
	b.attachments = append(b.attachments, filename)
	return nil
}

func (b *fakeBackend) Listen(BotCommandRunner) error {
	return b.listenErr
}

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

func TestPausedBlocksCommand(t *testing.T) {
	backend := &fakeBackend{}
	bot := NewBot(NewConfig(false, true), "testbot", "", backend)
	ran := false
	bot.AddCommand("ping", NewFuncCommand(func(_ string, _ []string) (string, error) {
		ran = true
		return "pong", nil
	}, "Ping", bot.Config()))

	err := bot.RunCommand([]string{"ping"}, "chan")
	require.NoError(t, err)
	require.False(t, ran)
	require.Equal(t, []string{"I can't do that, I'm paused."}, backend.messages)
}

func TestIgnorePauseRunsCommandWhilePaused(t *testing.T) {
	backend := &fakeBackend{}
	bot := NewBot(NewConfig(false, true), "testbot", "", backend)
	ranCh := make(chan struct{})
	bot.AddCommand("ping", NewFuncCommand(func(_ string, _ []string) (string, error) {
		close(ranCh)
		return "pong", nil
	}, "Ping", bot.Config()))

	err := bot.RunCommand([]string{"ping", "--ignore-pause"}, "chan")
	require.NoError(t, err)
	select {
	case <-ranCh:
	case <-time.After(5 * time.Second):
		t.Fatal("command did not run with --ignore-pause")
	}
	require.True(t, bot.Config().Paused(), "bot should remain paused")
}

func TestCommandPanicIsContained(t *testing.T) {
	backend := &fakeBackend{}
	bot := NewBot(NewConfig(false, false), "testbot", "", backend)
	command := NewFuncCommand(func(_ string, _ []string) (string, error) {
		panic("boom")
	}, "Panic", bot.Config())

	bot.run([]string{"panic"}, command, "chan")
	require.Equal(t, []string{`Oops, there was an internal error running "panic".`}, backend.messages)
}

func TestSendAttachment(t *testing.T) {
	backend := &fakeBackend{}
	bot := NewBot(NewConfig(false, false), "testbot", "", backend)

	require.NoError(t, bot.SendAttachment("/tmp/build.log", "build output", "chan"))
	require.Equal(t, []string{"/tmp/build.log"}, backend.attachments)
}

func TestListenReturnsBackendError(t *testing.T) {
	expectedErr := errors.New("listen failed")
	backend := &fakeBackend{listenErr: expectedErr}
	bot := NewBot(NewConfig(false, false), "testbot", "", backend)

	require.ErrorIs(t, bot.Listen(), expectedErr)
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
