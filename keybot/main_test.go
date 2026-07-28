// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/keybase/slackbot"
	"github.com/stretchr/testify/require"
)

func TestAddBasicCommands(t *testing.T) {
	bot, err := slackbot.NewTestBot()
	if err != nil {
		t.Fatal(err)
	}
	addBasicCommands(bot)
}

func TestPromoteRelease(t *testing.T) {
	bot, err := slackbot.NewTestBot()
	if err != nil {
		t.Fatal(err)
	}
	ext := &keybot{}
	out, err := ext.Run(bot, "", []string{"release", "promote", "darwin", "1.2.3"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "I would have run a launchd job (keybase.release.promote)\nPath: \"github.com/keybase/slackbot/scripts/release.promote.sh\"\nEnvVars: []launchd.EnvVar{launchd.EnvVar{Key:\"RELEASE_TO_PROMOTE\", Value:\"1.2.3\"}, launchd.EnvVar{Key:\"DRY_RUN\", Value:\"false\"}}" {
		t.Errorf("Unexpected output: %s", out)
	}

	out, err = ext.Run(bot, "", []string{"release", "promote", "darwin", "1.2.3", "--dry-run"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "I would have run a launchd job (keybase.release.promote)\nPath: \"github.com/keybase/slackbot/scripts/release.promote.sh\"\nEnvVars: []launchd.EnvVar{launchd.EnvVar{Key:\"RELEASE_TO_PROMOTE\", Value:\"1.2.3\"}, launchd.EnvVar{Key:\"DRY_RUN\", Value:\"true\"}}" {
		t.Errorf("Unexpected output: %s", out)
	}
}

func TestIgnorePauseFlagParses(t *testing.T) {
	bot, err := slackbot.NewTestBot()
	if err != nil {
		t.Fatal(err)
	}
	ext := &keybot{}
	out, err := ext.Run(bot, "", []string{"build", "darwin", "--ignore-pause"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "I would have run a launchd job (keybase.build.darwin)") {
		t.Errorf("Unexpected output: %s", out)
	}
}

func TestPausedBlocksRunScript(t *testing.T) {
	bot, err := slackbot.NewTestBot()
	require.NoError(t, err)
	bot.Config().SetDryRun(false)
	bot.Config().SetPaused(true)
	ext := &keybot{}
	out, err := ext.Run(bot, "", []string{"build", "darwin"})
	require.NoError(t, err)
	require.Equal(t, "I'm paused so I can't do that, but I would have run a launchd job (keybase.build.darwin)", out)
}

func TestGDiffPlainPathDoesNotPanic(t *testing.T) {
	goPath := t.TempDir()
	repository := filepath.Join(goPath, "src", "github.com", "keybase", "client")
	require.NoError(t, os.MkdirAll(repository, 0o750))
	t.Setenv("GOPATH", goPath)
	bot, err := slackbot.NewTestBot()
	require.NoError(t, err)
	ext := &keybot{}
	_, err = ext.Run(bot, "", []string{"gdiff", "github.com/keybase/client"})
	require.NoError(t, err)
}

func TestWinBotGCleanPlainPathDoesNotPanic(t *testing.T) {
	goPath := t.TempDir()
	repository := filepath.Join(goPath, "src", "github.com", "keybase", "client")
	require.NoError(t, os.MkdirAll(repository, 0o750))
	t.Setenv("GOPATH", goPath)
	bot, err := slackbot.NewTestBot()
	require.NoError(t, err)
	bot.Config().SetDryRun(false)
	ext := &winbot{}
	out, err := ext.Run(bot, "", []string{"gclean", "github.com/keybase/client"})
	require.NoError(t, err)
	require.Equal(t, "Not a git repo", out)
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
