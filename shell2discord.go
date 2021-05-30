package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
)

const (
	Version = "0.1"
)

// Bot parameters
var (
	GuildID        = flag.String("guild", os.Getenv("DISCORD_GUILD"), "Test guild ID. If not passed - bot registers commands globally")
	ChannelIDs     = flag.String("channels", os.Getenv("DISCORD_CHANNELS"), `Discord channels that are allowed to use the bot ("channel1,channel2")`)
	BotToken       = flag.String("token", os.Getenv("DISCORD_TOKEN"), "Bot access token")
	ExportVars     = flag.String("export-vars", "", `export environment vars to shell command ("VAR1,VAR2,...")`)
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
)

var session *discordgo.Session

var commands []*discordgo.ApplicationCommand
var commandHandlers map[string]string = make(map[string]string)

// Parse arguments
func init() {
	flag.Parse()

	// need >= 2 arguments and count of it must be even
	args := flag.Args()

	if len(args) < 2 || len(args)%2 == 1 {
		log.Fatalf("error: need pairs of shell-command")
	}

	for i := 0; i < len(args); i += 2 {
		shellCommand := args[i+1]
		commandName, params := parseBotCommand(args[i], shellCommand)
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

		commandHandlers[commandName] = shellCommand
		command := &discordgo.ApplicationCommand{
			Name: commandName,
			// @TODO: Parse command description from flag.
			Description: "Test command",
			Options:     options,
		}
		commands = append(commands, command)
	}
}

// Starts bot
func init() {
	var err error
	session, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

// Add command handlers
func init() {
	session.AddHandler(func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {

		if *ChannelIDs != "" && !strings.Contains(*ChannelIDs, interaction.ChannelID) {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionApplicationCommandResponseData{
					Content: "Unauthorized",
				},
			})

			return
		}

		var interactionName = interaction.Data.Name
		var options = interaction.Data.Options

		if shellCmd, ok := commandHandlers[interactionName]; ok {
			for i := 0; i < len(options); i++ {
				option := options[i]
				variable := fmt.Sprintf("${%s}", option.Name)
				shellCmd = strings.Replace(shellCmd, variable, option.StringValue(), -1)
			}

			log.Println("Executing:", shellCmd)

			ctx := context.Background()
			osExecCommand := exec.CommandContext(ctx, "sh", "-c", shellCmd)
			osExecCommand.Stderr = os.Stderr

			// Passthrough environment variables set from `--export-vars` to the shell.
			enVarRe := regexp.MustCompile(`,`)
			envVars := enVarRe.Split(*ExportVars, -1)
			for i := 0; i < len(envVars); i++ {
				envVar := envVars[i]
				osExecCommand.Env = append(
					osExecCommand.Env,
					fmt.Sprintf("%s=%s", envVar, os.Getenv(envVar)),
				)
			}

			var reply string
			shellOut, err := osExecCommand.Output()
			if err != nil {
				reply = fmt.Sprintf("exec error: %s", err)
			} else {
				reply = string(shellOut)
			}

			if reply == "" {
				reply = "Command Executed"
			}

			// Reply with shell command output.
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionApplicationCommandResponseData{
					Content: reply,
				},
			})
		}
	})
}

func main() {
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is up!")
	})
	err := session.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	for _, v := range commands {
		log.Printf("Adding command %v", v)
		command, err := session.ApplicationCommandCreate(session.State.User.ID, *GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%s' command: %v", v.Name, err)
		}
		log.Printf("Command /%s added", v.Name)
		// Store the command ID so we can remove it on exit.
		v.ID = command.ID
	}

	defer session.Close()

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Gracefully shutting down")

	if *RemoveCommands {
		log.Println("Removing commands")
		for _, v := range commands {
			if err = session.ApplicationCommandDelete(session.State.User.ID, *GuildID, v.ID); err != nil {
				log.Printf("Could not delete '%s' command: %v", v.Name, err)
			}
		}
	}
}
