package main

import (
	"fmt"
	"net"
	"time"
)

func RunPoints(conn net.Conn, channel string) {
	database := InitializeDB()
	currentUsers := GetViewers(conn, channel)
	var points int

	statement, _ := database.Prepare("INSERT INTO points (Username, Points) VALUES (?, ?)")

	defer database.Close()
	fmt.Println(currentUsers.Chatters.CurrentViewers)
	for range time.Tick(2 * time.Second) {
		points++
		for _, v := range currentUsers.Chatters.CurrentViewers {
			fmt.Println(v, points)

			statement.Exec(v, points)

		}
	}

}
