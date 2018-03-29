package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

type BotInfo struct {
	ChannelName    string
	ServerName     string
	BotOAuth       string
	BotName        string
	conn           net.Conn
	LongMessageCap int
}

type CustomCommand struct {
	Command []struct {
		ComKey      string
		ComResponse string
	}
}

type BadWord struct {
	BannableText []string
	TimeoutText  []string
}

type Goof struct {
	RepeatWords []string
}

func CreateBot() *BotInfo {
	var genconfig BotInfo
	_, conferr := toml.DecodeFile("config/config.toml", &genconfig)
	if conferr != nil {
		fmt.Println("Can't read toml file due to:", conferr)
	}

	return &BotInfo{
		ChannelName:    genconfig.ChannelName,
		ServerName:     genconfig.ServerName,
		BotOAuth:       genconfig.BotOAuth,
		BotName:        genconfig.BotName,
		LongMessageCap: genconfig.LongMessageCap,
	}
}

// All "Load" functions read the files for various chat features, like commands/bannable words.

/* Goofs serve no real purpose. Some chats like to have the bot 'repeat' what the user
types in, perhaps for a specific emote.*/

func LoadGoofs() Goof {
	var goofs Goof
	_, gooferr := toml.DecodeFile("config/goofs.toml", &goofs)
	if gooferr != nil {
		log.Fatal(gooferr)
	}

	return goofs
}

func LoadBadWords() BadWord {
	var badwords BadWord
	_, worderr := toml.DecodeFile("config/badwords.toml", &badwords)
	if worderr != nil {
		log.Fatal(worderr)
	}
	return badwords
}

func LoadCustomCommands() CustomCommand {
	var customcommand CustomCommand
	_, comerr := toml.DecodeFile("config/commands.toml", &customcommand)
	if comerr != nil {
		log.Fatal(comerr)
	}
	return customcommand
}

func BotSendMsg(conn net.Conn, channel string, message string) {
	//	fmt.Fprintf(conn, "PRIVMSG %s :%s\r\n", channel, message)
}

/* ConsoleInput function for reading user input in cmd line when
   program is running */

func ConsoleInput(conn net.Conn, channel string) {
	ConsoleReader := bufio.NewReader(os.Stdin)
	text, _ := ConsoleReader.ReadString('\n')

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

		if len(UsernameSplit) <= 1 { // Len if to handle index out of range error
			fmt.Println("Please type a username.")
		} else {
			BotSendMsg(conn, channel, "/ban "+UsernameSplit[1])
			fmt.Println(UsernameSplit[1] + " has been banned.")
		}
	}

}

// Connect to the Twitch IRC server
func (bot *BotInfo) Connect() {
	var err error
	bot.conn, err = net.Dial("tcp", bot.ServerName)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Connected to: %s\n", bot.ServerName)
}

// Confirm that config files are loaded
func CheckConfigs() {
	if _, err := os.Stat("config/config.toml"); err == nil {
		fmt.Println("config.toml loaded....")

	}

	if _, err := os.Stat("config/commands.toml"); err == nil {
		fmt.Println("commands.toml loaded....")

	}

	if _, err := os.Stat("config/goofs.toml"); err == nil {
		fmt.Println("goofs.toml loaded....")

	}

	if _, err := os.Stat("config/badwords.toml"); err == nil {
		fmt.Println("badwords.toml loaded....")

	}
	fmt.Printf("\n")
}

func main() {
	CheckConfigs()
	irc := CreateBot()
	irc.Connect()

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

	userargs := flag.String("--tagchat", "--tagchat", "detailed view")
	fmt.Println(*userargs)

	for {
		line, err := proto.ReadLine()
		if err != nil {
			break
		}

		// Run ConsoleInput on new thread
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
			/*	if *userargs == "--tagchat" {
					fmt.Println(line)
				} else {

				}*/

			//	fmt.Println("Character count of chat message: ", len(usermessage))

			// Make variables to load the different toml files
			goofs := LoadGoofs()
			badwords := LoadBadWords()
			customcommand := LoadCustomCommands()

			if len(usermessage) > irc.LongMessageCap {
				fmt.Println("Very long message detected.")
				botresponse := "/timeout " + username[1] + " 1" + "Message over max character limit."
				BotSendMsg(irc.conn, irc.ChannelName, botresponse)
				BotSendMsg(irc.conn, irc.ChannelName, "@"+username[1]+" please shorten your message")

			}

			// Check for occurences of values from arrays/maps etc
			for _, v := range goofs.RepeatWords {
				if usermessage == v {
					// If value is found, because it's a goof, repeat it in chat.
					BotSendMsg(irc.conn, irc.ChannelName, usermessage)
				}
			}

			for _, v := range badwords.BannableText {
				if usermessage == v {
					fmt.Println(username[1], "has been banned.")
					BotSendMsg(irc.conn, irc.ChannelName, usermessage)
				}
			}

			for _, v := range customcommand.Command {
				if usermessage == v.ComKey {
					BotSendMsg(irc.conn, irc.ChannelName, v.ComResponse)
				}
			}
			CheckForGoof := strings.Contains(usermessage, "!addgoof")
			if CheckForGoof == true {
				GoofSplit := strings.Split(usermessage, " ")
				fmt.Println(GoofSplit[1])
				f := append(goofs.RepeatWords, GoofSplit[1])

				//defer f.Close()
				fmt.Println(f)
				fmt.Println(GoofSplit)
				file, _ := os.OpenFile("config/goofs.toml", os.O_WRONLY|os.O_APPEND, 0644)
				defer file.Close()
				fmt.Fprintf(file, `"%s"`, GoofSplit[1])
			}

		} else if strings.Contains(line, "msg-param-sub-plan") {
			/*		line := string(line)

								subuser := strings.TrimPrefix(line, "USERNOTICE")
								fmt.Println(subuser)
					fmt.Println(strings.SplitAfter(line, "color"))
			*/
		}

	}

}
