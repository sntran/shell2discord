package main

import (
	"log"
	"regexp"
)

// parseBotCommand - parse command-line arguments for one bot command
func parseBotCommand(slashCommand, shellCommand string) (commandName string, command string) {
	if len(slashCommand) == 0 || slashCommand[0] != '/' {
		log.Fatalf("error: path %s doesn't start with /", slashCommand)
	}
	if stringIsEmpty(shellCommand) {
		log.Fatalf("error: shell command cannot be empty")
	}

	// Substring after "/" is the commandName.
	runes := []rune(slashCommand)
	commandName = string(runes[1:])

	// @TODO: check against Discord's `^[\w-]{1,32}$`.

	// @TODO: parse command arguments from shellCommand

	return commandName, shellCommand
}

// stringIsEmpty - check string is empty
func stringIsEmpty(str string) bool {
	isEmpty, _ := regexp.MatchString(`^\s*$`, str)
	return isEmpty
}
