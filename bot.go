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

// Bot describes a generic bot
type Bot interface {
	Name() string
	Runner() Runner
	AddCommand(trigger string, command Command)
	SendMessage(text string, channel string)
	HelpMessage() string
	SetHelp(help string)
	Label() string
	SetDefault(command Command)
	Listen()
}

// SlackBot is a Slack bot
type SlackBot struct {
	api            *slack.Client
	rtm            *slack.RTM
	commands       map[string]Command
	defaultCommand Command
	channelIDs     map[string]string
	help           string
	name           string
	label          string
	runner         Runner
}

// Runner can execute bot inputs
type Runner interface {
	Run(b Bot, channel string, args []string) (string, error)
}

// NewBot constructs a bot from a Slack token
func NewBot(token string, name string, label string, runner Runner) (Bot, error) {
	api := slack.New(token)
	//api.SetDebug(true)

	channelIDs, err := LoadChannelIDs(*api)
	if err != nil {
		return nil, err
	}

	bot := newBot(runner)
	bot.api = api
	bot.rtm = api.NewRTM()
	bot.channelIDs = channelIDs
	bot.name = name
	bot.label = label

	return bot, nil
}

func newBot(runner Runner) *SlackBot {
	bot := SlackBot{
		runner:   runner,
		commands: make(map[string]Command),
	}
	return &bot
}

// NewTestBot returns a bot for testing
func NewTestBot(runner Runner) (Bot, error) {
	return newBot(runner), nil
}

// AddCommand adds a command to the Bot
func (b *SlackBot) AddCommand(trigger string, command Command) {
	b.commands[trigger] = command
}

// Name returns bot name
func (b *SlackBot) Name() string {
	return b.name
}

// Label returns bot label
func (b *SlackBot) Label() string {
	return b.label
}

// Runner returns bot runner
func (b *SlackBot) Runner() Runner {
	return b.runner
}

// SetHelp sets the help info
func (b *SlackBot) SetHelp(help string) {
	b.help = help
}

// SetDefault is the default command, if no command added for trigger
func (b *SlackBot) SetDefault(command Command) {
	b.defaultCommand = command
}

// RunCommand runs a command
func (b *SlackBot) RunCommand(args []string, channel string) error {
	if len(args) == 0 || args[0] == "help" {
		b.SendHelpMessage(channel)
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

	log.Printf("Command: %#v\n", command)

	msg := fmt.Sprintf("Sure, I will `%s`.", strings.Join(args, " "))
	b.SendMessage(msg, channel)

	go b.run(args, command, channel)
	return nil
}

func (b *SlackBot) run(args []string, command Command, channel string) {
	out, err := command.Run(b, channel, args)
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
func (b *SlackBot) SendMessage(text string, channel string) {
	cid := b.channelIDs[channel]
	if cid == "" {
		cid = channel
	}
	if b.rtm != nil {
		b.rtm.SendMessage(b.rtm.NewOutgoingMessage(text, cid))
	}
}

// Triggers returns list of supported triggers
func (b *SlackBot) Triggers() []string {
	triggers := make([]string, 0, len(b.commands))
	for trigger := range b.commands {
		triggers = append(triggers, trigger)
	}
	sort.Strings(triggers)
	return triggers
}

// HelpMessage is the default help message for the bot
func (b *SlackBot) HelpMessage() string {
	w := new(tabwriter.Writer)
	buf := new(bytes.Buffer)
	w.Init(buf, 0, 8, 0, '\t', 0)
	fmt.Fprintln(w, "Command\tDescription")
	for _, trigger := range b.Triggers() {
		command := b.commands[trigger]
		fmt.Fprintln(w, fmt.Sprintf("%s\t%s", trigger, command.Description()))
	}
	_ = w.Flush()

	return SlackBlockQuote(buf.String())
}

// SendHelpMessage displays help message to the channel
func (b *SlackBot) SendHelpMessage(channel string) {
	help := b.help
	if help == "" {
		help = b.HelpMessage()
	}
	b.SendMessage(help, channel)
}

// Listen starts listening on the connection
func (b *SlackBot) Listen() {
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

// SlackBlockQuote returns the string block-quoted
func SlackBlockQuote(s string) string {
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
