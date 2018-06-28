package main

import (
	"net"
	"time"
)

func MakeTimedCommand(response string, timer time.Duration) *CustomTimedCommand {
	return &CustomTimedCommand{
		TimedResponse: response,
		Timer:         timer,
	}

}

func LoadTimedCommands() map[string]*CustomTimedCommand {
	database := InitializeDB()

	rows, _ := database.Query("SELECT TimedName, TimedResponse, Timer from timedcommands")

	com := make(map[string]*CustomTimedCommand)
	for rows.Next() {
		var TimedName, TimedResponse string
		var Timer time.Duration
		rows.Scan(&TimedName, &TimedResponse, &Timer)
		com[TimedResponse] = MakeTimedCommand(TimedResponse, Timer)
	}
	return com
}

func TimedCommands(conn net.Conn, channel string, name string) {
	timedcoms := LoadTimedCommands()
	for _, v := range timedcoms {
		for range time.NewTicker(v.Timer * time.Second).C {
			BotSendMsg(conn, channel, v.TimedResponse, name)
		}
	}
}
