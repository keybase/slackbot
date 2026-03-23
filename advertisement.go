package slackbot

import (
	"fmt"
	"slices"
	"strings"

	"github.com/keybase/go-keybase-chat-bot/kbchat/types/chat1"
)

type commandAdvertiser interface {
	AdvertiseCommands(commands []chat1.UserBotCommandInput) error
}

func (b *Bot) AddAdvertisements(commands ...chat1.UserBotCommandInput) {
	b.advertisements = append(b.advertisements, commands...)
}

func (b *Bot) AdvertisedCommands() []chat1.UserBotCommandInput {
	commands := []chat1.UserBotCommandInput{{
		Name:                "help",
		Description:         "Show available commands",
		Usage:               fmt.Sprintf("!%s help", b.name),
		ExtendedDescription: b.helpExtendedDescription(),
	}}

	for _, trigger := range b.triggers() {
		command := b.commands[trigger]
		commands = append(commands, chat1.UserBotCommandInput{
			Name:        trigger,
			Description: command.Description(),
			Usage:       fmt.Sprintf("!%s %s", b.name, trigger),
		})
	}

	extras := slices.Clone(b.advertisements)
	slices.SortFunc(extras, func(a, b chat1.UserBotCommandInput) int {
		return strings.Compare(a.Name, b.Name)
	})
	commands = append(commands, extras...)

	return commands
}

func (b *Bot) advertiseCommands() error {
	advertiser, ok := b.backend.(commandAdvertiser)
	if !ok {
		return nil
	}
	return advertiser.AdvertiseCommands(b.AdvertisedCommands())
}

func (b *Bot) helpExtendedDescription() *chat1.UserBotExtendedDescription {
	help := strings.TrimSpace(b.resolvedHelp())
	if help == "" {
		return nil
	}
	return &chat1.UserBotExtendedDescription{
		Title:       fmt.Sprintf("%s help", b.name),
		DesktopBody: help,
		MobileBody:  help,
	}
}
