// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package main

import (
	"flag"
	"log"
	"os"

	"github.com/keybase/slackbot"
	"github.com/nlopes/slack"
)

var ignoreError = flag.Bool("i", false, "Ignore error (always exit 0)")

func error(s string) {
	if *ignoreError {
		log.Printf("Error: %s", s)
		os.Exit(0)
	}
	log.Fatal(s)
}

func main() {
	flag.Parse()

	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		error("SLACK_TOKEN is not set")
	}

	channel := os.Getenv("SLACK_CHANNEL")
	if channel == "" {
		error("SLACK_CHANNEL is not set")
	}

	api := slack.New(token)
	//api.SetDebug(true)

	channelIDs, err := slackbot.LoadChannelIDs(*api)
	if err != nil {
		error(err.Error())
	}

	text := flag.Arg(0)

	params := slack.NewPostMessageParameters()
	params.AsUser = true
	_, _, err = api.PostMessage(channelIDs[channel], text, params)
	if err != nil {
		error(err.Error())
	}

}
