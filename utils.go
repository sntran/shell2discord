package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// parseBotCommand - parse command-line arguments for one bot command
func parseBotCommand(slashCommand string, shellCommand string) (commandName string, description string, params map[string]string) {
	commandRe := regexp.MustCompile(`^/([\w-]{1,32})$`)
	commandMatches := commandRe.FindStringSubmatch(slashCommand)

	if commandMatches == nil {
		log.Fatalf("error: invalid command %s", slashCommand)
	}
	commandName = commandMatches[1]
	description = commandName + " command"

	descriptionRe := regexp.MustCompile(`#([^\\]+)\\`)
	descriptionMatch := descriptionRe.FindStringSubmatch(shellCommand)

	if len(descriptionMatch) > 0 && descriptionMatch[1] != "" {
		description = strings.TrimSpace(descriptionMatch[1])
	}

	// Parse variable with optional default value, `${foo-bar}`
	paramsRe := regexp.MustCompile(`\${(\w+)(-[^}]*)?}`)
	paramMatches := paramsRe.FindAllStringSubmatch(shellCommand, -1)
	matchesLen := len(paramMatches)

	params = map[string]string{}
	for i := 0; i < matchesLen; i++ {
		name := paramMatches[i][1]
		value := paramMatches[i][2]

		if _, ok := params[name]; !ok {
			params[name] = value
		}
	}

	return commandName, description, params
}

// Executes a shell command.
func execShellCommand(shellCommand string, envVars []string) (shellOut []byte, err error) {
	ctx := context.Background()
	osExecCommand := exec.CommandContext(ctx, *ShellBinary, "-c", shellCommand)
	osExecCommand.Stderr = os.Stderr

	for i := 0; i < len(envVars); i++ {
		envVar := envVars[i]
		osExecCommand.Env = append(
			osExecCommand.Env,
			fmt.Sprintf("%s=%s", envVar, os.Getenv(envVar)),
		)
	}

	return osExecCommand.Output()
}
