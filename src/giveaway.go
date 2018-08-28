package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func BeginGiveaway(chatmessage string) {
	giveawayInfoSplit := strings.Split(chatmessage, " ")

	//giveawayTerm := giveawayInfoSplit[1]
	giveawayDuration, err := strconv.Atoi(giveawayInfoSplit[2])
	if err != nil {
		fmt.Println("Error making int for giveaway.")
	}

	for range time.NewTicker(time.Duration(giveawayDuration) * time.Second).C {
		fmt.Println("Hi. :)")
	}

}
