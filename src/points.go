package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"time"
)

func GetUserPoints(username string) int {
	database := InitializeDB()
	var Points int
	err := database.QueryRow("SELECT Points FROM points WHERE Username = ?", username).Scan(&Points)
	if err != nil {
		return 0
	}
	return Points

}

func UpdateUserPoints(username string, points int) {
	database := InitializeDB()
	statement, _ := database.Prepare("UPDATE points SET Points = ? WHERE username = ?")
	statement.Exec(points, username)
}

func UserInDB(db *sql.DB, username string) bool {
	checkUser := "SELECT Username FROM points WHERE Username = ?"
	err := db.QueryRow(checkUser, username).Scan(&username)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Print(err)
		}
		return false
	}
	return true
}

func GivePoints(db *sql.DB, username string, amount int) {
	currentPoints := GetUserPoints(username)
	newAmount := currentPoints + amount
	UpdateUserPoints(username, newAmount)
}

func RunPoints(timer time.Duration, modifier int, conn net.Conn, channel string) {
	database := InitializeDB()
	var Points int
	var allUsers []string
	for range time.NewTicker(timer * time.Second).C {
		currentUsers := GetViewers(conn, channel)
		tx, err := database.Begin()
		if err != nil {
			fmt.Println("Error starting points transaction: ", err)
		}

		allUsers = append(allUsers, currentUsers.Chatters.CurrentViewers...)
		allUsers = append(allUsers, currentUsers.Chatters.CurrentModerators...)

		currentUsers.Chatters.CurrentViewers = currentUsers.Chatters.CurrentViewers[:0]
		currentUsers.Chatters.CurrentModerators = currentUsers.Chatters.CurrentModerators[:0]

		for _, v := range allUsers {
			userCheck := UserInDB(database, v)
			if userCheck == false {
				statement, _ := tx.Prepare("INSERT INTO points (Username, Points) VALUES (?, ?)")
				statement.Exec(v, 1)
			} else {

				err = tx.QueryRow("Select Points FROM points WHERE Username = ?", v).Scan(&Points)
				if err != nil {

				} else {
					Points = Points + modifier
					statement, _ := tx.Prepare("UPDATE points SET Points = ? WHERE username = ?")
					statement.Exec(Points, v)
				}
			}
		}
		tx.Commit()
		allUsers = allUsers[:0]

	}

}
