// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package slackbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"os/user"
)

// Command is the interface the bot uses to run things
type Command interface {
	Run() (string, error)
	ShowResult() bool // Whether to output result back to channel
	Description() string
}

// ExecCommand is a Command that does an exec.Command(...) on the system
type ExecCommand struct {
	exec        string   // Command to execute
	args        []string // Args for exec.Command
	showResult  bool
	description string
}

// ConfigCommand is a Command that sets some saved state
type ConfigCommand struct {
	Desc    string
	Updater func(c Config) (Config, error)
}

// ToggleDryRunCommand is a special command that toggles dry run state
type ToggleDryRunCommand struct{}

// Config is the state of the build bot.
// DryRun will print out what it plans to do without doing it
// Paused will prevent any builds in the future from running
type Config struct {
	DryRun bool
	Paused bool
}

func getConfigPath() (string, error) {
	u, err := user.Current()

	if err != nil {
		return "", err
	}

	return u.HomeDir + "/.keybot", nil
}

func readConfigOrDefault() Config {
	defaultConfig := Config{
		DryRun: true,
		Paused: false,
	}

	path, err := getConfigPath()

	if err != nil {
		return defaultConfig
	}

	b, err := ioutil.ReadFile(path)

	if err != nil {
		fmt.Printf("Couldn't read config file:%s\n", err)
		return defaultConfig
	}

	var config Config
	err = json.Unmarshal(b, &config)
	if err != nil {
		fmt.Printf("Couldn't read config file:%s\n", err)
		return defaultConfig
	}

	return config
}

func updateConfig(updater func(c Config) (Config, error)) (Config, error) {
	config := readConfigOrDefault()
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

// NewExecCommand creates an ExecCommand
func NewExecCommand(exec string, args []string, showResult bool, description string) ExecCommand {
	return ExecCommand{
		exec:        exec,
		args:        args,
		showResult:  showResult,
		description: description,
	}
}

// Run runs the exec command
func (c ExecCommand) Run() (string, error) {
	config := readConfigOrDefault()

	if config.DryRun {
		return fmt.Sprintf("Dry Run: would have ran `%s` with args: %s", c.exec, c.args), nil
	}

	out, err := exec.Command(c.exec, c.args...).Output()
	outAsString := fmt.Sprintf("%s", out)
	return outAsString, err
}

// ShowResult decides whether to show the results from the exec
func (c ExecCommand) ShowResult() bool {
	config := readConfigOrDefault()
	return config.DryRun || c.showResult
}

// Description describes the command
func (c ExecCommand) Description() string {
	return c.description
}

// Run the config change
func (c ConfigCommand) Run() (string, error) {
	config := readConfigOrDefault()

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

// Run toggles the dry run state. (Itself is never run under dry run mode)
func (c ToggleDryRunCommand) Run() (string, error) {
	config, err := updateConfig(func(c Config) (Config, error) {
		c.DryRun = !c.DryRun
		return c, nil
	})

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Dry Run Value is now: %t", config.DryRun), nil
}

// ShowResult always shows results for toggling dry run
func (c ToggleDryRunCommand) ShowResult() bool {
	return true
}

// Description describes what it does
func (c ToggleDryRunCommand) Description() string {
	return "Toggles the Dry Run value"
}
