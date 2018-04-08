This is a simple Twitch bot written in Golang.

Just a little side-project because I got interested in Go and wanted to try it out by recreating a project I had done before.
It seems to run very fast, and can honestly be used in production for some very basic botting.

The name 'HappyBot' came to be because I've spent a fair bit of time watching the BobRoss Twitch stream while working on this bot.

<h1> Why make HappyBot? </h1>
Primary reason was for practice and working on my programming abilities, but I did have a few goals in mind.

- Make it fast. It is designed to be a command line only program, which means that it barely takes up any CPU or RAM resources.

- Make it run anywhere. Similar to 'make it fast', but I want to make it so that streamers can run it easily either on their machines or elsewhere. Want to put it on a server? Do it. Run it on a Raspberry Pi? Absolutely. 

- Make it configurable. If a bot can't be configured to a user's taste, it isn't a good bot in my mind. Users should be able to disable link checking, change the amount of characters before being purged for a long message, add commands easily etc.

- Make it crossplatform. Having it only run on one OS is not very configurable after all; so I wanted to avoid a language and design choices that favored one platform over another. 

- Make it open source. Open source is cool: it benefits everybody. Feel free to contribute changes and critique my code (I'm a newbie after all). 

<h2> Next steps before a proper release </h2>

- Take sqlite3 data and convert it into a map, primarily for commands which has multiple fields in the database. 

- Add a quote system. This will be very easy to do once commands work properly.


<h1> To run it </h1>

`go get github.com/BurntSushi/toml`

Then

`go run bot.go`

<h1> Building it </h1>

Though you can just keep using go run, if you want an easy executable, you can build it as well.

`go build bot.go`

<b> This will build a executable based off of your OS. Exe for Windows, sh for Linux etc.
