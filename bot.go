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
	ChannelName         string
	ServerName          string
	BotOAuth            string
	BotName             string
	conn                net.Conn
	CheckLongMessageCap bool
	LongMessageCap      int
	StreamerTimeToggle  bool
	MakeLog             bool
	SubResponse         string
	PurgeForLinks       bool
	LinkChecks          []string
}

type CustomCommand struct {
	CommandName     []string
	CommandResponse []string
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
	var genconfig BotInfo
	_, conferr := toml.DecodeFile("config/config.toml", &genconfig)
	if conferr != nil {
		fmt.Println("Can't read toml file due to:", conferr)
	}

	return &BotInfo{
		ChannelName:         genconfig.ChannelName,
		ServerName:          genconfig.ServerName,
		BotOAuth:            genconfig.BotOAuth,
		BotName:             genconfig.BotName,
		LongMessageCap:      genconfig.LongMessageCap,
		StreamerTimeToggle:  genconfig.StreamerTimeToggle,
		MakeLog:             genconfig.MakeLog,
		SubResponse:         genconfig.SubResponse,
		PurgeForLinks:       genconfig.PurgeForLinks,
		LinkChecks:          genconfig.LinkChecks,
		CheckLongMessageCap: genconfig.CheckLongMessageCap,
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

// Load all custom commands not included with the bot by default.
func LoadCustomCommands() map[string]string {
	//var customcommand CustomCommand
	database := InitializeDB()

	rows, _ := database.Query("SELECT CommandName, CommandResponse from commands")

	com := map[string]string{}
	for rows.Next() {
		var CommandName, CommandResponse string
		rows.Scan(&CommandName, &CommandResponse)
		com[CommandName] = CommandResponse
	}
	return com
}

// Load all quotes custom made by the user.
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

// Function used throughout the program for the bot to send IRC messages
func BotSendMsg(conn net.Conn, channel string, message string) {
	fmt.Fprintf(conn, "PRIVMSG %s :%s\r\n", channel, message)
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
func TimeCommands(TimeSetting string, conn net.Conn, channel string) string {
	currenttime := time.Now()
	datestring := currenttime.String()
	//datesplit := strings.Split(datestring, " ")

	if TimeSetting == "StreamerTime" {
		NewTime := currenttime.Format("3:04 PM MST") // Does not actually = 3:04 PM, golang pattern matching used here
		StreamerNameSplit := strings.Split(channel, "#")
		StreamerString := StreamerNameSplit[1] + "'s" + " time: " + NewTime
		BotSendMsg(conn, channel, StreamerString)
	}
	return datestring
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

	database, err := sql.Open("sqlite3", "./happybot.db")
	if err != nil {
		fmt.Printf("Error: %s", err)
	}

	statement, err := database.Prepare("CREATE TABLE IF NOT EXISTS commands (CommandID INTEGER PRIMARY KEY, CommandName TEXT, CommandResponse TEXT)")
	if err != nil {
		fmt.Printf("Error: %s", err)
	}
	statement.Exec()

	irc := CreateBot()
	irc.Connect()

	badwords := LoadBadWords()
	com := LoadCustomCommands()
	goofs := LoadGoofs()
	quotes := LoadQuotes()

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
	for {
		line, err := proto.ReadLine()
		if err != nil {
			break
		}

		/* Run ConsoleInput on new thread
		Allows user to type commands into terminal window */
		go ConsoleInput(irc.conn, irc.ChannelName)

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
					BotSendMsg(irc.conn, irc.ChannelName, botresponse)
					BotSendMsg(irc.conn, irc.ChannelName, "@"+username[1]+" please shorten your message")

				}
			}

			// For each value in LinkChecks array in config.toml, check whether to purge user or not.
			for _, v := range irc.LinkChecks {
				userbadges1 := strings.Split(line, "@badges=")
				userbadges2 := strings.Split(userbadges1[1], ";")
				if strings.Contains(usermessage, v) {
					if irc.PurgeForLinks == true {
						// Check for different types of user badges (should find a better way to check this)
						if strings.Contains(userbadges2[0], "subscriber") {
							fmt.Println("Link permitted: Sub.")
							fmt.Println("userbadge is: " + userbadges2[0])
						}
						if strings.Contains(userbadges2[0], "moderator") {
							fmt.Println("Link permitted: Moderator.")
						}
						if strings.Contains(userbadges2[0], "broadcaster") {
							fmt.Println("Link permitted: Broadcaster.")
						}
						if strings.Contains(userbadges2[0], "") {
							botresponse := "/timeout " + username[2] + " 1" + " Link when not a mod."
							BotSendMsg(irc.conn, irc.ChannelName, botresponse)
							BotSendMsg(irc.conn, irc.ChannelName, "@"+username[2]+" please ask for permission to post a link.")
						}
					}
				}
			}
			
			// Check for occurences of values from arrays/slices/maps etc

			for _, v := range goofs.GoofSlice {
				if usermessage == v {
					BotSendMsg(irc.conn, irc.ChannelName, v)
				}
			}

			for _, v := range badwords.BadwordSlice {
				if strings.Contains(usermessage, v) {
					fmt.Println(username[2], "has been banned.")
					BotSendMsg(irc.conn, irc.ChannelName, "/ban "+username[2])
				}
			}

			for k, v := range com {
				if usermessage == k {
					BotSendMsg(irc.conn, irc.ChannelName, v)
				}
			}

			for k, v := range quotes {
				if usermessage == "!quote "+k {
					BotSendMsg(irc.conn, irc.ChannelName, v)
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
					BotSendMsg(irc.conn, irc.ChannelName, QuoteContent)
				}
			}

			// Check if user typed in !addgofo in the chat
			CheckForGoof := strings.Contains(usermessage, "!addgoof")
			if CheckForGoof == true {
				// SQL statement: create goofs table if it does not exist into db
				statement, err := database.Prepare("CREATE TABLE IF NOT EXISTS goofs (GoofID INTEGER PRIMARY KEY, GoofName text)")
				if err != nil {
					fmt.Printf("Error: %s", err)
				}
				statement.Exec() // Needed to execute previous declaration of statement 

				// Split data to separate username from value to use as new goof
				GoofSplit := strings.Split(usermessage, " ")
				GoofString := string(GoofSplit[1])
				fmt.Println(GoofSplit[1])

				statement, err = database.Prepare("INSERT INTO goofs (GoofName) VALUES (?)")
				if err != nil {
					fmt.Printf("Error: %s", err)
				}
				statement.Exec(GoofString) // Inserts value of GoofString into the (?) in previous SQL statement
				
				// Append to the slice in this run session to make it useable right away
				goofs.GoofSlice = append(goofs.GoofSlice, GoofString)

			}

			CheckForAddQuote := strings.Contains(usermessage, "!addquote")
			if CheckForAddQuote == true {
				QuoteSplit := strings.Split(usermessage, "!addquote ")
				currenttime := time.Now()
				NewTime := currenttime.Format("2006-01-02")
				NewQuote := QuoteSplit[1] + " - " + NewTime
				fmt.Println(NewQuote)

				statement, err = database.Prepare("INSERT INTO quotes (QuoteContent) VALUES (?)")
				if err != nil {
					fmt.Printf("Error: %s", err)
				}
				statement.Exec(NewQuote)
			}

			// Respond to user the current time, currently locked to the computer the bot is running on
			if usermessage == "!time" {
				if irc.StreamerTimeToggle == true {
					TimeCommands("StreamerTime", irc.conn, irc.ChannelName)
				}
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
				BotSendMsg(irc.conn, irc.ChannelName, botsubresponse)
				// Append new sub to a list of new subs in current session for logging
				SubsCurrentStream = append(SubsCurrentStream, username2[0])
			}
		}

	}

}
