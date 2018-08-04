package main

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
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

type User struct {
	Data []struct {
		ID string `json:"id"`
	}
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
	//newChannel := SplitChannelName(channel)
	data := ApiCall(conn, channel, "GET", "https://api.twitch.tv/helix/streams?user_login=lil_lexi")
	// Create a new object of Stream and unmarshal JSON into it
	s := Stream{}
	json.Unmarshal(data, &s)
	return s

}

func PostStreamData(conn net.Conn, channel string, info string, value string) {
	newChannel := SplitChannelName(channel)
	data := ApiCall(conn, channel, "GET", "https://api.twitch.tv/helix/users?login="+newChannel)

	s := User{}
	json.Unmarshal(data, &s)

	newData := url.Values{}
	newData.Set(info, value)

	for _, val := range s.Data {
		//b := bytes.NewBufferString(newData.Encode())
		channelID := val.ID
		resp, _ := http.PostForm("https://api.twitch.tv/kraken/streams/"+channelID, newData)
		resp.Header.Set("Client-ID", "orsdrjf636aronx93hacdpk32xoi9k")
	}

	/*_, err := http.PostForm("https://api.twitch.tv/helix/streams?user_login="+newChannel, streamData)
	if err != nil {
		fmt.Println(err)
	}*/

}

func GetViewers(conn net.Conn, channel string) Viewers {
	newChannel := SplitChannelName(channel)
	data := ApiCall(conn, channel, "GET", "https://tmi.twitch.tv/group/user/"+newChannel+"/chatters")
	// Create a new object of Stream and unmarshal JSON into it
	s := Viewers{}
	json.Unmarshal(data, &s)
	return s

}

func PostPasteBin(apikey string, com map[string]*CustomCommand) string {
	//com := LoadCommands()
	//client := &http.Client{}

	file, _ := os.OpenFile("comstext.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// For each key and value in com map, append it to file for string conversion later.
	for k, v := range com {
		file.WriteString(k + " --- " + v.CommandResponse + "\n")
	}

	s, err := ioutil.ReadFile("comstext.txt")
	if err != nil {
		panic(err)
	}
	str := string(s)

	pasteBinData := url.Values{
		"api_dev_key":    {apikey},
		"api_option":     {"paste"},
		"api_paste_code": {str},
	}
	file.Close()
	err = os.Remove("comstext.txt")

	if err != nil {
		panic(err)
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
