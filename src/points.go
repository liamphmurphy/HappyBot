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

	pointsMap := make(map[string]int)

	for {
		tx, err := database.Begin()
		if err != nil {
			fmt.Println("Error starting points transaction: ", err)
		}

		statement, _ := tx.Prepare("INSERT INTO points (Username, Points) VALUES (?, ?)")
		for range time.NewTicker(5 * time.Second).C {
			points++
			for _, v := range currentUsers.Chatters.CurrentViewers {
				pointsMap[v] = points
				statement.Exec(v, points)

			}

		}
		for range time.NewTicker(15 * time.Second).C {
			tx.Commit()
		}
	}

}
