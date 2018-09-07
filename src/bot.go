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
	WebAppGUIEnabled            bool
	PointsSystemEnabled         bool
	PointsName                  string
	PointsValueModifier         int
	PointsIncrementTime         time.Duration
	PointsMessage               string
	GamesEnabled                bool
	RouletteEnabled             bool
	RouletteWinMessages         []string
	RouletteLossMessages        []string
	conn                        net.Conn
	LetModeratorsUseAllCommands bool
	CasterMessage               string
	CheckLongMessageCap         bool
	LongMessageCap              int
	StreamerTimeToggle          bool
	MakeLog                     bool
	RespondToSubs               bool
	SubResponse                 string
	PurgeForLinks               bool
	LinkChecks                  []string
	HydrateOn                   bool
	HydrateTime                 time.Duration
	HydrateMessage              string
	SendMessages                bool
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
	_, confErr := toml.DecodeFile("../config/config.toml", &genConfig)
	if confErr != nil {
		fmt.Println("Can't read toml file due to: ", confErr)
	}

	sendMessages := true
	for _, v := range os.Args {
		if v == "--quiet" {
			sendMessages = false
		} else if v == "-q" {
			sendMessages = false
		}
	}

	return &BotInfo{
		ChannelName:                 genConfig.ChannelName,
		ServerName:                  genConfig.ServerName,
		BotOAuth:                    genConfig.BotOAuth,
		BotName:                     genConfig.BotName,
		WebAppGUIEnabled:            genConfig.WebAppGUIEnabled,
		PointsSystemEnabled:         genConfig.PointsSystemEnabled,
		PointsName:                  genConfig.PointsName,
		PointsValueModifier:         genConfig.PointsValueModifier,
		PointsIncrementTime:         genConfig.PointsIncrementTime,
		PointsMessage:               genConfig.PointsMessage,
		GamesEnabled:                genConfig.GamesEnabled,
		RouletteEnabled:             genConfig.RouletteEnabled,
		RouletteWinMessages:         genConfig.RouletteWinMessages,
		RouletteLossMessages:        genConfig.RouletteLossMessages,
		LetModeratorsUseAllCommands: genConfig.LetModeratorsUseAllCommands,
		CasterMessage:               genConfig.CasterMessage,
		LongMessageCap:              genConfig.LongMessageCap,
		StreamerTimeToggle:          genConfig.StreamerTimeToggle,
		MakeLog:                     genConfig.MakeLog,
		RespondToSubs:               genConfig.RespondToSubs,
		SubResponse:                 genConfig.SubResponse,
		PurgeForLinks:               genConfig.PurgeForLinks,
		LinkChecks:                  genConfig.LinkChecks,
		CheckLongMessageCap:         genConfig.CheckLongMessageCap,
		HydrateOn:                   genConfig.HydrateOn,
		HydrateTime:                 genConfig.HydrateTime,
		HydrateMessage:              genConfig.HydrateMessage,
		SendMessages:                sendMessages,
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
	database, err := sql.Open("sqlite3", "happybot.db")
	if err != nil {
		fmt.Printf("Error: %s", err)
	}
	return database
}

// Series of commands to do with time, like uptime.
func TimeCommands(irc *BotInfo, TimeSetting string, channel string, name string, username string) string {
	currentTime := time.Now()
	dateString := currentTime.String()
	//datesplit := strings.Split(datestring, " ")

	// Uses system time instead of twitch api data.
	if TimeSetting == "StreamerTime" {
		newTime := currentTime.Format("3:04 PM MST") // Does not actually = 3:04 PM, golang pattern matching used here
		streamerNameSplit := strings.Split(channel, "#")
		streamerString := streamerNameSplit[1] + "'s" + " time: " + newTime
		BotSendMsg(irc, streamerString)
	}

	if TimeSetting == "Uptime" {
		s := StreamData(irc.conn, channel)
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

				BotSendMsg(irc, newMessage)
			}
			// if no data in 's', stream is not live.
		} else {
			BotSendMsg(irc, "Stream is not live.")
		}
	}
	return dateString
}

// Function to remove '#' from channel name, typically for URL purposes in API's.
func SplitChannelName(channel string) string {
	newChannel := strings.Split(channel, "#")
	return newChannel[1]
}

// Bans user with provided username.
func BanUser(irc *BotInfo, user string) {
	fmt.Println(user, "has been banned.")
	BotSendMsg(irc, "/ban "+user)
}

// Time out user with provided username.
func TimeOutUser(irc *BotInfo, user string) {
	fmt.Println(user, "has been timed out.")
	BotSendMsg(irc, "/timeout "+user+" 60")
}

// Purge user with provided username.
func PurgeUser(irc *BotInfo, user string) {
	fmt.Println(user, "has been purged.")
	BotSendMsg(irc, "/timeout "+user+" 1")
}

// Function to add a new quote and return a map of quotes, including new one.
func AddQuote(irc *BotInfo, message string, usermessage string, name string) map[string]string {
	database := InitializeDB()

	quoteSplit := strings.Split(usermessage, "!addquote ")
	currentTime := time.Now()
	newTime := currentTime.Format("2006-01-02")
	newQuote := quoteSplit[1] + " -- " + newTime

	rows, _ := database.Query("SELECT QuoteID, QuoteContent from quotes")

	quotes := map[string]string{}
	var counter int
	for rows.Next() {
		counter++
		var QuoteID string
		var QuoteContent string
		rows.Scan(&QuoteID, &QuoteContent)
		quotes[QuoteID] = QuoteContent
	}

	newQuoteID := counter + 1

	statement, err := database.Prepare("INSERT INTO quotes (QuoteID, QuoteContent) VALUES (?, ?)")
	if err != nil {
		fmt.Printf("Error: %s", err)
	}
	statement.Exec(newQuoteID, newQuote)
	BotSendMsg(irc, "Quote added!")

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

func HydrateReminder(irc *BotInfo) {
	hydrateTo := SplitChannelName(irc.ChannelName)
	for range time.NewTicker(irc.HydrateTime * time.Second * 60).C {
		BotSendMsg(irc, "@"+hydrateTo+" "+irc.HydrateMessage)
	}
}

// Check to see if user is in the permitted slice
func UserInSlice(user string, perm []string) bool {
	for _, username := range perm {
		if username == user {
			return true
		}
	}
	return false
}

// For sake of removing after a link is posted, iterate through the slice and get the element index
func GetSlicePosition(user string, perm []string) int {
	x := 0
	for _, username := range perm {
		x++
		if username == user {
			return x - 1
		}
	}
	return -1
}

// With the element index known, remove user from slice
func RemoveFromSlice(index int, perm []string) []string {
	perm[index] = perm[len(perm)-1]
	perm[len(perm)-1] = ""
	perm = perm[:len(perm)-1]

	return perm
}

// Function used throughout the program for the bot to send IRC messages
func BotSendMsg(irc *BotInfo, message string) {
	if irc.SendMessages == true {
		fmt.Fprintf(irc.conn, "PRIVMSG %s :%s\r\n", irc.ChannelName, message)
		fmt.Println(irc.BotName + ": " + message) // Display bot's message in terminal
	}
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

	// For more information on this goroutine, look at the server.go file.
	if irc.WebAppGUIEnabled == true {
		go ServerMain()
	}

	// If user wants it, have the bot remind them to hydrate.
	if irc.HydrateOn == true {
		go HydrateReminder(irc)
	}
	if irc.PointsSystemEnabled == true {
		go RunPoints(irc.PointsIncrementTime, irc.PointsValueModifier, irc.conn, irc.ChannelName)
	}

	TimedCommands(irc)

	/* Below are variables we need to initialize so the values are kept throughout each iteration of the for loop.
	   In the case of games like a raffle, this is unoptimal, because there are variables for raffles hanging around in the for loop
	   even if the user turns raffles off in config.toml. */

	//var raffleRunning bool

	// Prepare variable for users permitted to post links.
	var permUsers []string

	// Prepare variables needed for giveaways.
	giveawayEntryTerm := "giveawayisnil"
	var giveawayUsers []string

	for {
		/* Run ConsoleInput on new thread
		Allows user to type commands into terminal window */
		go ConsoleInput(irc)
		line, err := proto.ReadLine()
		if err != nil {
			break
		}

		// When Twitch servers send a ping, respond with pong to avoid disconnections.
		if strings.Contains(line, "PING") {
			pong := strings.Split(line, "PING")
			fmt.Fprintf(irc.conn, "PONG %s\r\n", pong[1])

			// Parse the data received from each chat message into something readable.
		} else if strings.Contains(line, ".tmi.twitch.tv PRIVMSG "+irc.ChannelName) {
			userdata := strings.Split(line, ".tmi.twitch.tv PRIVMSG "+irc.ChannelName)
			splitdata := strings.Split(userdata[0], "@")
			username := splitdata[2]
			usermessage := strings.Replace(userdata[1], " :", "", 1)

			// Display the whole cleaned up message
			fmt.Println(username + ": " + usermessage)

			/* If a moderator or broadcaster, their message may be to edit / add / delete a command.
			If they are, run CreateCommands, which updates these values for the chat to use. This may not be the optimal solution,
			but it makes it so normal users' messages aren't checked. */
			if CheckUserStatus(line, "moderator", irc) == "true" || CheckUserStatus(line, "broadcaster", irc) == "true" {
				com, quotes, goofs.GoofSlice = CreateCommands(irc, com, quotes, badwords, goofs, usermessage, database, line)
			}

			// Default commands for the bot are put in DefaultCommands. Things like !caster, !permit etc can be seen there.
			DefaultCommands(irc, username, usermessage, line, com, quotes, badwords, goofs, permUsers, giveawayEntryTerm, giveawayUsers, database)

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

		} else if strings.Contains(line, "USERNOTICE") {
			currenttime := time.Now()
			datestring := currenttime.String()
			datesplit := strings.Split(datestring, " ")
			// user variables used to split the twitch tag string to get the username
			if strings.Contains(line, "msg-param-sub-plan") {
				if irc.RespondToSubs == true {
					var subsCurrentStream []string
					username1 := strings.Split(line, "display-name=")
					username2 := strings.Split(username1[1], ";")

					// Thank the user for subbing
					botSubResponse := strings.Replace(irc.SubResponse, "target", username2[0], -1)
					BotSendMsg(irc, botSubResponse)
					// Append new sub to a list of new subs in current session for logging
					if irc.MakeLog == true {
						subsCurrentStream = append(subsCurrentStream, username2[0])
						logLocation := "logs/NewSubs " + datesplit[0] + ".txt"
						logMessage := username2[0] + "\n"
						WriteToLog(logLocation, logMessage)
					}
				}
			}
		}

	}

}
