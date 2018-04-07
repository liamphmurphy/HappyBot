package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func ConsoleInput(conn net.Conn, channel string) {
	ConsoleScanner := bufio.NewScanner(os.Stdin)
	ConsoleScanner.Scan()
	text := ConsoleScanner.Text()

	ChatMsgCheck := strings.Contains(text, "!msg")
	if ChatMsgCheck == true {
		MsgSplit := strings.Split(text, "!msg ")
		if len(MsgSplit) <= 1 { // Len if to handle index out of range error
			fmt.Println("Please type a message.")
		} else {
			BotSendMsg(conn, channel, MsgSplit[1])
		}
	}

	ChatHelpCheck := strings.Contains(text, "!help")
	if ChatHelpCheck == true {
		fmt.Println("Current console options: !msg <text message to send to chat>")
	}

	ChatBanCheck := strings.Contains(text, "!ban")
	if ChatBanCheck == true {
		UsernameSplit := strings.Split(text, "!ban ")

		if len(UsernameSplit) <= 1 { // If len to handle index out of range error
			fmt.Println("Please type a username.")
		} else {
			BotSendMsg(conn, channel, "/ban "+UsernameSplit[1])
			fmt.Println(UsernameSplit[1] + " has been banned.")
		}
	}

	ChatUnBanCheck := strings.Contains(text, "!unban")
	if ChatUnBanCheck == true {
		UsernameSplit := strings.Split(text, "!unban ")

		if len(UsernameSplit) <= 1 { // If len to handle index out of range error
			fmt.Println("Please type a username.")
		} else {
			BotSendMsg(conn, channel, "/unban "+UsernameSplit[1])
			fmt.Println(UsernameSplit[1] + " has been unbanned.")
		}
	}

	ChatPurgeCheck := strings.Contains(text, "!purge")
	if ChatPurgeCheck == true {
		UsernameSplit := strings.Split(text, "!purge ")

		if len(UsernameSplit) <= 1 { // If len to handle index out of range error
			fmt.Println("Please type a username.")
		} else {
			ChatCommand := ("/timeout " + UsernameSplit[1] + " 1" + " Message over max character limit.")
			fmt.Println(ChatCommand)
			BotSendMsg(conn, channel, ChatCommand)
			fmt.Println(UsernameSplit[1] + " has been purged.")
		}
	}

	ChatAddBadWordCheck := strings.Contains(text, "!addbw")
	if ChatAddBadWordCheck == true {
		badwords := LoadBadWords()
		BadwordSplit := strings.Split(text, " ")
		database := InitializeDB()
		statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS badwords (BadwordID INTEGER PRIMARY KEY, Badword TEXT)")
		statement.Exec()

		if len(BadwordSplit) <= 1 { // Len if to handle index out of range error
			fmt.Println("Please type a username.")
		} else {
			database := InitializeDB()
			statement, err := database.Prepare("INSERT INTO badwords (Badword) VALUES (?)")
			if err != nil {
				fmt.Printf("Error: %s", err)
			}
			UsernameString := string(BadwordSplit[1])
			statement.Exec(UsernameString)
			// Append to the slice in this run session to make it useable right away
			badwords.BadwordSlice = append(badwords.BadwordSlice, BadwordSplit[1])
		}
	}

}
