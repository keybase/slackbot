// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/keybase/slackbot"
	"github.com/keybase/slackbot/launchd"
	"github.com/stretchr/testify/require"
)

func TestLastBuildPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path, err := lastBuildPath("keybase.build.ios")
	if err != nil {
		t.Fatal(err)
	}
	expected := filepath.Join(home, ".keybot.lastbuild.keybase.build.ios")
	if path != expected {
		t.Errorf("Expected %q, got %q", expected, path)
	}
}

func TestSetEnvVar(t *testing.T) {
	envVars := []launchd.EnvVar{
		{Key: "CLIENT_COMMIT", Value: ""},
		{Key: "CHECK_CI", Value: "false"},
	}
	envVars = setEnvVar(envVars, "CLIENT_COMMIT", "abc123")
	if len(envVars) != 2 || envVars[0].Value != "abc123" {
		t.Errorf("Expected existing CLIENT_COMMIT replaced, got %+v", envVars)
	}
	envVars = setEnvVar(envVars, "NEW_KEY", "v")
	if len(envVars) != 3 || envVars[2].Key != "NEW_KEY" || envVars[2].Value != "v" {
		t.Errorf("Expected NEW_KEY appended, got %+v", envVars)
	}
}

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
	bot, err := slackbot.NewTestBot()
	require.NoError(t, err)
	bot.Config().SetDryRun(false)
	bot.Config().SetPaused(true)
	ext := &keybot{}
	out, err := ext.Run(bot, "", []string{"build", "mobile", "--automated"})
	require.NoError(t, err)
	require.Equal(t, "I'm paused so I can't do that, but I would have run a launchd job (keybase.build.mobile)", out)
}
