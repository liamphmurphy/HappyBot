package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"time"
)

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

func RunPoints(timer time.Duration, modifier int, conn net.Conn, channel string) {
	database := InitializeDB()
	var Points int
	for range time.NewTicker(timer * time.Second).C {
		currentUsers := GetViewers(conn, channel)
		tx, err := database.Begin()
		if err != nil {
			fmt.Println("Error starting points transaction: ", err)
		}

		for _, v := range currentUsers.Chatters.CurrentViewers {
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
	}

}
