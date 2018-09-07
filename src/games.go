package main

import (
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

func RafflePoints(irc *BotInfo, username string, message string, participating map[string]chan int) {

	points := strings.Split(message, " ")
	pointsInt, _ := strconv.Atoi(points[1])
	participating[username] <- pointsInt

}

func Roulette(irc *BotInfo, username string, message string) {
	rouletteMap := make(map[int]string)

	winningKey := "red"

	optionsSplit := strings.Split(message, " ")

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
			if v == winningKey {
				winnings := userPoints + (bet * 2)
				UpdateUserPoints(username, winnings)
				randomMessageIndex := RandomInt(0, len(irc.RouletteWinMessages))
				baseString := irc.RouletteWinMessages[randomMessageIndex]
				replaceTarget := ReplaceStrings(baseString, "{target}", username)
				replaceValue := ReplaceStrings(replaceTarget, "{value}", strconv.Itoa(winnings))
				replaceCurrency := ReplaceStrings(replaceValue, "{currency}", irc.PointsName)
				BotSendMsg(irc, replaceCurrency)
			} else {
				penalty := userPoints - bet
				UpdateUserPoints(username, penalty)
				randomMessageIndex := RandomInt(0, len(irc.RouletteLossMessages))
				baseString := irc.RouletteLossMessages[randomMessageIndex]
				replaceTarget := ReplaceStrings(baseString, "{target}", username)
				replaceValue := ReplaceStrings(replaceTarget, "{value}", strconv.Itoa(bet))
				replaceCurrency := ReplaceStrings(replaceValue, "{currency}", irc.PointsName)
				BotSendMsg(irc, replaceCurrency)
			}
		}
	}

}

// GameRoot will serve as a way to check if a game is enabled in config.toml before it is run when a user calls it.
func GameRoot(irc *BotInfo, username string, message string, game string) {
	if game == "roulette" {
		if irc.RouletteEnabled == true {
			Roulette(irc, username, message)
		} else {
			BotSendMsg(irc, "@"+username+", that game is not enabled.")
		}
	}

}
