package slackbot

import (
	"fmt"
	"log"
	"os"

	"github.com/keybase/go-keybase-chat-bot/kbchat"
	"github.com/keybase/go-keybase-chat-bot/kbchat/types/chat1"
)

type KeybaseChatBotBackend struct {
	name   string
	convID chat1.ConvIDStr
	kbc    *kbchat.API
}

func NewKeybaseChatBotBackend(name string, convID string, opts kbchat.RunOptions) (BotBackend, error) {
	if convID == "" {
		return nil, fmt.Errorf("KEYBASE_CHAT_CONVID is not set")
	}
	var err error
	bot := &KeybaseChatBotBackend{
		convID: chat1.ConvIDStr(convID),
		name:   name,
	}
	if bot.kbc, err = kbchat.Start(opts); err != nil {
		return nil, err
	}
	return bot, nil
}

// KeybaseRunOptionsFromEnv returns Keybase chat options configured from the
// environment shared by the bot binaries.
func KeybaseRunOptionsFromEnv(debugTag string) kbchat.RunOptions {
	opts := kbchat.RunOptions{
		KeybaseLocation: os.Getenv("KEYBASE_LOCATION"),
		HomeDir:         os.Getenv("KEYBASE_HOME"),
		DebugTag:        debugTag,
	}
	username := os.Getenv("KEYBASE_ONESHOT_USERNAME")
	paperKey := os.Getenv("KEYBASE_ONESHOT_PAPERKEY")
	if username != "" && paperKey != "" {
		opts.Oneshot = &kbchat.OneshotOptions{
			Username: username,
			PaperKey: paperKey,
		}
	}
	return opts
}

func (b *KeybaseChatBotBackend) SendMessage(text string, convID string) {
	if chat1.ConvIDStr(convID) != b.convID {
		// bail out if not on configured conv ID
		log.Printf("SendMessage: refusing to send on non-configured convID: %s != %s\n", convID, b.convID)
		return
	}
	if len(text) == 0 {
		log.Printf("SendMessage: skipping blank message")
		return
	}
	log.Printf("sending message: convID: %s text: %s", convID, text)
	if _, err := b.kbc.SendMessageByConvID(chat1.ConvIDStr(convID), "%s", text); err != nil {
		log.Printf("SendMessage: failed to send: %s\n", err)
	}
}

func (b *KeybaseChatBotBackend) SendAttachment(filename, title, convID string) error {
	if chat1.ConvIDStr(convID) != b.convID {
		return fmt.Errorf("refusing to send attachment on non-configured convID: %s != %s", convID, b.convID)
	}
	_, err := b.kbc.SendAttachmentByConvID(b.convID, filename, title)
	return err
}

func (b *KeybaseChatBotBackend) AdvertiseCommands(commands []chat1.UserBotCommandInput) error {
	if b.convID == "" {
		return nil
	}
	_, err := b.kbc.AdvertiseCommands(kbchat.Advertisement{
		Alias: b.name,
		Advertisements: []chat1.AdvertiseCommandAPIParam{{
			Typ:      "conv",
			Commands: commands,
			ConvID:   b.convID,
		}},
	})
	return err
}

func (b *KeybaseChatBotBackend) Listen(runner BotCommandRunner) {
	sub, err := b.kbc.ListenForNewTextMessages()
	if err != nil {
		log.Printf("failed to set up listen: %s", err)
		return
	}
	commandPrefix := "!" + b.name
	for {
		msg, err := sub.Read()
		if err != nil {
			log.Printf("Listen: failed to read message: %s", err)
			continue
		}
		if msg.Message.Content.TypeName != "text" {
			continue
		}
		args := parseInput(msg.Message.Content.Text.Body)
		if len(args) > 0 && args[0] == commandPrefix && b.convID == msg.Message.ConvID {
			cmd := args[1:]
			if err := runner.RunCommand(cmd, string(b.convID)); err != nil {
				log.Printf("unable to run command: %s", err)
			}
		}
	}
}
