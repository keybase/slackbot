// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package slackbot

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os/user"
	"path/filepath"
	"strings"
)

// ConfigCommand is a Command that sets some saved state
type ConfigCommand struct {
	Desc    string
	Updater func(c Config) (Config, error)
}

// Config is the state of the build bot.
// DryRun will print out what it plans to do without doing it
// Paused will prevent any builds in the future from running
type Config struct {
	DryRun bool
	Paused bool
}

func getConfigPath() (string, error) {
	currentUser, err := user.Current()

	if err != nil {
		return "", err
	}

	return filepath.Join(currentUser.HomeDir, ".keybot"), nil
}

// ReadConfigOrDefault returns config
func ReadConfigOrDefault() Config {
	defaultConfig := Config{
		DryRun: false,
		Paused: false,
	}

	path, err := getConfigPath()

	if err != nil {
		return defaultConfig
	}

	fileBytes, err := ioutil.ReadFile(path)

	if err != nil {
		return defaultConfig
	}

	var config Config
	err = json.Unmarshal(fileBytes, &config)
	if err != nil {
		log.Printf("Couldn't read config file: %s\n", err)
		return defaultConfig
	}

	return config
}

func updateConfig(updater func(c Config) (Config, error)) (Config, error) {
	config := ReadConfigOrDefault()
	newConfig, err := updater(config)

	if err != nil {
		return config, err
	}

	b, err := json.Marshal(newConfig)

	if err != nil {
		return config, err
	}

	path, err := getConfigPath()

	if err != nil {
		return config, err
	}

	err = ioutil.WriteFile(path, b, 0644)

	if err != nil {
		return config, err
	}

	return newConfig, nil
}

// Run the config change
func (c ConfigCommand) Run(_ string, _ []string) (string, error) {
	newConfig, err := updateConfig(c.Updater)

	if err != nil {
		return "", err
	}

	if !newConfig.Paused && !newConfig.DryRun {
		return "I'm running normally.", nil
	}

	lines := []string{}
	if newConfig.Paused {
		lines = append(lines, "I'm paused.")
	}
	if newConfig.DryRun {
		lines = append(lines, "I'm in dry run mode.")
	}

	return strings.Join(lines, " "), nil
}

// ShowResult will always show the results of a config change
func (c ConfigCommand) ShowResult() bool {
	return true
}

// Description describes how it will change the config
func (c ConfigCommand) Description() string {
	return c.Desc
}
