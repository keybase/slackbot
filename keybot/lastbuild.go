// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/keybase/slackbot"
	"github.com/keybase/slackbot/launchd"
)

const clientRepoURL = "https://github.com/keybase/client.git"

// lastBuildPath is the file holding the commit of the last successful
// automated build for a job. run.sh writes it when the build script finishes
// successfully.
func lastBuildPath(label string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".keybot.lastbuild."+label), nil
}

func readLastBuiltCommit(path string) string {
	fileBytes, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(fileBytes))
}

// currentClientCommit returns the commit an automated build would build: the
// explicit client commit if one was given, otherwise the remote HEAD of the
// client repo.
func currentClientCommit(clientCommit string) (string, error) {
	if clientCommit != "" {
		return clientCommit, nil
	}
	//nolint:gosec,noctx // git is a trusted system binary with safe arguments, no context available
	out, err := exec.Command("git", "ls-remote", clientRepoURL, "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("Error in git ls-remote: %w", err)
	}
	fields := strings.Fields(string(out))
	if len(fields) == 0 {
		return "", fmt.Errorf("Empty output from git ls-remote")
	}
	return fields[0], nil
}

// runBuildScript runs a build like runScript, but automated (timed) builds
// skip the build if the commit was already built for this job. Build numbers
// auto increment now, so rebuilding the same commit would produce a duplicate
// build. The commit is recorded by run.sh only after the build script
// succeeds, so a failed build is retried on the next timed run.
func runBuildScript(bot *slackbot.Bot, channel string, env launchd.Env, script launchd.Script, ignorePause bool, automated bool, clientCommit string) (string, error) {
	if !automated || bot.Config().DryRun() || (bot.Config().Paused() && !ignorePause) {
		return runScript(bot, channel, env, script, ignorePause)
	}

	statePath, err := lastBuildPath(script.Label)
	if err != nil {
		log.Printf("Couldn't get last build path for %s (%s), building anyway\n", script.Label, err)
		return runScript(bot, channel, env, script, ignorePause)
	}

	commit, err := currentClientCommit(clientCommit)
	if err != nil {
		log.Printf("Couldn't determine client commit for %s (%s), building anyway\n", script.Label, err)
		return runScript(bot, channel, env, script, ignorePause)
	}

	if commit == readLastBuiltCommit(statePath) {
		return fmt.Sprintf("I already did an automated build of `%s` for commit `%s`, so I'm skipping this one.", script.Label, commit), nil
	}

	// Pin the build to the resolved commit so the state file can't record a
	// different commit than was built if remote HEAD moves before the build
	// script fetches.
	script.EnvVars = setEnvVar(script.EnvVars, "CLIENT_COMMIT", commit)
	script.EnvVars = append(script.EnvVars,
		launchd.EnvVar{Key: "AUTOMATED_BUILD_COMMIT", Value: commit},
		launchd.EnvVar{Key: "AUTOMATED_BUILD_COMMIT_PATH", Value: statePath},
	)
	return runScript(bot, channel, env, script, ignorePause)
}

// setEnvVar replaces the value of key in envVars, appending it if not present.
func setEnvVar(envVars []launchd.EnvVar, key string, value string) []launchd.EnvVar {
	for i := range envVars {
		if envVars[i].Key == key {
			envVars[i].Value = value
			return envVars
		}
	}
	return append(envVars, launchd.EnvVar{Key: key, Value: value})
}
