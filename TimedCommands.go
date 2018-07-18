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

func LoadTimedCommands() map[string]time.Duration {
	database := InitializeDB()

	rows, _ := database.Query("SELECT TimedResponse, Timer from timedcommands")

	com := make(map[string]time.Duration)
	for rows.Next() {
		var TimedResponse string
		var Timer time.Duration
		rows.Scan(&TimedResponse, &Timer)
		com[TimedResponse] = Timer
	}
	return com
}

func TimedCommands(conn net.Conn, channel string, name string) {
	timedcoms := LoadTimedCommands()
	for k, v := range timedcoms {
		go func(conn net.Conn, channel, name, k string, v time.Duration) {
			for range time.NewTicker(v * time.Second).C {
				BotSendMsg(conn, channel, k, name)
			}
		}(conn, channel, name, k, v)
	}
}
