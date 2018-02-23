package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

type BotInfo struct {
	ChannelName string
	ServerName  string
	BotOAuth    string
	BotName     string
	conn        net.Conn
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
	_, conferr := toml.DecodeFile("config.toml", &genconfig)
	if conferr != nil {
		fmt.Println("Can't read toml file due to:", conferr)
	}

	return &BotInfo{
		ChannelName: genconfig.ChannelName,
		ServerName:  genconfig.ServerName,
		BotOAuth:    genconfig.BotOAuth,
		BotName:     genconfig.BotName,
	}
}

func LoadGoofs() Goof {
	var goofs Goof
	_, gooferr := toml.DecodeFile("goofs.toml", &goofs)
	if gooferr != nil {
		log.Fatal(gooferr)
	}

	return goofs
}

func LoadBadWords() BadWord {
	var badwords BadWord
	_, worderr := toml.DecodeFile("badwords.toml", &badwords)
	if worderr != nil {
		log.Fatal(worderr)
	}
	return badwords
}

func LoadCustomCommands() CustomCommand {
	var customcommand CustomCommand
	_, comerr := toml.DecodeFile("commands.toml", &customcommand)
	if comerr != nil {
		log.Fatal(comerr)
	}
	return customcommand
}

func BotSendMsg(channel string, message string) {
	fmt.Println("reached function")
	//fmt.Fprintf(BotInfo.conn, "PRIVMSG %s :%s\r\n", channel, message)
}

func (bot *BotInfo) Connect() {
	var err error
	fmt.Println(bot.ServerName)
	bot.conn, err = net.Dial("tcp", bot.ServerName)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Connected to: %s\n", bot.ServerName)
}

func main() {
	irc := CreateBot()
	irc.Connect()

	fmt.Fprintf(irc.conn, "PASS %s\r\n", irc.BotOAuth)
	fmt.Fprintf(irc.conn, "NICK %s\r\n", irc.BotName)
	fmt.Fprintf(irc.conn, "JOIN %s\r\n", irc.ChannelName)
	fmt.Fprintf(irc.conn, "CAP REQ :twitch.tv/commands\r\n")
	fmt.Printf("Channel: " + irc.ChannelName + "\n")

	defer irc.conn.Close()
	reader := bufio.NewReader(irc.conn)
	proto := textproto.NewReader(reader)
	for {
		line, err := proto.ReadLine()
		//fmt.Println(line)
		if err != nil {
			break
		}
		if strings.Contains(line, "PING") {
			pong := strings.Split(line, "PING")
			fmt.Fprintf(irc.conn, "PONG %s\r\n", pong[1])
		} else if strings.Contains(line, ".tmi.twitch.tv PRIVMSG "+irc.ChannelName) {
			userdata := strings.Split(line, ".tmi.twitch.tv PRIVMSG "+irc.ChannelName)
			username := strings.Split(userdata[0], "@")
			usermessage := strings.Replace(userdata[1], " :", "", 1)
			fmt.Printf(username[1] + ": " + usermessage + "\n")

			// Make variables to load the different toml files
			goofs := LoadGoofs()
			badwords := LoadBadWords()
			customcommand := LoadCustomCommands()

			if usermessage == "hi" {
				BotSendMsg(irc.ChannelName, usermessage)
			}
			// Check for occurences of values from arrays/maps etc
			for _, v := range goofs.RepeatWords {
				if usermessage == v {
					// If value is found, because it's a goof, repeat it in chat.
					fmt.Fprintf(irc.conn, "PRIVMSG %s :%s\r\n", irc.ChannelName, v)
				}
			}

			for _, v := range badwords.BannableText {
				if usermessage == v {
					fmt.Fprintf(irc.conn, "PRIVMSG %s :%s\r\n", irc.ChannelName, "/ban "+username[1])
					fmt.Println(username[1], "has been banned.")
				}
			}

			/*for _, v := range badwords.TimeoutText {
				if usermessage == v {
					connect.Privmsg(genconfig.ChannelName, "/timeout "+event.Nick)
					connect.Privmsg(genconfig.ChannelName, "@"+event.Nick+" please watch your language.")
					fmt.Println(event.Nick + " has been timed out.")
				}
			}*/

			for _, v := range customcommand.Command {
				if usermessage == v.ComKey {
					fmt.Fprintf(irc.conn, "PRIVMSG %s :%s\r\n", irc.ChannelName, v.ComResponse)
				}
			}
			CheckForGoof := strings.Contains(usermessage, "!addgoof")
			if CheckForGoof == true {
				GoofSplit := strings.Split(usermessage, " ")
				fmt.Println(GoofSplit[1])
				f, err := os.OpenFile("commands.toml", os.O_APPEND|os.O_WRONLY, 0600)
				if err != nil {
					panic(err)
				}

				defer f.Close()
				fmt.Fprintf(f, "%s", GoofSplit[1])
				fmt.Println(GoofSplit)
			}

		}
	}
}
