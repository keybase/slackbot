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

func handleError(s string, text string) {
	if *ignoreError {
		log.Printf("[Unable to send: %s] %s", s, text)
		os.Exit(0)
	}
	log.Fatal(s)
}

func main() {
	flag.Parse()
	text := flag.Arg(0)

	channel := os.Getenv("SLACK_CHANNEL")
	if channel == "" {
		handleError("SLACK_CHANNEL is not set", text)
	}

	api := slack.New(slackbot.GetTokenFromEnv())
	//api.SetDebug(true)

	channelIDs, err := slackbot.LoadChannelIDs(*api)
	if err != nil {
		handleError(err.Error(), text)
	}

	params := slack.NewPostMessageParameters()
	params.AsUser = true
	channelID := channelIDs[channel]
	_, _, err = api.PostMessage(channelID, text, params)
	if err != nil {
		handleError(err.Error(), text)
	} else {
		log.Printf("[%s (%s)] %s\n", channel, channelID, text)
	}
}
