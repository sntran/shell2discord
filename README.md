# shell2discord

Create Discord splash commands from command-line.

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
```

## Examples

```shell
shell2discord /hello 'echo "World"'
shell2discord /mirror 'curl "${url}" > "${outfile}"'
```
