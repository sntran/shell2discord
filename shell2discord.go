package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
)

const (
	Version = "0.1"
)

// Bot parameters
var (
	GuildID        = flag.String("guild", os.Getenv("DISCORD_GUILD"), "Test guild ID. If not passed - bot registers commands globally")
	BotToken       = flag.String("token", os.Getenv("DISCORD_TOKEN"), "Bot access token")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
)

var s *discordgo.Session

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
		commandName, shellCmd := parseBotCommand(args[i], args[i+1])
		commandHandlers[commandName] = shellCmd
		command := &discordgo.ApplicationCommand{
			Name:        commandName,
			Description: "Test command",
		}
		commands = append(commands, command)
	}
}

// Starts bot
func init() {
	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

// Add command handlers
func init() {
	s.AddHandler(func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		var interactionName = interaction.Data.Name
		if shellCmd, ok := commandHandlers[interactionName]; ok {
			log.Printf("/%s: `%s", interactionName, shellCmd)
		}
	})
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is up!")
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	for _, v := range commands {
		log.Printf("Adding command %v", v)
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, *GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%s' command: %v", v.Name, err)
		}
		log.Printf("Command /%s added", v.Name)
		// Store the command ID so we can remove it on exit.
		v.ID = cmd.ID
	}

	defer s.Close()

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Gracefully shutting down")

	if *RemoveCommands {
		log.Println("Removing commands")
		for _, v := range commands {
			if err = s.ApplicationCommandDelete(s.State.User.ID, *GuildID, v.ID); err != nil {
				log.Printf("Could not delete '%s' command: %v", v.Name, err)
			}
		}
	}
}
