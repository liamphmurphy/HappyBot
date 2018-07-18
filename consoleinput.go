package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func ConsoleInput(conn net.Conn, channel string, name string) {
	ConsoleScanner := bufio.NewScanner(os.Stdin)
	ConsoleScanner.Scan()
	text := ConsoleScanner.Text()

	if strings.Contains(text, "!dumpcommands") {
		database := InitializeDB()

		rows, _ := database.Query("SELECT CommandName, CommandResponse, CommandPermission from commands")

		file, _ := os.Create("commands.csv")
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()
		for rows.Next() {
			var CommandName, CommandResponse, CommandPermission string
			rows.Scan(&CommandName, &CommandResponse, &CommandPermission)

		}
	}

	if strings.Contains(text, "!editcom") || strings.Contains(text, "!addcom") || strings.Contains(text, "!setperm") {
		CommandOperations(text)
	}

	if strings.Contains(text, "!edittimed") || strings.Contains(text, "!addtimed") {
		TimedCommandOperations(text)
	}

	ChatMsgCheck := strings.Contains(text, "!msg")
	if ChatMsgCheck == true {
		MsgSplit := strings.Split(text, "!msg ")
		if len(MsgSplit) <= 1 { // Len if to handle index out of range error
			fmt.Println("Please type a message.")
		} else {
			BotSendMsg(conn, channel, MsgSplit[1], name)
		}
	}

	if text == "!help" {
		fmt.Println("Current console options: ")
		fmt.Println("!msg - This command followed by an text afterward will be posted by the bot in chat.")
		fmt.Println("!exit - This closes the bot safely. You may also use the 'CTRL+' shortcut.")
		fmt.Println("!ban - This command followed by a username will ban the user from your channel.")
		fmt.Println("!unban - This command followed by a username will unban the user from your channel.")
	}

	ChatBanCheck := strings.Contains(text, "!ban")
	if ChatBanCheck == true {
		UsernameSplit := strings.Split(text, "!ban ")

		if len(UsernameSplit) <= 1 { // If len to handle index out of range error
			fmt.Println("Please type a username.")
		} else {
			BotSendMsg(conn, channel, "/ban "+UsernameSplit[1], name)
			fmt.Println(UsernameSplit[1] + " has been banned.")
		}
	}

	ChatUnBanCheck := strings.Contains(text, "!unban")
	if ChatUnBanCheck == true {
		UsernameSplit := strings.Split(text, "!unban ")

		if len(UsernameSplit) <= 1 { // If len to handle index out of range error
			fmt.Println("Please type a username.")
		} else {
			BotSendMsg(conn, channel, "/unban "+UsernameSplit[1], name)
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
			BotSendMsg(conn, channel, ChatCommand, name)
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
			fmt.Println("Please type a word.")
		} else {
			database := InitializeDB()
			statement, err := database.Prepare("INSERT INTO badwords (Badword) VALUES (?)")
			if err != nil {
				fmt.Printf("Error: %s", err)
			}
			statement.Exec(BadwordSplit[1])
			// Append to the slice in this run session to make it useable right away
			badwords.BadwordSlice = append(badwords.BadwordSlice, BadwordSplit[1])
			fmt.Println(badwords.BadwordSlice)
		}
	}

	ChatAddComCheck := strings.Contains(text, "!addcom")
	if ChatAddComCheck == true {
		CommandSplit := strings.Split(text, " ")

		if len(CommandSplit) <= 1 { // If len to handle index out of range error
			fmt.Println("Please type a proper command.")
		} else {
			fmt.Println(CommandSplit[1])
			fmt.Println(CommandSplit[2:])
		}
	}

	ChatAddQuoteCheck := strings.Contains(text, "!addquote")
	if ChatAddQuoteCheck == true {
		QuoteSplit := strings.Split(text, "!addquote ")

		if len(QuoteSplit) <= 1 {
			fmt.Println("Please type a new quote. ")
		} else {
			currenttime := time.Now()
			NewTime := currenttime.Format("2006-01-02")
			NewQuote := QuoteSplit[1] + " - " + NewTime
			fmt.Println(NewQuote)
		}
	}

	ExitCheck := strings.Contains(text, "!exit")
	if ExitCheck == true {
		os.Exit(3)
	}
}
