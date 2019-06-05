package slackbot

import (
	"log"

	"github.com/nlopes/slack"
)

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
