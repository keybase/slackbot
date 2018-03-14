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

// Config is the state of the build bot
type Config interface {
	// Paused will prevent any commands from running
	Paused() bool
	// SetPaused changes paused
	SetPaused(paused bool)
	// DryRun will print out what it plans to do without doing it
	DryRun() bool
	// SetDryRun changes dry run
	SetDryRun(dryRun bool)
	// Save persists config
	Save() error
}

type config struct {
	// These must be public for json serialization.
	DryRunField bool
	PausedField bool
}

// Paused if paused
func (c config) Paused() bool {
	return c.PausedField
}

// DryRun if dry run enabled
func (c config) DryRun() bool {
	return c.DryRunField
}

// SetPaused changes paused
func (c *config) SetPaused(paused bool) {
	c.PausedField = paused
}

// SetDryRun changes dry run
func (c *config) SetDryRun(dryRun bool) {
	c.DryRunField = dryRun
}

func getConfigPath() (string, error) {
	currentUser, err := user.Current()

	if err != nil {
		return "", err
	}

	return filepath.Join(currentUser.HomeDir, ".keybot"), nil
}

// NewConfig returns default config
func NewConfig(dryRun, paused bool) Config {
	return &config{
		DryRunField: dryRun,
		PausedField: paused,
	}
}

// ReadConfigOrDefault returns config stored or default
func ReadConfigOrDefault() Config {
	cfg := readConfigOrDefault()
	return &cfg
}

func readConfigOrDefault() config {
	defaultConfig := config{
		DryRunField: true,
		PausedField: false,
	}

	path, err := getConfigPath()

	if err != nil {
		return defaultConfig
	}

	fileBytes, err := ioutil.ReadFile(path)

	if err != nil {
		return defaultConfig
	}

	var cfg config
	err = json.Unmarshal(fileBytes, &cfg)
	if err != nil {
		log.Printf("Couldn't read config file: %s\n", err)
		return defaultConfig
	}

	return cfg
}

func (c config) Save() error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	path, err := getConfigPath()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, b, 0644)
	if err != nil {
		return err
	}

	return nil
}

// NewShowConfigCommand returns command that shows config
func NewShowConfigCommand(config Config) Command {
	return &showConfigCommand{config: config}
}

type showConfigCommand struct {
	config Config
}

func (c showConfigCommand) Run(_ string, _ []string) (string, error) {
	if !c.config.Paused() && !c.config.DryRun() {
		return "I'm running normally.", nil
	}
	lines := []string{}
	if c.config.Paused() {
		lines = append(lines, "I'm paused.")
	}
	if c.config.DryRun() {
		lines = append(lines, "I'm in dry run mode.")
	}
	return strings.Join(lines, " "), nil
}

func (c showConfigCommand) ShowResult() bool {
	return true
}

func (c showConfigCommand) Description() string {
	return "Shows config"
}

// NewToggleDryRunCommand returns toggle dry run command
func NewToggleDryRunCommand(config Config) Command {
	return &toggleDryRunCommand{config: config}
}

type toggleDryRunCommand struct {
	config Config
}

func (c *toggleDryRunCommand) Run(_ string, _ []string) (string, error) {
	c.config.SetDryRun(!c.config.DryRun())
	err := c.config.Save()
	if err != nil {
		return "", err
	}

	if c.config.DryRun() {
		return "We are in dry run mode.", nil
	}
	return "We are not longer in dry run mode", nil
}

func (c toggleDryRunCommand) ShowResult() bool {
	return true
}

func (c toggleDryRunCommand) Description() string {
	return "Toggles the dry run mode"
}

// NewPauseCommand pauses
func NewPauseCommand(config Config) Command {
	return &pauseCommand{
		config: config,
		pauses: true,
	}
}

// NewResumeCommand resumes
func NewResumeCommand(config Config) Command {
	return &pauseCommand{
		config: config,
		pauses: false,
	}
}

type pauseCommand struct {
	config Config
	pauses bool
}

// Run toggles the dry run state. (Itself is never run under dry run mode)
func (c *pauseCommand) Run(_ string, _ []string) (string, error) {
	log.Printf("Setting paused: %v\n", c.pauses)
	c.config.SetPaused(c.pauses)
	err := c.config.Save()
	if err != nil {
		return "", err
	}

	if c.config.Paused() {
		return "I am paused.", nil
	}
	return "I have resumed.", nil
}

// ShowResult always shows results for toggling dry run
func (c pauseCommand) ShowResult() bool {
	return true
}

// Description describes what it does
func (c pauseCommand) Description() string {
	if c.pauses {
		return "Pauses the bot"
	}
	return "Resumes the bot"
}
