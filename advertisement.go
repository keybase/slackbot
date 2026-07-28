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
		Name:                b.advertisedCommandName("help"),
		Description:         "Show available commands",
		ExtendedDescription: b.helpExtendedDescription(),
	}}

	for _, trigger := range b.triggers() {
		command := b.commands[trigger]
		commands = append(commands, chat1.UserBotCommandInput{
			Name:        b.advertisedCommandName(trigger),
			Description: command.Description(),
		})
	}

	extras := slices.Clone(b.advertisements)
	slices.SortFunc(extras, func(a, b chat1.UserBotCommandInput) int {
		return strings.Compare(a.Name, b.Name)
	})
	for i := range extras {
		extras[i].Name = b.advertisedCommandName(extras[i].Name)
	}
	commands = append(commands, extras...)

	return commands
}

func (b *Bot) advertisedCommandName(command string) string {
	return fmt.Sprintf("%s %s", b.name, command)
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
