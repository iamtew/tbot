# the start of tbot
___
*This README explains how to configure, run, and extend tbot.*
*tbot - tew's bot (because naming things is hard!)*

IRC bot designed and implemented by @iamtew, for him and his friends. It suits his needs, maybe not yours!
___

## Functionality of tbot

tbot is a compact, eyes-on IRC bot for people who want command-driven control and fast extensibility.

### What makes tbot special

- Lightweight, single-network IRC client
- Configured with a human-readable TOML file
- Admin-safe private control via IRC usermasks
- Modular barrel architecture for feature plugins
- Built-in logging and runtime management

### Core capabilities

| Feature | Description |
|---|---|
| Single-network connection | One IRC server, one nick, multiple channels |
| Config file | TOML-based configuration for network, bot, admins, barrel |
| Admin auth | Exact or wildcard usermask authorization for private admin commands |
| Public commands | Command prefix configurable; default is `.` |
| Runtime control | Reload config, write config, reconnect, stop without killing the bot |
| Logging | Console output with optional disk log capture |
| Barrel plugins | Enable or disable submodules live without restarting |

### Admin command list

Admins can talk to tbot in private chat and run commands like:

- `help` / `commands` — list available admin commands
- `reload` — reload the configuration file at runtime
- `barrel list` — show available barrel status
- `barrel enable <name>` / `barrel disable <name>` — toggle features live
- `get <config.key>` — inspect current bot settings
- `get <config.*>` — search config keys using wildcard patterns
- `set <config.key> <value>` — change runtime configuration
- `write` — save the running configuration back to disk
- `reconnect` — reconnect to IRC without exiting the process
- `stop` / `shutdown` — shut the bot down cleanly

### Barrel system

tbot uses "barrels" to package optional behavior:

- `url` and `fish` are included by default
- Barrels can inspect channel messages and respond automatically
- Barrels can add commands to public chat or private admin chat
- Barrels can be configured and toggled from the TOML file using `barrel.*` sections

## Getting started

### Windows developer setup

If you are developing on Windows, install a UNIX-compatible build shell and make utilities before using `make` or building from the repository.

Recommended setup:

- Install MSYS2 with `winget install MSYS2.MSYS2`
- Open the MSYS2 shell
- Run `pacman -Syu` to update the package database
- Install build tools with `pacman -S base-devel`

Once MSYS2 and development tools are installed, you can use the provided `Makefile` on Windows.

Build the bot:

- `go build`

Generate a starter config:

- `tbot -e tbot.example.toml`

Run with your config:

- `tbot ./tbot.toml`

Set a custom pid file if you want to manage the running service explicitly:

- `tbot -P ./tbot.pid ./tbot.toml`

Stop a running bot using the pid file:

- `tbot -S -P ./tbot.pid`

> By default, the pid file is written next to the config file using the same name and a `.pid` extension. Public commands use the configured prefix (default `.`). Admin commands must be sent as private messages from a configured admin mask. Admin masks support simple wildcards like `tew!~tew@*.user.oftc.net`.

## Barrels

tbot ships with a compact standard barrel library that is designed to be extended.

| Barrel | Purpose | Commands |
|---|---|---|
| `url` | Detects links in chat, resolves page titles, and stores the last URL details per channel | `more` |
| `fish` | Adds a playful fish joke command to lighten the mood | `fish [nick ...]` |

### Barrel behavior

`url` barrel
- Watches channel messages for `http://` or `https://` links
- Resolves each unique URL once per configured cooldown interval
- Fetches the page title and posts it back to chat
- Supports `more` for extra metadata on the latest URL

`fish` barrel
- Adds a playful `fish` command
- Replies with a random fish joke
- Can mention one or more nicknames in the response

## Command line options

- `-h`, `--help` — show help text
- `-e`, `--example` — write an example config file and exit
- `-D`, `--daemon` — run quietly in the background
- `-L`, `--loglevel` — `debug`, `verbose`, `info`, `warn`, `error`
- `-Q`, `--quiet` — suppress runtime output
- `-P`, `--pidfile` — set the PID file path explicitly
- `-S`, `--stop` — send a stop signal to the bot referenced by the PID file
- `-V`, `--version` — display the current version number and exit
- `-v`, `--verbose` — alias for `--loglevel=verbose`
- `-d`, `--debug` — alias for `--loglevel=debug`

The PID file can also be specified in the config under `[bot] pidfile = "..."`, otherwise it defaults to the config file name with `.pid` extension.

## Why tbot?

tbot is designed to be a practical, no-nonsense IRC bot with strong runtime control and a simple plugin model. It’s ideal for small teams who want an IRC assistant that can be managed live and extended with new barrel behavior over time.

Use it when you want:

- fast setup with `go build`
- secure admin control via usermasks
- live feature toggling without restart
- a compact bot that still feels powerful