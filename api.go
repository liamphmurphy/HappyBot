package main

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type Stream struct {
	Data []struct {
		ID           string    `json:"id"`
		UserID       string    `json:"user_id"`
		GameID       string    `json:"game_id"`
		CommunityIds []string  `json:"community_ids"`
		Type         string    `json:"type"`
		Title        string    `json:"title"`
		ViewerCount  int       `json:"viewer_count"`
		StartedAt    time.Time `json:"started_at"`
	} `json:"data"`
}

func GetAllData(conn net.Conn, channel string, name string) Stream {
	client := &http.Client{}
	newChannel := SplitChannelName(channel)
	req, _ := http.NewRequest("GET", "https://api.twitch.tv/helix/streams?user_login="+newChannel, nil)
	req.Header.Set("Client-ID", "orsdrjf636aronx93hacdpk32xoi9k")
	resp, err := client.Do(req)
	if err != nil {
		panic(err.Error())
	}
	/*body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}*/
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}
	s := Stream{}
	json.Unmarshal(body, &s)

	return s

}
