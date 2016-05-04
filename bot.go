// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package slackbot

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/nlopes/slack"
)

// Bot defines a Slack bot
type Bot struct {
	api        *slack.Client
	rtm        *slack.RTM
	commands   map[string]Command
	channelIDs map[string]string
}

// NewBot constructs a bot from a Slack token
func NewBot(token string) (*Bot, error) {
	api := slack.New(token)
	//api.SetDebug(true)

	channelIDs, err := LoadChannelIDs(*api)
	if err != nil {
		return nil, err
	}

	bot := newBot()
	bot.api = api
	bot.rtm = api.NewRTM()
	bot.channelIDs = channelIDs

	return bot, nil
}

func newBot() *Bot {
	bot := Bot{}
	bot.commands = make(map[string]Command)
	return &bot
}

// NewTestBot returns a bot for testing
func NewTestBot() (*Bot, error) {
	return newBot(), nil
}

// AddCommand adds a command to the Bot
func (b *Bot) AddCommand(trigger string, command Command) {
	b.commands[trigger] = command
}

// RunCommand runs a command
func (b *Bot) RunCommand(args []string, channel string) {
	if len(args) == 0 || args[0] == "help" {
		b.Help(channel)
		return
	}

	command, ok := b.commands[args[0]]
	if !ok {
		log.Printf("Unrecognized command: %q", args)
		return
	}

	log.Printf("Command: %#v\n", command)
	b.SendMessage(fmt.Sprintf("Sure, I will %s.", args[0]), channel)

	go b.run(args, command, channel)
}

func (b *Bot) run(args []string, command Command, channel string) {
	out, err := command.Run(channel, args)
	if err != nil {
		log.Printf("Error %s running: %#v; %s\n", err, command, out)
		b.SendMessage(fmt.Sprintf("Oops, there was an error in %q:\n%s", strings.Join(args, " "), SlackBlockQuote(out)), channel)
		return
	}
	log.Printf("Output: %s\n", out)
	if command.ShowResult() {
		b.SendMessage(out, channel)
	}
}

// SendMessage sends a message to a channel
func (b *Bot) SendMessage(text string, channel string) {
	cid := b.channelIDs[channel]
	if cid == "" {
		cid = channel
	}
	b.rtm.SendMessage(b.rtm.NewOutgoingMessage(text, cid))
}

// Triggers returns list of supported triggers
func (b *Bot) Triggers() []string {
	triggers := make([]string, 0, len(b.commands))
	for trigger := range b.commands {
		triggers = append(triggers, trigger)
	}
	sort.Strings(triggers)
	return triggers
}

func (b *Bot) helpMessage() string {
	msgs := []string{}
	triggers := b.Triggers()
	for _, trigger := range triggers {
		command := b.commands[trigger]
		msgs = append(msgs, fmt.Sprintf("!%s: %s", trigger, command.Description()))
	}

	return SlackBlockQuote(strings.Join(msgs, "\n"))
}

// Help displays help message to the channel
func (b *Bot) Help(channel string) {
	b.SendMessage(b.helpMessage(), channel)
}

// Listen starts listening on the connection
func (b *Bot) Listen() {
	go b.rtm.ManageConnection()

	auth, err := b.api.AuthTest()
	if err != nil {
		panic(err)
	}
	// The Slack bot "tuxbot" should expect commands to start with "!tuxbot".
	log.Printf("Connected to Slack as %q", auth.User)
	commandPrefix := "!" + auth.User

Loop:
	for {
		select {
		case msg := <-b.rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.HelloEvent:

			case *slack.ConnectedEvent:

			case *slack.MessageEvent:
				args := strings.Fields(ev.Text)
				if len(args) > 0 && args[0] == commandPrefix {
					cmd := args[1:]
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

// SlackBlockQuote returns the string block-quoted
func SlackBlockQuote(s string) string {
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	return "```\n" + s + "```"
}
