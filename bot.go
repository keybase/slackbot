// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package slackbot

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/nlopes/slack"
)

type BotCommandRunner interface {
	RunCommand(args []string, channel string) error
}

type BotBackend interface {
	SendMessage(text string, channel string)
	Listen(BotCommandRunner)
}

// Bot describes a generic bot
type Bot struct {
	backend        BotBackend
	help           string
	name           string
	label          string
	config         Config
	commands       map[string]Command
	defaultCommand Command
}

func NewBot(config Config, name, label string, backend BotBackend) *Bot {
	return &Bot{
		backend:  backend,
		config:   config,
		commands: make(map[string]Command),
		name:     name,
		label:    label,
	}
}

func (b *Bot) Name() string {
	return b.name
}

func (b *Bot) Config() Config {
	return b.config
}

func (b *Bot) AddCommand(trigger string, command Command) {
	b.commands[trigger] = command
}

func (b *Bot) triggers() []string {
	triggers := make([]string, 0, len(b.commands))
	for trigger := range b.commands {
		triggers = append(triggers, trigger)
	}
	sort.Strings(triggers)
	return triggers
}

// HelpMessage is the default help message for the bot
func (b *Bot) HelpMessage() string {
	w := new(tabwriter.Writer)
	buf := new(bytes.Buffer)
	w.Init(buf, 8, 8, 8, ' ', 0)
	fmt.Fprintln(w, "Command\tDescription")
	for _, trigger := range b.triggers() {
		command := b.commands[trigger]
		fmt.Fprintln(w, fmt.Sprintf("%s\t%s", trigger, command.Description()))
	}
	_ = w.Flush()
	return BlockQuote(buf.String())
}

func (b *Bot) SetHelp(help string) {
	b.help = help
}

func (b *Bot) Label() string {
	return b.label
}

func (b *Bot) SetDefault(command Command) {
	b.defaultCommand = command
}

// RunCommand runs a command
func (b *Bot) RunCommand(args []string, channel string) error {
	if len(args) == 0 || args[0] == "help" {
		b.sendHelpMessage(channel)
		return nil
	}

	command, ok := b.commands[args[0]]
	if !ok {
		if b.defaultCommand != nil {
			command = b.defaultCommand
		} else {
			return fmt.Errorf("Unrecognized command: %q", args)
		}
	}

	if args[0] != "resume" && args[0] != "config" && b.Config().Paused() {
		b.backend.SendMessage("I can't do that, I'm paused.", channel)
		return nil
	}

	go b.run(args, command, channel)
	return nil
}

func (b *Bot) run(args []string, command Command, channel string) {
	out, err := command.Run(channel, args)
	if err != nil {
		log.Printf("Error %s running: %#v; %s\n", err, command, out)
		b.backend.SendMessage(fmt.Sprintf("Oops, there was an error in %q:\n%s", strings.Join(args, " "),
			BlockQuote(out)), channel)
		return
	}
	log.Printf("Output: %s\n", out)
	if command.ShowResult() {
		b.backend.SendMessage(out, channel)
	}
}

func (b *Bot) sendHelpMessage(channel string) {
	help := b.help
	if help == "" {
		help = b.HelpMessage()
	}
	b.backend.SendMessage(help, channel)
}

func (b *Bot) SendMessage(text string, channel string) {
	b.backend.SendMessage(text, channel)
}

func (b *Bot) Listen() {
	b.backend.Listen(b)
}

// SlackBotBackend is a Slack bot backend
type SlackBotBackend struct {
	api *slack.Client
	rtm *slack.RTM

	channelIDs map[string]string
}

// NewSlackBotBackend constructs a bot backend from a Slack token
func NewSlackBotBackend(token string) (BotBackend, error) {
	api := slack.New(token)
	//api.SetDebug(true)

	channelIDs, err := LoadChannelIDs(*api)
	if err != nil {
		return nil, err
	}

	bot := &SlackBotBackend{}
	bot.api = api
	bot.rtm = api.NewRTM()
	bot.channelIDs = channelIDs
	return bot, nil
}

// NewTestBot returns a bot for testing
func NewTestBot() (*Bot, error) {
	return nil, nil
}

// SendMessage sends a message to a channel
func (b *SlackBotBackend) SendMessage(text string, channel string) {
	cid := b.channelIDs[channel]
	if cid == "" {
		cid = channel
	}

	if channel == "" {
		log.Printf("No channel to send message: %s", text)
		return
	}

	if b.rtm != nil {
		b.rtm.SendMessage(b.rtm.NewOutgoingMessage(text, cid))
	} else {
		log.Printf("Unable to send message: %s", text)
	}
}

// Listen starts listening on the connection
func (b *SlackBotBackend) Listen(runner BotCommandRunner) {
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
		msg := <-b.rtm.IncomingEvents
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:

		case *slack.ConnectedEvent:

		case *slack.MessageEvent:
			args := parseInput(ev.Text)
			if len(args) > 0 && args[0] == commandPrefix {
				cmd := args[1:]
				if err := runner.RunCommand(cmd, ev.Channel); err != nil {
					log.Printf("failed to run command: %s\n", err)
				}
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

// BlockQuote returns the string block-quoted
func BlockQuote(s string) string {
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	return "```\n" + s + "```"
}

// GetTokenFromEnv returns slack token from the environment
func GetTokenFromEnv() string {
	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		log.Fatal("SLACK_TOKEN is not set")
	}
	return token
}

func isSpace(r rune) bool {
	switch r {
	case ' ', '\t', '\r', '\n':
		return true
	}
	return false
}

func parseInput(s string) []string {
	buf := ""
	args := []string{}
	var escaped, doubleQuoted, singleQuoted bool
	for _, r := range s {
		if escaped {
			buf += string(r)
			escaped = false
			continue
		}

		if r == '\\' {
			if singleQuoted {
				buf += string(r)
			} else {
				escaped = true
			}
			continue
		}

		if isSpace(r) {
			if singleQuoted || doubleQuoted {
				buf += string(r)
			} else if buf != "" {
				args = append(args, buf)
				buf = ""
			}
			continue
		}

		switch r {
		case '"':
			if !singleQuoted {
				doubleQuoted = !doubleQuoted
				continue
			}
		case '\'':
			if !doubleQuoted {
				singleQuoted = !singleQuoted
				continue
			}
		}

		buf += string(r)
	}
	if buf != "" {
		args = append(args, buf)
	}
	return args
}
