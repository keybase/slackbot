// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"bytes"
	"fmt"

	"github.com/keybase/slackbot"
	"github.com/keybase/slackbot/cli"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type extension struct{}

func (e *extension) Run(bot slackbot.Bot, channel string, args []string) (string, error) {
	app := kingpin.New("examplebot", "Kingpin extension")
	app.Terminate(nil)
	stringBuffer := new(bytes.Buffer)
	app.Writer(stringBuffer)

	testCmd := app.Command("echo", "Echo")
	testCmdEchoFlag := testCmd.Flag("output", "Output to echo").Required().String()

	cmd, usage, cmdErr := cli.Parse(app, args, stringBuffer)
	if usage != "" || cmdErr != nil {
		return usage, cmdErr
	}

	if bot.Config().DryRun() {
		return fmt.Sprintf("I would have run: `%#v`", cmd), nil
	}

	switch cmd {
	case testCmd.FullCommand():
		return *testCmdEchoFlag, nil
	}
	return cmd, nil
}

func (e *extension) Help(bot slackbot.Bot) string {
	out, err := e.Run(bot, "", nil)
	if err != nil {
		return fmt.Sprintf("Error getting help: %s", err)
	}
	return out
}
