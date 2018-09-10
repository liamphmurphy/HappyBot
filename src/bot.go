package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/textproto"
	"os"
	"strconv"
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
	LetModeratorsUseAllCommands bool
	CasterMessage               string
	CheckLongMessageCap         bool
	WarnUsersForLongMsg         bool
	WarnAmountLongMsg           int
	WarnTimeoutMsgLength        int
	LongMessageCap              int
	StreamerTimeToggle          bool
	MakeLog                     bool
	RespondToSubs               bool
	SubResponse                 string
	PurgeForLinks               bool
	WarnUserForLinks            bool
	WarnAmountLinks             int
	WarnTimeoutLinkLength       int
	LinkChecks                  []string
	HydrateOn                   bool
	HydrateTime                 time.Duration
	HydrateMessage              string
	SendMessages                bool
	PastebinKey                 string
	PointsSystemEnabled         bool
	PointsName                  string
	PointsValueModifier         int
	PointsIncrementTime         time.Duration
	PointsMessage               string
	GamesEnabled                bool
	DuelsEnabled                bool
	RouletteEnabled             bool
	RouletteWinMessages         []string
	RouletteLossMessages        []string
	EightBallEnabled            bool
	EightBallMessages           []string
	conn                        net.Conn
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

type Warning struct {
	Amount int
	Reason string
}

func CreateBot() *BotInfo {
	var genConfig BotInfo
	_, confErr := toml.DecodeFile("../config/config.toml", &genConfig)
	if confErr != nil {
		fmt.Println("Can't read toml file due to: ", confErr)
	}

	// If an arg from user is --quiet or -q, stop bot from sending any messages to chat. This is mainly using for debugging and testing.
	sendMessages := true
	for _, v := range os.Args {
		if v == "--quiet" {
			fmt.Println("Quiet mode activated")
			sendMessages = false
		} else if v == "-q" {
			sendMessages = false
		}
	}

	if sendMessages == false {
		fmt.Println("\n\tI AM QUIET... I AM... THE ABSENCE OF WORDS.")
		fmt.Println("\tNo chat messages from bot will be sent.\n")
	}

	return &BotInfo{
		ChannelName:                 genConfig.ChannelName,
		ServerName:                  genConfig.ServerName,
		BotOAuth:                    genConfig.BotOAuth,
		BotName:                     genConfig.BotName,
		WebAppGUIEnabled:            genConfig.WebAppGUIEnabled,
		LetModeratorsUseAllCommands: genConfig.LetModeratorsUseAllCommands,
		CasterMessage:               genConfig.CasterMessage,
		CheckLongMessageCap:         genConfig.CheckLongMessageCap,
		WarnUsersForLongMsg:         genConfig.WarnUsersForLongMsg,
		WarnAmountLongMsg:           genConfig.WarnAmountLongMsg,
		WarnTimeoutMsgLength:        genConfig.WarnTimeoutMsgLength,
		LongMessageCap:              genConfig.LongMessageCap,
		StreamerTimeToggle:          genConfig.StreamerTimeToggle,
		MakeLog:                     genConfig.MakeLog,
		RespondToSubs:               genConfig.RespondToSubs,
		SubResponse:                 genConfig.SubResponse,
		PurgeForLinks:               genConfig.PurgeForLinks,
		WarnUserForLinks:            genConfig.WarnUserForLinks,
		WarnAmountLinks:             genConfig.WarnAmountLinks,
		WarnTimeoutLinkLength:       genConfig.WarnTimeoutLinkLength,
		LinkChecks:                  genConfig.LinkChecks,
		HydrateOn:                   genConfig.HydrateOn,
		HydrateTime:                 genConfig.HydrateTime,
		HydrateMessage:              genConfig.HydrateMessage,
		SendMessages:                sendMessages,
		PastebinKey:                 genConfig.PastebinKey,
		PointsSystemEnabled:         genConfig.PointsSystemEnabled,
		PointsName:                  genConfig.PointsName,
		PointsValueModifier:         genConfig.PointsValueModifier,
		PointsIncrementTime:         genConfig.PointsIncrementTime,
		PointsMessage:               genConfig.PointsMessage,
		GamesEnabled:                genConfig.GamesEnabled,
		DuelsEnabled:                genConfig.DuelsEnabled,
		RouletteEnabled:             genConfig.RouletteEnabled,
		RouletteWinMessages:         genConfig.RouletteWinMessages,
		RouletteLossMessages:        genConfig.RouletteLossMessages,
		EightBallEnabled:            genConfig.EightBallEnabled,
		EightBallMessages:           genConfig.EightBallMessages,
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
func TimeCommands(irc *BotInfo, TimeSetting string, channel string, name string, userName string) string {
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
				newMessage := "@" + userName + " " + newSplit + " seconds."

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

// Bans user with provided userName.
func BanUser(irc *BotInfo, user string) {
	fmt.Println(user, "has been banned.")
	BotSendMsg(irc, "/ban "+user)
}

// Time out user with provided userName.
func TimeOutUser(irc *BotInfo, user string, length int) {
	stringLength := strconv.Itoa(length)
	BotSendMsg(irc, "/timeout "+user+" "+stringLength)
}

// Purge user with provided userName.
func PurgeUser(irc *BotInfo, user string) {
	fmt.Println(user, "has been purged.")
	BotSendMsg(irc, "/timeout "+user+" 1")
}

// Function to add a new quote and return a map of quotes, including new one.
func AddQuote(irc *BotInfo, message string, userMessage string, name string) map[string]string {
	database := InitializeDB()

	quoteSplit := strings.Split(userMessage, "!addquote ")
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
func AddGoof(userMessage string) Goof {
	database := InitializeDB()
	// Split data to separate userName from value to use as new goof
	GoofSplit := strings.Split(userMessage, " ")
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

func MakeWarning(amount int, reason string) *Warning {
	return &Warning{
		Amount: amount,
		Reason: reason,
	}
}

func RemoveStringDuplicates(slice []string) []string {
	m := make(map[string]bool)
	for _, v := range slice {
		if _, ok := m[v]; ok {

		} else {
			m[v] = true
		}
	}

	var result []string
	for v := range m {
		result = append(result, v)
	}
	return result
}

// Check to see if user is in the permitted slice
func UserInSlice(user string, perm []string) bool {
	for _, userName := range perm {
		if userName == user {
			return true
		}
	}
	return false
}

// For sake of removing after a link is posted, iterate through the slice and get the element index
func GetSlicePosition(user string, perm []string) int {
	x := 0
	for _, userName := range perm {
		x++
		if userName == user {
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

func Giveaway(irc *BotInfo, userName string, message string, state string, users []string, running bool) (bool, []string, string) {
	var entryTerm string
	if state == "new" {
		running = true
		messageSplit := strings.Split(message, " ")
		entryTerm = messageSplit[1]
		BotSendMsg(irc, "A new giveaway has started! Type "+entryTerm+" in chat to enter!")
	} else if state == "end" {
		running = false
		rand.Seed(time.Now().Unix())
		winner := users[rand.Intn(len(users))]
		users = users[:0]
		BotSendMsg(irc, winner+" is the winner!")

	} else if state == "entry" {
		users = append(users, userName)
	}

	return running, users, entryTerm
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

	gameRunning := false

	//var allPoints []int
	var allUsers []string
	var allPoints []int
	raffleRunning := false
	// Prepare variable for users permitted to post links.
	var permUsers []string

	allDuels := make(map[string]*Duel)

	// WarnMap consists of all users who are warned in the key. There is a two variable value pair indicating warning amount, and what they are being warned for.
	warnMap := make(map[string]*Warning)

	// Prepare variables needed for giveaways.
	giveawayEntryTerm := "giveawayisnil"
	var giveawayRunning bool
	var giveawayUsers []string
	/* Run ConsoleInput on new thread
	Allows user to type commands into terminal window */
	go ConsoleInput(irc)
	for {
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
			userData := strings.Split(line, ".tmi.twitch.tv PRIVMSG "+irc.ChannelName)
			splitData := strings.Split(userData[0], "@")
			userName := splitData[2]
			userMessage := strings.Replace(userData[1], " :", "", 1)
			userMessageSplit := strings.Split(userMessage, " ")
			potentialCommand := userMessageSplit[0]

			// Display the whole cleaned up message
			fmt.Println(userName + ": " + userMessage)

			// Check if user set MakeLog in config.toml to true, if so, run
			if irc.MakeLog == true {
				// Use current date to mark which day the chat log is for
				currenttime := time.Now()
				datestring := currenttime.String()
				datesplit := strings.Split(datestring, " ")
				loglocation := "logs/chat/" + datesplit[0] + ".txt"
				logmessage := (userName + ": " + userMessage + "\n")
				WriteToLog(loglocation, logmessage)
			}

			if irc.GamesEnabled == true {
				if strings.Contains(userMessage, "raffle") {
					allUsers, allPoints, gameRunning, raffleRunning, _ = GameRoot(irc, userName, userMessage, "raffle", line, allUsers, allPoints, raffleRunning, gameRunning, allDuels)
				} else if potentialCommand == "!8ball" {
					_, _, _, _, _ = GameRoot(irc, userName, userMessage, "8ball", line, allUsers, allPoints, raffleRunning, gameRunning, allDuels)
				} else if potentialCommand == "!duel" {
					_, _, _, _, allDuels = GameRoot(irc, userName, userMessage, "duel", line, allUsers, allPoints, raffleRunning, gameRunning, allDuels)
				} else if potentialCommand == "!roulette" {
					_, _, _, _, _ = GameRoot(irc, userName, userMessage, "roulette", line, allUsers, allPoints, raffleRunning, gameRunning, allDuels)
				}
			}

			/* If a moderator or broadcaster, their message may be to edit / add / delete a command.
			If they are, run CreateCommands, which updates these values for the chat to use. This may not be the optimal solution,
			but it makes it so normal users' messages aren't checked. */
			if CheckUserStatus(line, "moderator", irc) == "true" || CheckUserStatus(line, "broadcaster", irc) == "true" {
				com, quotes, goofs.GoofSlice = CreateCommands(irc, userMessage, potentialCommand, com, quotes, badwords, goofs, database, line)

				if potentialCommand == "!newgiveaway" {
					giveawayRunning, giveawayUsers, giveawayEntryTerm = Giveaway(irc, userName, userMessage, "new", giveawayUsers, false)
				} else if userMessage == "!endgiveaway" {
					if giveawayRunning == true {
						giveawayRunning, giveawayUsers, _ = Giveaway(irc, userName, userMessage, "end", giveawayUsers, true)
					} else {
						BotSendMsg(irc, "Giveaway is not running.")
					}
				}
			}

			if userMessage == giveawayEntryTerm {
				if giveawayRunning == true {
					_, giveawayUsers, _ = Giveaway(irc, userName, userMessage, "entry", giveawayUsers, true)
				}
			}

			// Default commands for the bot are put in DefaultCommands. Things like !caster, !permit etc can be seen there.
			permUsers, warnMap = DefaultCommands(irc, userName, userMessage, potentialCommand, warnMap, line, com, quotes, badwords, goofs, permUsers, database)

			go UserCommands(irc, userName, userMessage, line, com, quotes, badwords, goofs, permUsers, giveawayEntryTerm, giveawayUsers, database)

		} else if strings.Contains(line, "USERNOTICE") {
			currenttime := time.Now()
			datestring := currenttime.String()
			datesplit := strings.Split(datestring, " ")
			// user variables used to split the twitch tag string to get the userName
			if strings.Contains(line, "msg-param-sub-plan") {
				if irc.RespondToSubs == true {
					var subsCurrentStream []string
					userName1 := strings.Split(line, "display-name=")
					userName2 := strings.Split(userName1[1], ";")

					// Thank the user for subbing
					botSubResponse := strings.Replace(irc.SubResponse, "target", userName2[0], -1)
					BotSendMsg(irc, botSubResponse)
					// Append new sub to a list of new subs in current session for logging
					if irc.MakeLog == true {
						subsCurrentStream = append(subsCurrentStream, userName2[0])
						logLocation := "logs/NewSubs " + datesplit[0] + ".txt"
						logMessage := userName2[0] + "\n"
						WriteToLog(logLocation, logMessage)
					}
				}
			}
		}

	}

}
