package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Command struct {
	*discordgo.ApplicationCommand
	Script string
	Env    []string
}

// Creates a new command from `/command 'shell command'`.
func NewCommand(slashCommand string, shellCommand string) *Command {
	commandName, params := parseBotCommand(slashCommand, shellCommand)
	paramsLen := len(params)

	options := make([]*discordgo.ApplicationCommandOption, paramsLen)
	for i := 0; i < paramsLen; i++ {
		options[i] = &discordgo.ApplicationCommandOption{
			// Shell variables have no type, so we just use String in Discord.
			Type: discordgo.ApplicationCommandOptionString,
			Name: params[i],
			// @TODO: Parse option description from flag.
			Description: params[i],
			Required:    true,
		}
	}

	return &Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        commandName,
			Description: commandName + " command",
			Options:     options,
		},
		Script: shellCommand,
	}
}

// Executes a command based on Discord's interaction to it.
func (command Command) Exec(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionApplicationCommandResponseData{
			// Makes this response ephemeral—only the invoking user can see it
			Flags:   1 << 6,
			Content: "Thinking...",
		},
	})

	var options = interaction.Data.Options

	shellCommand := command.Script
	for i := 0; i < len(options); i++ {
		option := options[i]
		variable := fmt.Sprintf("${%s}", option.Name)
		shellCommand = strings.Replace(shellCommand, variable, option.StringValue(), -1)
	}

	var reply string
	// Execute shell command, passing through environment variables.
	shellOut, err := execShellCommand(shellCommand, command.Env)

	if err != nil {
		reply = fmt.Sprintf("exec error: %s", err)
	} else {
		reply = string(shellOut)
	}

	if reply == "" {
		reply = "Command Executed"
	}

	if len(reply) > 2000 {
		fileName := command.Name

		for i := 0; i < len(options); i++ {
			option := options[i]
			value := option.StringValue()
			if value != "" {
				fileName += "_" + option.Name + "_" + value
			}
		}

		currentTime := time.Now()
		fileName += "_" + currentTime.Format("2006-01-02T15:04:05") + ".txt"

		_, err = session.FollowupMessageCreate(session.State.User.ID, interaction.Interaction, true, &discordgo.WebhookParams{
			Files: []*discordgo.File{
				{
					Name:   fileName,
					Reader: strings.NewReader(reply),
				},
			},
		})
	} else {
		_, err = session.FollowupMessageCreate(session.State.User.ID, interaction.Interaction, true, &discordgo.WebhookParams{
			Content: reply,
		})
	}

	if err != nil {
		session.FollowupMessageCreate(session.State.User.ID, interaction.Interaction, true, &discordgo.WebhookParams{
			Content: "Something went wrong",
		})
		return
	}
}
