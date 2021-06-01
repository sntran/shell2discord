package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
)

const (
	Version = "0.3"
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

var commands map[string]*Command = make(map[string]*Command)

// Parse arguments
func init() {
	flag.Parse()

	// need >= 2 arguments and count of it must be even
	args := flag.Args()

	if len(args) < 2 || len(args)%2 == 1 {
		log.Fatalf("error: need pairs of shell-command")
	}

	enVarRe := regexp.MustCompile(`,`)
	envVars := enVarRe.Split(*ExportVars, -1)

	for i := 0; i < len(args); i += 2 {
		command := NewCommand(args[i], args[i+1])
		command.Env = envVars
		commands[command.Name] = command
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
		if command, ok := commands[interactionName]; ok {
			command.Exec(session, interaction)
		}
	})
}

func main() {
	fmt.Printf("shell2discord v%s\n\n", Version)

	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is up!")
	})
	err := session.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	for _, v := range commands {
		log.Printf("Adding command %v", v)
		command, err := session.ApplicationCommandCreate(session.State.User.ID, *GuildID, v.ApplicationCommand)
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
