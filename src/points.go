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
	var x int
	statement, _ := database.Prepare("INSERT INTO points (Username, Points) VALUES (?, ?)")
	pointsMap := make(map[string]int)

	fmt.Println(currentUsers.Chatters.CurrentViewers)
	for {
		for range time.NewTicker(5 * time.Second).C {
			points++
			x++

			for _, v := range currentUsers.Chatters.CurrentViewers {
				pointsMap[v] = points

				if x > 9 {
					statement.Exec(v, points)
					x = 0
				}

			}
		}
	}
}
