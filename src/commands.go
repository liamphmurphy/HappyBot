package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type CustomCommand struct {
	CommandName       string
	CommandResponse   string
	CommandPermission string
}

type CustomTimedCommand struct {
	TimedName     string
	TimedResponse string
	Timer         time.Duration
}

// MakeCommand assigns the response and permissions for each command
func MakeCommand(response, permission string) *CustomCommand {
	return &CustomCommand{
		CommandResponse:   response,
		CommandPermission: permission,
	}
}

func CommandOperations(chatmessage string) map[string]*CustomCommand {
	// Create a slice of the elements in a users message
	comSplit := strings.Split(chatmessage, " ")

	// Get the key and new value for sake of database
	comKey := comSplit[1]
	comNewValue := strings.Join(comSplit[2:], " ")

	database := InitializeDB()
	if comSplit[0] == "!editcom" {
		rows, err := database.Prepare("UPDATE commands SET CommandResponse = ? WHERE CommandName = ?")
		if err != nil {
			fmt.Println(err)
		}
		rows.Exec(comNewValue, comKey)
	}

	if comSplit[0] == "!addcom" {
		rows, err := database.Prepare("INSERT INTO commands (CommandName, CommandResponse) VALUES(?,?)")
		if err != nil {
			fmt.Println(err)
		}
		rows.Exec(comKey, comNewValue)
	}

	if comSplit[0] == "!setperm" {
		rows, err := database.Prepare("UPDATE commands SET CommandPermission = ? WHERE CommandName = ?")
		if err != nil {
			fmt.Println(err)
		}
		rows.Exec(comNewValue, comKey)
	}

	return LoadCommands()

}

// LoadCommands takes all commands from sqlite3 db and puts them in a map through CustomCommand struct
func LoadCommands() map[string]*CustomCommand {
	database := InitializeDB()

	rows, _ := database.Query("SELECT CommandName, CommandResponse, CommandPermission from commands")

	com := make(map[string]*CustomCommand)
	for rows.Next() {
		var CommandName, CommandResponse, CommandPermission string
		rows.Scan(&CommandName, &CommandResponse, &CommandPermission)
		com[CommandName] = MakeCommand(CommandResponse, CommandPermission)
	}
	return com
}

func CreateCommands(irc *BotInfo, usermessage string, potentialCommand string, com map[string]*CustomCommand, quotes map[string]string, badwords BadWord, goofs Goof, database *sql.DB, line string) (map[string]*CustomCommand, map[string]string, []string) {

	if potentialCommand == "!editcom" || potentialCommand == "!addcom" || potentialCommand == "!setperm" || potentialCommand == "!delcom" {
		com = CommandOperations(usermessage)
	}

	if potentialCommand == "!edittimed" || potentialCommand == "!addtimed" {
		TimedCommandOperations(usermessage)
	}

	// Check if user typed in !addgoof in the chat
	if potentialCommand == "!addgoof" {
		// Split data to separate username from value to use as new goof
		GoofSplit := strings.Split(usermessage, " ")
		GoofString := string(GoofSplit[1])
		fmt.Println(GoofSplit[1])

		statement, err := database.Prepare("INSERT INTO goofs (GoofName) VALUES (?)")
		if err != nil {
			fmt.Printf("Error: %s", err)
		}
		statement.Exec(GoofString) // Inserts value of GoofString into the (?) in previous SQL statement

		// Append to the slice in this run session to make it useable right away
		goofs.GoofSlice = append(goofs.GoofSlice, GoofString)

	}

	// Check if usermessage has !addquote in it
	if potentialCommand == "!addgoof" {
		// Check if user is moderator or broadcaster
		if CheckUserStatus(line, "moderator", irc) == "true" {
			quotes = AddQuote(irc, line, usermessage, irc.BotName)
		} else if CheckUserStatus(line, "broadcaster", irc) == "true" {
			quotes = AddQuote(irc, line, usermessage, irc.BotName)
		} else {
			BotSendMsg(irc, "Must be a moderator to add a new quote.")
		}

	}

	return com, quotes, goofs.GoofSlice
}

func DefaultCommands(irc *BotInfo, username string, usermessage string, potentialCommand string, line string, com map[string]*CustomCommand, quotes map[string]string, badwords BadWord, goofs Goof, permUsers []string, database *sql.DB) {

	// Check if a user is moderator or broadcaster before checking conditions for multiple commands.
	if CheckUserStatus(line, "moderator", irc) == "true" || CheckUserStatus(line, "broadcaster", irc) == "true" {
		if potentialCommand == "!settitle" {
			changeTitleSplit := strings.Split(usermessage, " ")
			PostStreamData(irc, irc.conn, irc.ChannelName, "title", changeTitleSplit[1:])
		}

		if potentialCommand == "!setgame" {
			changeGameSplit := strings.Split(usermessage, " ")
			PostStreamData(irc, irc.conn, irc.ChannelName, "game", changeGameSplit[1:])
		}

		if potentialCommand == "!givepoints" {
			splitMessage := strings.Split(usermessage, " ")
			pointsToGive, _ := strconv.Atoi(splitMessage[2])
			GivePoints(database, username, pointsToGive)
		}

		if potentialCommand == "!caster" {
			casterSplit := strings.Split(usermessage, " ")
			casterTargetMessage := strings.Replace(irc.CasterMessage, "{target}", casterSplit[1], -1)
			BotSendMsg(irc, casterTargetMessage)
		}

		if usermessage == "!listcoms" {
			paste := PostPasteBin(irc.PastebinKey, com)
			BotSendMsg(irc, "Command list: "+paste)

		}

		if potentialCommand == "!permit" {
			permitSplit := strings.Split(usermessage, " ")
			permUsers = append(permUsers, permitSplit[1])
			BotSendMsg(irc, permitSplit[1]+" can now post one link in chat.")
		}
	}

	if usermessage == "!"+irc.PointsName {
		userPoints := GetUserPoints(username)
		pointString := strconv.Itoa(userPoints)
		pointsTargetMessage := strings.Replace(irc.PointsMessage, "{target}", username, -1)
		pointsTargetMessage = strings.Replace(pointsTargetMessage, "{value}", pointString, -1)
		pointsTargetMessage = ReplaceStrings(pointsTargetMessage, "{currency}", irc.PointsName)
		BotSendMsg(irc, pointsTargetMessage)
	}

	if usermessage == "!game" {
		game := GetGame(irc.conn, irc.ChannelName)
		var gameName string
		if len(game.Data) > 0 {
			for _, val := range game.Data {
				gameName = val.Name
				BotSendMsg(irc, "@"+username+", "+gameName)
			}
		} else {
			BotSendMsg(irc, "@"+username+", stream is offline.")
		}
	}

	if irc.GamesEnabled == true {

	}

	for k, v := range quotes {
		if usermessage == "!quote "+k {
			BotSendMsg(irc, v)
		}

	}
	if usermessage == "!quote" {
		rows, err := database.Query("SELECT QuoteID, QuoteContent from quotes ORDER BY RANDOM() LIMIT 1;")
		if err != nil {
			fmt.Printf("Error: %s", err)
		}
		for rows.Next() {
			var QuoteID string
			var QuoteContent string
			rows.Scan(&QuoteID, &QuoteContent)
			quotes[QuoteID] = QuoteContent
			BotSendMsg(irc, QuoteContent)
		}
	}

	// Respond to user the current time, currently locked to the computer the bot is running on
	if usermessage == "!time" {
		if irc.StreamerTimeToggle == true {
			TimeCommands(irc, "StreamerTime", irc.ChannelName, irc.BotName, username)
		}
	}

	if usermessage == "!uptime" {
		TimeCommands(irc, "Uptime", irc.ChannelName, irc.BotName, username)
	}

	// Check if user set MakeLog in config.toml to true, if so, run
	if irc.MakeLog == true {
		// Use current date to mark which day the chat log is for
		currenttime := time.Now()
		datestring := currenttime.String()
		datesplit := strings.Split(datestring, " ")
		loglocation := "logs/chat/" + datesplit[0] + ".txt"
		logmessage := (username + ": " + usermessage + "\n")
		WriteToLog(loglocation, logmessage)
	}

	// Check if user set CheckLongMessageCap in config.toml to true, if so, run
	if irc.CheckLongMessageCap == true {
		if len(usermessage) > irc.LongMessageCap {
			fmt.Println("Very long message detected.")
			PurgeUser(irc, username)
			BotSendMsg(irc, "@"+username+" please shorten your message")

		}
	}

	// For each value in LinkChecks array in config.toml, check whether to purge user or not.
	for _, v := range irc.LinkChecks {
		if strings.Contains(usermessage, v) {
			if irc.PurgeForLinks == true {

				// Check if user is in the permitted slice
				userCheck := UserInSlice(username, permUsers)
				// If user is a moderator / broadcaster, just let them post link
				if CheckUserStatus(line, "moderator", irc) == "true" || CheckUserStatus(line, "broadcaster", irc) == "true" {
					fmt.Println("Link permitted.")
				} else if userCheck == true { // If not a moderator / broadcaster, but is in the permitted slice, let them post link then remove them
					position := GetSlicePosition(username, permUsers)
					permUsers = RemoveFromSlice(position, permUsers)
				} else { // If none of the above is true, purge user
					PurgeUser(irc, username)
				}
			}
		}
	}
}

func UserCommands(irc *BotInfo, username string, usermessage string, line string, com map[string]*CustomCommand, quotes map[string]string, badwords BadWord, goofs Goof, permUsers []string, giveawayEntryTerm string, giveawayUsers []string, database *sql.DB) {
	/* These iterations are not put in DefaultCommands because these include custom values, such as commands from the user.
	This delineation is made more for code organization, not because the placement makes a huge difference. */
	for _, v := range goofs.GoofSlice {
		if usermessage == v {
			BotSendMsg(irc, v)
		}
	}

	for _, v := range badwords.BadwordSlice {
		if strings.Contains(usermessage, v) {
			BanUser(irc, username)
		}
	}

	for k, v := range com {
		if usermessage == k {
			if CheckUserStatus(line, v.CommandPermission, irc) == "true" {
				BotSendMsg(irc, v.CommandResponse)
			} else if v.CommandPermission == "all" {
				BotSendMsg(irc, v.CommandResponse)
			} else {
				BotSendMsg(irc, "@"+username+" Insufficient perms to run that command.")
			}
		}
	}
}
