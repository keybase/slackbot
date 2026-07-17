// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/keybase/slackbot"
)

func TestReadLastBuiltCommit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lastbuild")
	if commit := readLastBuiltCommit(path); commit != "" {
		t.Errorf("Expected empty commit for missing file, got %q", commit)
	}
	// run.sh writes the commit with a trailing newline.
	if err := os.WriteFile(path, []byte("abc123\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if commit := readLastBuiltCommit(path); commit != "abc123" {
		t.Errorf("Expected abc123, got %q", commit)
	}
}

func TestCurrentClientCommitExplicit(t *testing.T) {
	commit, err := currentClientCommit("abc123")
	if err != nil {
		t.Fatal(err)
	}
	if commit != "abc123" {
		t.Errorf("Expected explicit commit to be returned, got %s", commit)
	}
}

func TestAutomatedFlagParses(t *testing.T) {
	bot, err := slackbot.NewTestBot()
	if err != nil {
		t.Fatal(err)
	}
	ext := &keybot{}
	// Test bot is in dry run mode, so the dedupe check is bypassed and no
	// network access or state file is needed.
	out, err := ext.Run(bot, "", []string{"build", "ios", "--automated"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "I would have run a launchd job (keybase.build.ios)") {
		t.Errorf("Unexpected output: %s", out)
	}
}

func TestAutomatedPausedBlocksBuild(t *testing.T) {
	bot := slackbot.NewBot(slackbot.NewConfig(false, true), "testbot", "", &slackbot.SlackBotBackend{})
	ext := &keybot{}
	out, err := ext.Run(bot, "", []string{"build", "mobile", "--automated"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "I'm paused so I can't do that, but I would have run a launchd job (keybase.build.mobile)" {
		t.Errorf("Unexpected output: %s", out)
	}
}
