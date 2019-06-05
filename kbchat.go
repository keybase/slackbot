package slackbot

import (
	"fmt"
	"log"

	"github.com/keybase/go-keybase-chat-bot/kbchat"
)

type KeybaseChatBotBackend struct {
	name   string
	convID string
	kbc    *kbchat.API
}

func NewKeybaseChatBotBackend(name string, convID string, opts kbchat.RunOptions) (BotBackend, error) {
	var err error
	bot := &KeybaseChatBotBackend{
		convID: convID,
		name:   name,
	}
	if bot.kbc, err = kbchat.Start(opts); err != nil {
		return nil, err
	}
	return bot, nil
}

func (b *KeybaseChatBotBackend) SendMessage(text string, convID string) {
	if convID != b.convID {
		// bail out if not on configured conv ID
		log.Printf("SendMessage: refusing to send on non-configured convID: %s != %s\n", convID, b.convID)
		return
	}
	if len(text) == 0 {
		log.Printf("SendMessage: skipping blank message")
		return
	}
	log.Printf("sending message: convID: %s text: %s", convID, text)
	if err := b.kbc.SendMessageByConvID(convID, text); err != nil {
		log.Printf("SendMessage: failed to send: %s\n", err)
	}
}

func (b *KeybaseChatBotBackend) Listen(runner BotCommandRunner) {
	sub, err := b.kbc.ListenForNewTextMessages()
	if err != nil {
		panic(fmt.Sprintf("failed to set up listen: %s", err))
	}
	commandPrefix := "!" + b.name
	for {
		msg, err := sub.Read()
		if err != nil {
			log.Printf("Listen: failed to read message: %s", err)
			continue
		}
		if msg.Message.Content.Type != "text" {
			continue
		}
		args := parseInput(msg.Message.Content.Text.Body)
		if len(args) > 0 && args[0] == commandPrefix && b.convID == msg.Message.ConversationID {
			cmd := args[1:]
			runner.RunCommand(cmd, b.convID)
		}
	}
}
