package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	_ "github.com/mattn/go-sqlite3"
)

type BotInfo struct {
	ChannelName                 string
	ServerName                  string
	BotOAuth                    string
	BotName                     string
	conn                        net.Conn
	LetModeratorsUseAllCommands bool
	CheckLongMessageCap         bool
	LongMessageCap              int
	StreamerTimeToggle          bool
	MakeLog                     bool
	SubResponse                 string
	PurgeForLinks               bool
	LinkChecks                  []string
	HydrateOn                   bool
	HydrateTime                 time.Duration
	HydrateMessage              string
	PastebinKey                 string
}

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

type BadWord struct {
	BadWordItem  string
	BadwordSlice []string
	TimeoutText  []string
}

type Goof struct {
	GoofName  string
	GoofSlice []string
}

func CreateBot() *BotInfo {
	var genConfig BotInfo
	_, confErr := toml.DecodeFile("config/config.toml", &genConfig)
	if confErr != nil {
		fmt.Println("Can't read toml file due to: ", confErr)
	}

	return &BotInfo{
		ChannelName:                 genConfig.ChannelName,
		ServerName:                  genConfig.ServerName,
		BotOAuth:                    genConfig.BotOAuth,
		BotName:                     genConfig.BotName,
		LetModeratorsUseAllCommands: genConfig.LetModeratorsUseAllCommands,
		LongMessageCap:              genConfig.LongMessageCap,
		StreamerTimeToggle:          genConfig.StreamerTimeToggle,
		MakeLog:                     genConfig.MakeLog,
		SubResponse:                 genConfig.SubResponse,
		PurgeForLinks:               genConfig.PurgeForLinks,
		LinkChecks:                  genConfig.LinkChecks,
		CheckLongMessageCap:         genConfig.CheckLongMessageCap,
		HydrateOn:                   genConfig.HydrateOn,
		HydrateTime:                 genConfig.HydrateTime,
		HydrateMessage:              genConfig.HydrateMessage,
		PastebinKey:                 genConfig.PastebinKey,
	}
}

// All "Load" functions read the toml files/databases for various chat features, like commands/bannable words.

/* Goofs serve no real purpose. Some chats like to have the bot 'repeat' what the user
types in, perhaps for a specific emote.*/

func LoadGoofs() Goof {
	var goofs Goof
	database := InitializeDB()

	rows, _ := database.Query("SELECT GoofName FROM goofs")
	for rows.Next() {
		rows.Scan(&goofs.GoofName)
		goofs.GoofSlice = append(goofs.GoofSlice, goofs.GoofName)
	}

	return goofs
}

// Loads all words that are to be banned: if user types bad word in chat, the user is banned.
func LoadBadWords() BadWord {
	var badwords BadWord
	database := InitializeDB()

	statement, err := database.Prepare("CREATE TABLE IF NOT EXISTS badwords (BadwordID INTEGER PRIMARY KEY, Badword TEXT)")
	statement.Exec()
	if err != nil {
		fmt.Printf("Error: %s", err)
	}

	rows, _ := database.Query("SELECT Badword FROM badwords")
	for rows.Next() {
		rows.Scan(&badwords.BadWordItem)
		badwords.BadwordSlice = append(badwords.BadwordSlice, badwords.BadWordItem)
	}
	return badwords
}

// LoadQuotes grabs all quotes from sqlite3 db.
func LoadQuotes() map[string]string {
	database := InitializeDB()
	rows, _ := database.Query("SELECT QuoteID, QuoteContent from quotes")

	quotes := map[string]string{}
	for rows.Next() {
		var QuoteID string
		var QuoteContent string
		rows.Scan(&QuoteID, &QuoteContent)
		quotes[QuoteID] = QuoteContent
	}
	return quotes
}

// Write to log function, when called, will run when set to true in config.
func WriteToLog(log string, text string) {
	f, _ := os.OpenFile(log, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	f.WriteString(text)
}

// Init database and then return it
func InitializeDB() *sql.DB {
	database, err := sql.Open("sqlite3", "./happybot.db")
	if err != nil {
		fmt.Printf("Error: %s", err)
	}
	return database
}

// Series of commands to do with time, like uptime.
func TimeCommands(TimeSetting string, conn net.Conn, channel string, name string, username string) string {
	currentTime := time.Now()
	dateString := currentTime.String()
	//datesplit := strings.Split(datestring, " ")

	// Uses system time instead of twitch api data.
	if TimeSetting == "StreamerTime" {
		newTime := currentTime.Format("3:04 PM MST") // Does not actually = 3:04 PM, golang pattern matching used here
		streamerNameSplit := strings.Split(channel, "#")
		streamerString := streamerNameSplit[1] + "'s" + " time: " + newTime
		BotSendMsg(conn, channel, streamerString, name)
	}

	if TimeSetting == "Uptime" {
		s := StreamData(conn, channel)
		// If variable 's' has data returned, stream is live and will continue.
		if len(s.Data) > 0 {
			for _, val := range s.Data {
				// Grabs the StartedAt value from JSON, showing timestamp when stream went live.
				timeSince := time.Since(val.StartedAt)

				// Use timeSince to calculate the difference between timestamp and current time.
				sinceSplit := strings.Split(timeSince.String(), ".")
				// Begin replacing single characters for time units to full words and make it nicer looking.
				newSplit := strings.Replace(sinceSplit[0], "h", " hours, ", -1)
				newSplit = strings.Replace(newSplit, "m", " minutes, ", -1)
				newMessage := "@" + username + " " + newSplit + " seconds."

				BotSendMsg(conn, channel, newMessage, name)
			}
			// if no data in 's', stream is not live.
		} else {
			BotSendMsg(conn, channel, "Stream is not live.", name)
		}
	}
	return dateString
}

// Function to remove '#' from channel name, typically for URL purposes in API's.
func SplitChannelName(channel string) string {
	newChannel := strings.Split(channel, "#")
	return newChannel[1]
}

// Function to add a new quote and return a map of quotes, including new one.
func AddQuote(conn net.Conn, channel string, message string, usermessage string, name string) map[string]string {
	database := InitializeDB()

	quoteSplit := strings.Split(usermessage, "!addquote ")
	currentTime := time.Now()
	newTime := currentTime.Format("2006-01-02")
	newQuote := quoteSplit[1] + " - " + newTime
	statement, err := database.Prepare("INSERT INTO quotes (QuoteContent) VALUES (?)")
	if err != nil {
		fmt.Printf("Error: %s", err)
	}
	statement.Exec(newQuote)
	BotSendMsg(conn, channel, "Quote added!", name)

	return LoadQuotes()
}

// Function to add a new goof and return a slice of goofs, including new one.
func AddGoof(usermessage string) Goof {
	database := InitializeDB()
	// Split data to separate username from value to use as new goof
	GoofSplit := strings.Split(usermessage, " ")
	GoofString := string(GoofSplit[1])
	fmt.Println(GoofSplit[1])

	statement, err := database.Prepare("INSERT INTO goofs (GoofName) VALUES (?)")
	if err != nil {
		fmt.Printf("Error: %s", err)
	}
	statement.Exec(GoofString) // Inserts value of GoofString into the (?) in previous SQL statement
	return LoadGoofs()
}

// CheckUserStatus checks if user is allowed to run a command
func CheckUserStatus(chatmessage string, permcheck string, irc *BotInfo) string {

	firstBadgeSplit := strings.Split(chatmessage, "@badges=")
	endBadgeSplit := strings.Split(firstBadgeSplit[1], ";")
	strings.Contains(endBadgeSplit[0], permcheck)
	if strings.Contains(endBadgeSplit[0], permcheck) {
		boolcheck := "true"
		return boolcheck
	}
	if strings.Contains(endBadgeSplit[0], "all") {
		boolcheck := "true"
		return boolcheck
	}
	if irc.LetModeratorsUseAllCommands == true {
		if strings.Contains(endBadgeSplit[0], "moderator") {
			boolcheck := "true"
			return boolcheck
		}
	}

	if strings.Contains(endBadgeSplit[0], "broadcaster") {
		boolcheck := "true"
		return boolcheck
	} else {
		boolcheck := "false"
		return boolcheck
	}
	return ""
}

func HydrateReminder(irc *BotInfo, conn net.Conn, channel string) {
	if irc.HydrateOn == true {
		hydrateTo := SplitChannelName(channel)
		for range time.NewTicker(irc.HydrateTime * time.Second * 60).C {
			BotSendMsg(conn, channel, "@"+hydrateTo+" "+irc.HydrateMessage, irc.BotName)
		}
	}
}

// Function used throughout the program for the bot to send IRC messages
func BotSendMsg(conn net.Conn, channel string, message string, name string) {
	fmt.Fprintf(conn, "PRIVMSG %s :%s\r\n", channel, message)
	fmt.Println(name + ": " + message) // Display bot's message in terminal
}

/* ConsoleInput function for reading user input in cmd line when
   program is running */

// Connect to the Twitch IRC server
func (bot *BotInfo) Connect() {
	var err error
	bot.conn, err = net.Dial("tcp", bot.ServerName)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Connected to: %s\n", bot.ServerName)
}

func main() {
	var database *sql.DB
	database = InitializeDB()
	irc := CreateBot()
	irc.Connect()

	// Declare commands, quotes etc maps here so that it can be changed dynamically later on.
	var com map[string]*CustomCommand
	var quotes map[string]string

	badwords := LoadBadWords()
	com = LoadCommands()
	goofs := LoadGoofs()
	quotes = LoadQuotes()

	// Pass info to HTTP request
	fmt.Fprintf(irc.conn, "PASS %s\r\n", irc.BotOAuth)
	fmt.Fprintf(irc.conn, "NICK %s\r\n", irc.BotName)
	fmt.Fprintf(irc.conn, "JOIN %s\r\n", irc.ChannelName)

	// Twitch specific information, like badges, mod status etc.
	fmt.Fprintf(irc.conn, "CAP REQ :twitch.tv/membership\r\n")
	fmt.Fprintf(irc.conn, "CAP REQ :twitch.tv/tags\r\n")
	fmt.Fprintf(irc.conn, "CAP REQ :twitch.tv/commands\r\n")

	fmt.Printf("Channel: " + irc.ChannelName + "\n")

	defer irc.conn.Close()
	reader := bufio.NewReader(irc.conn)
	proto := textproto.NewReader(reader)

	currenttime := time.Now()
	datestring := currenttime.String()
	datesplit := strings.Split(datestring, " ")

	// If user wants it, have the bot remind them to hydrate.
	go HydrateReminder(irc, irc.conn, irc.ChannelName)
	//go RunPoints(irc.conn, irc.ChannelName)

	for {
		go ConsoleInput(irc.conn, irc.ChannelName, irc.BotName)
		line, err := proto.ReadLine()
		if err != nil {
			break
		}

		/* Run ConsoleInput on new thread
		Allows user to type commands into terminal window */

		// When Twitch servers send a ping, respond with pong to avoid disconnections.
		if strings.Contains(line, "PING") {
			pong := strings.Split(line, "PING")
			fmt.Fprintf(irc.conn, "PONG %s\r\n", pong[1])

			// Parse the data received from each chat message into something readable.
		} else if strings.Contains(line, ".tmi.twitch.tv PRIVMSG "+irc.ChannelName) {
			userdata := strings.Split(line, ".tmi.twitch.tv PRIVMSG "+irc.ChannelName)
			username := strings.Split(userdata[0], "@")
			usermessage := strings.Replace(userdata[1], " :", "", 1)

			// Display the whole cleaned up message
			fmt.Printf(username[2] + ": " + usermessage + "\n")

			// Check if user set MakeLog in config.toml to true, if so, run
			if irc.MakeLog == true {
				// Use current date to mark which day the chat log is for
				loglocation := "logs/chat/" + datesplit[0] + ".txt"
				logmessage := (username[2] + ": " + usermessage + "\n")
				WriteToLog(loglocation, logmessage)
			}

			// Check if user set CheckLongMessageCap in config.toml to true, if so, run
			if irc.CheckLongMessageCap == true {
				if len(usermessage) > irc.LongMessageCap {
					fmt.Println("Very long message detected.")
					botresponse := "/timeout " + username[1] + " 1" + "Message over max character limit."
					BotSendMsg(irc.conn, irc.ChannelName, botresponse, irc.BotName)
					BotSendMsg(irc.conn, irc.ChannelName, "@"+username[1]+" please shorten your message", irc.BotName)

				}
			}

			// For each value in LinkChecks array in config.toml, check whether to purge user or not.
			for _, v := range irc.LinkChecks {
				firstBadgeSplit := strings.Split(line, "@badges=")
				endBadgeSplit := strings.Split(firstBadgeSplit[1], ";")
				if strings.Contains(usermessage, v) {
					if irc.PurgeForLinks == true {
						// Check for different types of user badges (should find a better way to check this)
						if CheckUserStatus(line, "subscriber", irc) == "true" {
							fmt.Println("Link permitted: Sub.")
							fmt.Println("userbadge is: " + endBadgeSplit[0])
						}
						if CheckUserStatus(line, "moderator", irc) == "true" {
							fmt.Println("Link permitted: Moderator.")
						}
						if CheckUserStatus(line, "broadcaster", irc) == "true" {
							fmt.Println("Link permitted: Broadcaster.")
						} else {
							botresponse := "/timeout " + username[2] + " 1"
							BotSendMsg(irc.conn, irc.ChannelName, botresponse, irc.BotName)
							BotSendMsg(irc.conn, irc.ChannelName, "@"+username[2]+" please ask for permission to post a link.", irc.BotName)
							fmt.Println(botresponse)
						}
					}
				}
			}

			if strings.Contains(usermessage, "!editcom") || strings.Contains(usermessage, "!addcom") || strings.Contains(usermessage, "!setperm") {
				if CheckUserStatus(line, "moderator", irc) == "true" || CheckUserStatus(line, "broadcaster", irc) == "true" {
					com = CommandOperations(usermessage)
				} else {
					BotSendMsg(irc.conn, irc.ChannelName, "@"+username[2]+" Insufficient permissions to change commands.", irc.BotName)
				}
			}

			if usermessage == "!listcoms" {
				paste := PostPasteBin(irc.PastebinKey, com)
				BotSendMsg(irc.conn, irc.ChannelName, "Command list: "+paste, irc.BotName)

			}

			// Check for occurences of values from arrays/slices/maps etc

			for _, v := range goofs.GoofSlice {
				if usermessage == v {
					BotSendMsg(irc.conn, irc.ChannelName, v, irc.BotName)
				}
			}

			for _, v := range badwords.BadwordSlice {
				if strings.Contains(usermessage, v) {
					fmt.Println(username[2], "has been banned.")
					BotSendMsg(irc.conn, irc.ChannelName, "/ban "+username[2], irc.BotName)
				}
			}

			for k, v := range com {
				if usermessage == k {
					if CheckUserStatus(line, v.CommandPermission, irc) == "true" {
						BotSendMsg(irc.conn, irc.ChannelName, v.CommandResponse, irc.BotName)
					} else if v.CommandPermission == "all" {
						BotSendMsg(irc.conn, irc.ChannelName, v.CommandResponse, irc.BotName)
					} else {
						BotSendMsg(irc.conn, irc.ChannelName, "@"+username[2]+" Insufficient perms to run that command.", irc.BotName)
					}
				}
			}

			for k, v := range quotes {
				if usermessage == "!quote "+k {
					BotSendMsg(irc.conn, irc.ChannelName, v, irc.BotName)
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
					BotSendMsg(irc.conn, irc.ChannelName, QuoteContent, irc.BotName)
				}
			}

			// Check if user typed in !addgoof in the chat
			checkForGoof := strings.Contains(usermessage, "!addgoof")
			if checkForGoof == true {
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
			CheckForAddQuote := strings.Contains(usermessage, "!addquote")
			if CheckForAddQuote == true {
				// Check if user is moderator or broadcaster
				if CheckUserStatus(line, "moderator", irc) == "true" {
					AddQuote(irc.conn, irc.ChannelName, line, usermessage, irc.BotName)
				} else if CheckUserStatus(line, "broadcaster", irc) == "true" {
					AddQuote(irc.conn, irc.ChannelName, line, usermessage, irc.BotName)
				} else {
					BotSendMsg(irc.conn, irc.ChannelName, "Must be a moderator to add a new quote.", irc.BotName)
				}

			}

			// Respond to user the current time, currently locked to the computer the bot is running on
			if usermessage == "!time" {
				if irc.StreamerTimeToggle == true {
					TimeCommands("StreamerTime", irc.conn, irc.ChannelName, irc.BotName, username[2])
				}
			}

			if usermessage == "!uptime" {
				TimeCommands("Uptime", irc.conn, irc.ChannelName, irc.BotName, username[2])
			}

		} else if strings.Contains(line, "USERNOTICE") {
			// user variables used to split the twitch tag string to get the username
			if strings.Contains(line, "msg-param-sub-plan") {
				var SubsCurrentStream []string
				username1 := strings.Split(line, "display-name=")
				username2 := strings.Split(username1[1], ";")

				// Thank the user for subbing
				botsubresponse := "@" + username2[0] + " " + irc.SubResponse
				fmt.Println(botsubresponse)
				BotSendMsg(irc.conn, irc.ChannelName, botsubresponse, irc.BotName)
				// Append new sub to a list of new subs in current session for logging
				SubsCurrentStream = append(SubsCurrentStream, username2[0])
				if irc.MakeLog == true {
					logLocation := "logs/NewSubs " + datesplit[0] + ".txt"
					logMessage := username2[0] + "\n"
					WriteToLog(logLocation, logMessage)
				}
			}
		}

	}

}
