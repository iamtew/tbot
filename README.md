# the start of tbot
___
*tbot - tew's bot (because naming things is hard!)*

IRC bot designed and implemented by @iamtew, for him and his friends. It suits his needs, maybe not yours!
___

## Functionality of tbot

IRC bot connects to server and...
- uses TOML file for configuration
- bot connect to single IRC network
- configuration only supports one IRC network
- configuration includes:
- - connection settings to IRC network
- - bot name and details
- - bot admins identified by `nickname!username@example.net` user mask
- - array of `barrels` that's enabled and disabled. default state disabled, if a barrel isn't listed it will be disabled
- bot answers to commands in public chat
- commands will be written in chat, prefixed by a special character
- command character can be configured, default is a period .
- public chat commands can be restricted by various levels of users, such as op, halfop, voice, normal
- bot answers to admins in private chat, for admin tasks
- bot has multiple commands that can be executed by admins in private chat
- commands available in base bot:
- - reload, which reloads configuration
- - barrel, list status of available barrels, and enable/disable barrels at runtime
- - get, get configuration information about the bot
- - set, set various configuration changes at runtime
- - write, write running configuration to disk
- - reconnect, reconnect to the IRC network, don't exit process
- - stop/shutdown, stop the bot and exit process
- extensive logging of bot events and IRC events
- logging level can be confiured in configuration file and also at runtime 
- logging can be written to disk if configured
- `barrels` or `barrels of fun` are submodules that extend the bots
- barrels can: 
- - can monitor channel and react with actions to patterns, regex
- - can add new commands to be available in public channels, or private chat
- - can be enabled/disabled at runtime
- - can be configured in the config file
- tbot is designed to be run from the command line
- tbot works on both linux and windows
- tbot is written in Golang 
- tbot is designed to be installed on a system in the path
- tbot will use the configuration file specified on the command line, example: `tbot ~/path/to/tbot.toml`
- tbot will not start without a configuration specified, unless the write config template option is true
- tbot will print all events of what is happening in terminal, configured by the log level option, and can be disabled with the quiet option
- tbot command line options are:
- - -h,--help this help page
- - -e,--example write example config on path specified and exit
- - -D,--daemon run tbot in background, detach from shell, also implies quiet option true
- - -L,--loglevel debug, verbose, info
- - -Q,--quiet no output when running
- - -v,--verbose verbose logging, alias of --loglevel=verbose
- - -d,--debug debug logging, alias of --loglevel=debug

## Barrels 

the following barrels are part of the standard library of barrels
the barrels live in a subdirectory or something.
- barrel.url, listens to messages in a channel and look up the title of the webpage every time someone post a http or https link
- barrel.url, when someone types `more` after a url has been resolved, more details about the web page should be displayed.
- barrel.fish, adds the command fish, which when executed makes the bot write out a joke about fish.
- barrel.fish, command parses any argument for nicknames in channel and writes the joke about them.
- barrel.fish, has an extensive library of fish jokes in an array or list in the code.