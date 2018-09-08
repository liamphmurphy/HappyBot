package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type Raffle struct {
	allUsers  []string
	allPoints []int
}

func ReplaceStrings(wholeString string, old string, new string) string {
	newString := strings.Replace(wholeString, old, new, -1)
	return newString
}

func RandomInt(min int, max int) int {
	rand.Seed(time.Now().Unix())
	return min + rand.Intn(max-min)
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
			if bet > userPoints {
				BotSendMsg(irc, "@"+username+", you betted more points than you have.")
			} else {
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

}

// GameRoot will serve as a way to check if a game is enabled in config.toml before it is run when a user calls it.
func GameRoot(irc *BotInfo, username string, usermessage string, game string, line string, allUsers []string, allPoints []int, gameRunning bool, raffleRunning bool) ([]string, []int, bool, bool) {

	if strings.Contains(usermessage, "!roulette") {
		//go GameRoot(irc, username, usermessage, "roulette")
	}

	if game == "raffle" {
		if gameRunning == true {
			if raffleRunning == true {

				if strings.Contains(usermessage, "!raffle") {
					// Prepare multi-directional channels, needed to remain values of points and users throughout main for loop iteration.
					userIn := make(chan []string)
					userOut := make(chan []string)
					pointsIn := make(chan []int)
					pointsOut := make(chan []int)
					var points int
					var err error

					// Split the user message and get points.
					messageSplit := strings.Split(usermessage, " ")
					getPointsString := messageSplit[1]
					if messageSplit[1] == "all" {
						points = GetUserPoints(username)
					} else {
						points, err = strconv.Atoi(getPointsString)
						if err != nil {
							// If an error occurs from user's message, notify them.
							BotSendMsg(irc, "@"+username+", please enter a valid number to join the raffle.")
						}
					}

					duplicateCheck := UserInSlice(username, allUsers)
					if duplicateCheck == true {
						BotSendMsg(irc, "@"+username+", you've already entered this raffle.")
					} else {
						userActualPoints := GetUserPoints(username)

						// If the points the user type is higher then what they actually have in the database, notify them and stop the raffle submission.
						if points > userActualPoints {
							BotSendMsg(irc, "@"+username+", you've submitted more points then you actually have.")
						} else {
							// Run in new goroutine, use channels to keep the data of allUsers and allPoints throughout main for loop iteration.
							go func(irc *BotInfo, username string, points int, userIn <-chan []string, userOut chan []string, pointsIn <-chan []int, pointsOut chan []int) {
								// If user is already in allUsers, meaning they are trying to enter the raffle twice, stop submission.
								allUsers = append(allUsers, username)
								userOut <- allUsers

								allPoints = append(allPoints, points)
								pointsOut <- allPoints

							}(irc, username, points, userIn, userOut, pointsIn, pointsOut)
							fmt.Println(<-userOut, <-pointsOut)
						}
					}
				} else if usermessage == "!endraffle" {
					// Initialize and set totalPoints to 0 to begin calculation of all points from user submissions.
					totalPoints := 0
					x := 0
					allUsers = RemoveStringDuplicates(allUsers)
					// For all values in allPoints, add them up together to get the winner their earnings.
					for _, v := range allPoints {
						x++
						totalPoints = totalPoints + v
					}
					// Generate a random number to pick a random index for allUsers to pick winner.
					randomElement := RandomInt(0, len(allUsers))
					pointsString := strconv.Itoa(totalPoints)
					winner := allUsers[randomElement]
					fmt.Println(allPoints)
					x = 0
					for _, v := range allUsers {
						x++
						if v != winner {
							fmt.Println("X:", x)
							getSubmission := allPoints[x-1]
							fmt.Println("Submission:", getSubmission)
							currentPoints := GetUserPoints(username)
							newPoints := currentPoints - getSubmission
							UpdateUserPoints(v, newPoints)
						}
					}
					BotSendMsg(irc, "@"+winner+" is the winner! They just won "+pointsString+" "+irc.PointsName+"!")
					// Update points user has in database if they won.
					UpdateUserPoints(username, totalPoints)
					// Set game and raffle running to false so further submissions will not be taken.
					allUsers = allUsers[:0]
					allPoints = allPoints[:0]
					gameRunning = false
					raffleRunning = false
				}
			}
		} else if gameRunning == false {
			// If game running is false and user is a moderator / broadcaster, they are probably starting a raffle! This stops normal users from starting raffles on their own.
			if CheckUserStatus(line, "moderator", irc) == "true" || CheckUserStatus(line, "broadcaster", irc) == "true" {
				if usermessage == "!startraffle" {
					BotSendMsg(irc, "A new raffle has started! Pool all your "+irc.PointsName+" together, and one winner takes it all!")
					gameRunning = true
					raffleRunning = true
				} else {

				}
			}
		}
	}
	return allUsers, allPoints, gameRunning, raffleRunning
}
