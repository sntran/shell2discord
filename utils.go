package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
)

// parseBotCommand - parse command-line arguments for one bot command
func parseBotCommand(slashCommand string, shellCommand string) (commandName string, params map[string]string) {
	commandRe := regexp.MustCompile(`^/([\w-]{1,32})$`)
	commandMatches := commandRe.FindStringSubmatch(slashCommand)

	if commandMatches == nil {
		log.Fatalf("error: invalid command %s", slashCommand)
	}
	commandName = commandMatches[1]

	// Parse variable with optional default value, `${foo-bar}`
	paramsRe := regexp.MustCompile(`\${(\w+)(-\w*)?}`)
	matches := paramsRe.FindAllStringSubmatch(shellCommand, -1)
	matchesLen := len(matches)

	params = map[string]string{}
	for i := 0; i < matchesLen; i++ {
		name := matches[i][1]
		value := matches[i][2]

		if _, ok := params[name]; !ok {
			params[name] = value
		}
	}

	return commandName, params
}

// Executes a shell command.
func execShellCommand(shellCommand string, envVars []string) (shellOut []byte, err error) {
	ctx := context.Background()
	osExecCommand := exec.CommandContext(ctx, "sh", "-c", shellCommand)
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
