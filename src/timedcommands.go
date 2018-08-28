package main

import (
	"fmt"
	"net"
	"strings"
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
		var timedName, timedResponse string
		var timer time.Duration
		rows.Scan(&timedName, &timedResponse, &timer)
		com[timedName] = MakeTimedCommand(timedResponse, timer)
	}
	return com
}

func TimedCommands(conn net.Conn, channel string, name string) {
	timedcoms := LoadTimedCommands()
	for _, v := range timedcoms {
		go func(conn net.Conn, channel, name, response string, timer time.Duration) {
			time.Sleep(1 * time.Millisecond)
			for range time.NewTicker(timer * time.Second).C {
				BotSendMsg(conn, channel, response, name)
			}
		}(conn, channel, name, v.TimedResponse, v.Timer)
	}
}

func TimedCommandOperations(chatmessage string) map[string]*CustomTimedCommand {
	// Create a slice of the elements in a users message
	comSplit := strings.Split(chatmessage, " ")

	// Get the key and new value for sake of database
	comKey := comSplit[1]
	comNewValue := strings.Join(comSplit[3:], " ")
	comTimer := comSplit[2]
	fmt.Println("Key: " + comKey)
	fmt.Println("Response: " + comNewValue)
	fmt.Println("Timer: " + comTimer)

	database := InitializeDB()
	if strings.Contains(chatmessage, "!edittimed") {
		rows, err := database.Prepare("UPDATE commands SET Timer = ? WHERE CommandName = ?")
		if err != nil {
			fmt.Println(err)
		}
		rows.Exec(comNewValue, comKey)
	}

	if strings.Contains(chatmessage, "!addtimed") {
		rows, err := database.Prepare("INSERT INTO timedcommands (TimedName, TimedResponse, Timer) VALUES(?,?,?)")
		if err != nil {
			fmt.Println(err)
		}
		rows.Exec(comKey, comNewValue, comTimer)
	}

	return LoadTimedCommands()

}
