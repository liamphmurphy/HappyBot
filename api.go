package main

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
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

type Viewers struct {
	ChatterCount int `json:"chatter_count"`
	Chatters     struct {
		CurrentModerators []string `json:"moderators"`
		CurrentViewers    []string `json:"viewers"`
	} `json:"chatters"`
}

func ApiCall(conn net.Conn, channel string, httpType string, apiUrl string) []byte {
	client := &http.Client{}
	// Split the # from channel name to be used for URL in GET
	req, _ := http.NewRequest(httpType, apiUrl, nil)
	req.Header.Set("Client-ID", "orsdrjf636aronx93hacdpk32xoi9k")
	resp, err := client.Do(req)
	if err != nil {
		panic(err.Error())
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}
	return body
}

func StreamData(conn net.Conn, channel string) Stream {
	// Create a http client
	newChannel := SplitChannelName(channel)
	data := ApiCall(conn, channel, "GET", "https://api.twitch.tv/helix/streams?user_login="+newChannel)
	// Create a new object of Stream and unmarshal JSON into it
	s := Stream{}
	json.Unmarshal(data, &s)
	return s

}

func GetViewers(conn net.Conn, channel string) Viewers {
	//newChannel := SplitChannelName(channel)
	data := ApiCall(conn, channel, "GET", "https://tmi.twitch.tv/group/user/dansgaming/chatters")
	// Create a new object of Stream and unmarshal JSON into it
	s := Viewers{}
	json.Unmarshal(data, &s)
	return s

}

func PostPasteBin(apikey string) string {
	//com := LoadCommands()
	//client := &http.Client{}
	pasteBinData := url.Values{
		"api_dev_key":    {apikey},
		"api_option":     {"paste"},
		"api_paste_code": {"test"},
	}
	resp, err := http.PostForm("https://pastebin.com/api/api_post.php", pasteBinData)
	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}
	bodyString := string(body)
	return bodyString
}
