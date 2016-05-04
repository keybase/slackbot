// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package slackbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/user"
	"path/filepath"
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

func ReadConfigOrDefault() Config {
	defaultConfig := Config{
		DryRun: true,
		Paused: false,
	}

	path, err := getConfigPath()

	if err != nil {
		return defaultConfig
	}

	fileBytes, err := ioutil.ReadFile(path)

	if err != nil {
		log.Printf("Couldn't read config file:%s\n", err)
		return defaultConfig
	}

	var config Config
	err = json.Unmarshal(fileBytes, &config)
	if err != nil {
		log.Printf("Couldn't read config file:%s\n", err)
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
	config := ReadConfigOrDefault()

	if config.DryRun {
		return fmt.Sprintf("Dry Run: %s", c.Description()), nil
	}

	newConfig, err := updateConfig(c.Updater)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Config is now: %+v", newConfig), nil
}

// ShowResult will always show the results of a config change
func (c ConfigCommand) ShowResult() bool {
	return true
}

// Description describes how it will change the config
func (c ConfigCommand) Description() string {
	return c.Desc
}
