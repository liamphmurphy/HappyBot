package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func ReplaceStrings(wholeString string, old string, new string) string {
	newString := strings.Replace(wholeString, old, new, -1)
	return newString
}

func RandomInt(min int, max int) int {
	rand.Seed(time.Now().Unix())
	return min + rand.Intn(max-min)
}

func Roulette(irc *BotInfo, username string, message string) {
	fmt.Println("test lol")
	rouletteMap := make(map[int]string)

	winningKey := "red"

	optionsSplit := strings.Split(message, " ")
	fmt.Println(optionsSplit)

	var bet int
	// Bet is the points the user is risking, either a normal number or 'all'
	if optionsSplit[1] == "all" {
		bet = GetUserPoints(username)
	} else {
		bet, _ = strconv.Atoi(optionsSplit[1])

	}

	rouletteMap[1] = "red"
	rouletteMap[2] = "black"
	rouletteMap[3] = "red"
	rouletteMap[4] = "black"
	rouletteMap[5] = "red"
	rouletteMap[6] = "black"
	rouletteMap[7] = "red"
	rouletteMap[8] = "black"
	rouletteMap[9] = "red"
	rouletteMap[10] = "black"
	rouletteMap[11] = "black"
	rouletteMap[12] = "red"
	rouletteMap[13] = "black"
	rouletteMap[14] = "red"
	rouletteMap[15] = "black"
	rouletteMap[16] = "red"
	rouletteMap[17] = "black"
	rouletteMap[18] = "red"
	rouletteMap[19] = "red"
	rouletteMap[20] = "black"
	rouletteMap[21] = "red"
	rouletteMap[22] = "black"
	rouletteMap[23] = "red"
	rouletteMap[24] = "black"
	rouletteMap[25] = "red"
	rouletteMap[26] = "black"
	rouletteMap[27] = "red"
	rouletteMap[28] = "black"
	rouletteMap[29] = "black"
	rouletteMap[30] = "red"
	rouletteMap[31] = "black"
	rouletteMap[32] = "red"
	rouletteMap[33] = "black"
	rouletteMap[34] = "red"
	rouletteMap[35] = "black"
	rouletteMap[36] = "red"

	randomNumber := RandomInt(1, 36)

	userPoints := GetUserPoints(username)

	for k, v := range rouletteMap {
		if k == randomNumber {
			fmt.Println("Key: ", k)
			if v == winningKey {
				winnings := userPoints + (bet * 2)
				UpdateUserPoints(username, winnings)
				baseString := "Dang, it's your lucky day! {target} just got {value} points!"
				replaceTarget := ReplaceStrings(baseString, "{target}", username)
				replaceValue := ReplaceStrings(replaceTarget, "{value}", strconv.Itoa(winnings))
				BotSendMsg(irc.conn, irc.ChannelName, replaceValue, irc.BotName)
			} else {
				penalty := userPoints - bet
				UpdateUserPoints(username, penalty)
				baseString := "A sad day indeed... {target} just lost {value} points."
				replaceTarget := ReplaceStrings(baseString, "{target}", username)
				replaceValue := ReplaceStrings(replaceTarget, "{value}", strconv.Itoa(bet))
				BotSendMsg(irc.conn, irc.ChannelName, replaceValue, irc.BotName)
			}
		}
	}

}
