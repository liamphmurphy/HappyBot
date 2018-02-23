This is a simple Twitch bot written in Golang.

Just a little side-project because I got interested in Go and wanted to try it out by recreating a project I had done before.
It seems to run very fast, and can honestly be used in production for some very basic botting.

The name 'HappyBot' came to be because I've spent a fair bit of time watching the BobRoss Twitch stream while working on this bot.

Main things to be worked on is the capability of adding commands from chat, and have the bot update the list of commands when a new one is added.

Edit config.toml with the information needed then execute the below commands.

<h1> To run it </h1>

`go get github.com/BurntSushi/toml`

Then

`go run bot.go`

<h1> Building it </h1>

Though you can just keep using go run, if you want an easy executable, you can build it as well.

`go build bot.go`

<b> This will build a executable based off of your OS. Exe for Windows, sh for Linux etc.
