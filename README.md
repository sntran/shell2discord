# shell2discord

Create Discord splash commands from command-line.

## Disclaimer

This program is written by a developer first-time using Go. Bugs and smells are expected.
Pull requests are welcome to help make the program and its author better.

Golang is not the developer's language of choice, but due to the nature of this program,
the ability to have cross-platform static binary from Go makes it a suitable choice.

## Usage

Get a Discord token from Developer portal, and set `DISCORD_TOKEN` environment variable in shell.
Optionally, set `DISCORD_GUILD` to a specific GUILD to make the commands only available to that GUILD.
Otherwise, they are registered globally, and may take up to an hour to see changes.

If the bot is only allowed to be used in certain channels, specify their IDs in `--channels` flag,
separated by comma. Commands sent from outside those channels will get a "Unauthorized" reply.

```shell
export DISCORD_TOKEN=*******
export DISCORD_GUILD=*******
shell2discord [options] /slash_command1 'shell command' /slash_command2 'shell command2'...
options:
    --token=<TOKEN> : Discord token
    --guild=<GUILD_ID>: Discord GUILD ID to limit commands to.
    --channels=<CHANNEL_ID>,<CHANNEL_ID>: Comma-separated list of channel IDs in which the bot can be used.
    --export-vars=<VAR1>,<VAR2>: Comma-separated list of environment variables to pass to the shell command.
    --shell=<SHELL_BINARY_NAME>: Name of the shell binary to execute command with via the -c flag. Defaults to 'sh'.
```

## Shell commands

The shell command argument for the slash command should be written as if it would be run in a normal shell,
except that it must be wrapped in a pair single quotes.

That is:

- Variables are enclosed in `${}`. They become required options.
- Optional variable with default value as `${variable-default}`.
- `default` can be an empty string, and the variable is still optional.
- Description is written as the first comment line.

## Examples

```shell
# Command with no param.
shell2discord /hello 'echo "World"'
# Command with description
shell2discord /hello '# Command description here \
    echo "World"'
# Optional param `word` with empty default value.
shell2discord /hello 'echo "Hello ${word-}"'
# Optional param `word` with default value of "World".
shell2discord /hello 'echo "Hello ${word-World}"'
# Command with required params.
shell2discord /mirror 'curl "${url}" > "${outfile}"'
# Command with allowed environment variables.
shell2discord --export-vars=GOOGLE_MAPS_API_KEY \
    /geocode 'curl "https://maps.googleapis.com/maps/api/geocode/json?latlng=${latlng}&key=$GOOGLE_MAPS_API_KEY"'
```
