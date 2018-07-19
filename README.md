<h1> A word of warning </h1>
On twitch, most bots do not have special privileges in regards to rules on the website. If you setup your bot using this software in such a way that is against TOS and/or gets your accounts banned, me or any of the other contributors are not responsible. The bot serves a specific purpose, and it is up to the user on how to configure it. 

Thanks :)

<h1> How-to / Configuration </h1>

All configuration is done in the config.toml file in the config folder. Edit those values to configure the bot. 
 
Primary reason was for practice and working on my programming abilities, but I did have a few goals in mind.
You can use notepad or any other text editor you are familiar with.

Below is an example config file that is filled out.

    ChannelName = "#lphm"
    ServerName = "irc.chat.twitch.tv:6667"
    # Oauth from Twitch here
    BotOAuth = "oauth token"
    BotName = "MrDestructoid"
    ##### CHAT PROPERTIES #####
    LetModeratorsUseAllCommands = true # This allows a moderator to use any command, including broadcaster specific commands.

    CheckLongMessageCap = true
    LongMessageCap = 300

    StreamerTimeToggle = true # Toggle whether !time, !uptime is enabled or not

    MakeLog = false  # Disabled by default, logs can take up space fast depending on chat activity

    RespondToSubs = true
    SubResponse = "thanks for subscribing!"

    PurgeForLinks = true
    LinkChecks = [".com", ".net", ".org", ".tv"]

    HydrateOn = false
    HydrateTime = 30 # In minutes, how often you want to be reminded to hydrate!
    HydrateMessage = "It's hydrating time!"`

Be sure to not edit the ServerName. Keep the # in front of your ChannelName. Any value set to 'true' or 'false' can be changed to the other value based on what the user wants.

<h1> Why make HappyBot? </h1>

- Make it fast. It is designed to be a command line only program, which means that it barely takes up any CPU or RAM resources.

- Make it run anywhere. Similar to 'make it fast', but I want to make it so that streamers can run it easily either on their machines or elsewhere. Want to put it on a server? Do it. Run it on a Raspberry Pi? Absolutely. 

- Make it configurable. If a bot can't be configured to a user's taste, it isn't a good bot in my mind. Users should be able to disable link checking, change the amount of characters before being purged for a long message, add commands easily etc.

- Make it crossplatform. Having it only run on one OS is not very configurable after all; so I wanted to avoid a language and design choices that favored one platform over another. 

- Make it open source. Open source is cool: it benefits everybody. Feel free to contribute changes and critique my code (I'm a newbie after all). 


<h1> To run it </h1>

`go get github.com/BurntSushi/toml`
 
`go get github.com/mattn/go-sqlite3`

Then, on Linux / Mac OS

`go run *.go`

On Windows (previous method with wildcard was buggy, but maybe it'll work for you

`go run bot.go consoleinput.go commands.go api.go points.go timedcommands.go`

<h1> Building it </h1>

Though you can just keep using go run, if you want an easy executable, you can build it as well.

`go build bot.go consoleinput.go`

<b> This will build a executable based off of your OS. Exe for Windows, sh for Linux etc.
