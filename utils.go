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
func parseBotCommand(slashCommand string, shellCommand string) (commandName string, params []string) {
	commandRe := regexp.MustCompile(`^/([\w-]{1,32})$`)
	commandMatches := commandRe.FindStringSubmatch(slashCommand)

	if commandMatches == nil {
		log.Fatalf("error: invalid command %s", slashCommand)
	}
	commandName = commandMatches[1]

	paramsRe := regexp.MustCompile(`\${(\w+)}`)
	matches := paramsRe.FindAllStringSubmatch(shellCommand, -1)
	matchesLen := len(matches)

	paramsMap := map[string]bool{}

	for i := 0; i < matchesLen; i++ {
		param := matches[i][1]
		if !paramsMap[param] {
			params = append(params, param)
			paramsMap[param] = true
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

func chunks(s string, chunkSize int) []string {
	if chunkSize >= len(s) {
		return []string{s}
	}
	var chunks []string
	chunk := make([]rune, chunkSize)
	len := 0
	for _, r := range s {
		chunk[len] = r
		len++
		if len == chunkSize {
			chunks = append(chunks, string(chunk))
			len = 0
		}
	}
	if len > 0 {
		chunks = append(chunks, string(chunk[:len]))
	}
	return chunks
}
