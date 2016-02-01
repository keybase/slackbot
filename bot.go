// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package slackbot

import (
	"fmt"
	"log"
	"strings"

	"github.com/nlopes/slack"
)

type Bot struct {
	api        *slack.Client
	rtm        *slack.RTM
	commands   map[string]Command
	channelIDs map[string]string
}

func NewBot(token string) (*Bot, error) {
	api := slack.New(token)
	//api.SetDebug(true)

	channelIDs, err := LoadChannelIDs(*api)
	if err != nil {
		return nil, err
	}

	rtm := api.NewRTM()
	commands := make(map[string]Command)

	bot := Bot{api: api, rtm: rtm, commands: commands, channelIDs: channelIDs}
	return &bot, nil
}

func (b *Bot) AddCommand(trigger string, command Command) {
	b.commands[trigger] = command
}

func (b *Bot) RunCommand(trigger string, channel string) {
	if trigger == "help" {
		b.Help(channel)
		return
	}

	command, ok := b.commands[trigger]
	if !ok {
		log.Printf("Unrecognized command: %s", trigger)
		return
	}

	log.Printf("Command: %#v\n", command)
	b.SendMessage(fmt.Sprintf("Sure, I will !%s", trigger), channel)

	go b.run(trigger, command, channel)
}

func (b *Bot) run(trigger string, command Command, channel string) {
	out, err := command.Run()
	if err != nil {
		log.Printf("Error %s running: %#v; %s\n", err, command, out)
		b.SendMessage(fmt.Sprintf("Oops, there was an error in !%s", trigger), channel)
		return
	}
	log.Printf("Output: %s\n", out)
	if command.ShowResult() {
		b.SendMessage(out, channel)
	}
}

func (b *Bot) SendMessage(text string, channel string) {
	cid := b.channelIDs[channel]
	if cid == "" {
		cid = channel
	}
	b.rtm.SendMessage(b.rtm.NewOutgoingMessage(text, cid))
}

func (b *Bot) Help(channel string) {
	msgs := []string{}
	for trigger, command := range b.commands {
		msgs = append(msgs, fmt.Sprintf("%s: %s", trigger, command.Description()))
	}
	b.SendMessage(strings.Join(msgs, "\n"), channel)
}

func (b *Bot) Listen() {
	go b.rtm.ManageConnection()

Loop:
	for {
		select {
		case msg := <-b.rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.HelloEvent:

			case *slack.ConnectedEvent:

			case *slack.MessageEvent:
				text := strings.TrimSpace(ev.Text)
				if strings.HasPrefix(text, "!") {
					cmd := text[1:]
					b.RunCommand(cmd, ev.Channel)
				}

			case *slack.PresenceChangeEvent:
				//log.Printf("Presence Change: %v\n", ev)

			case *slack.LatencyReport:
				//log.Printf("Current latency: %v\n", ev.Value)

			case *slack.RTMError:
				log.Printf("Error: %s\n", ev.Error())

			case *slack.InvalidAuthEvent:
				log.Printf("Invalid credentials\n")
				break Loop

			default:
				// log.Printf("Unexpected: %v\n", msg.Data)
			}
		}
	}
}
