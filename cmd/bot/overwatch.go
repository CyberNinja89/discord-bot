package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	url    = `http://overwatchy.com/profile/pc/us/%s`
	owRank = `Your OW competitive rank is %d`
)

type userStats struct {
	Username    string  `json:"username"`
	Level       int     `json:"level"`
	Endorsement endorse `json:"endorsement"`
	Private     bool    `json:"private"`
	Competitive comp    `json:"competitive"`
}

type endorse struct {
	Sports rate `json:"sportsmanship"`
	Shot   rate `json:"shotcaller"`
	Team   rate `json:"teammate"`
	Level  int  `json:"level"`
}

type comp struct {
	Rank int `json:"rank"`
}

type rate struct {
	Rate int `json:"rate"`
}

// owStats looks up the current user's OW competitive rank
func owStats(s *discordgo.Session, m *discordgo.MessageCreate) {
	var us userStats
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// check if the message is "!owstats"
	if strings.HasPrefix(m.Content, "!owstats") {
		urlString := fmt.Sprintf(url, owMap[m.Author.ID])
		resp, err := http.Get(urlString)
		if err != nil {
			fmt.Println(err)
			// Could not reach API
			return
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			// Could not reach read response
			return
		}
		err = json.Unmarshal(body, &us)
		if err != nil {
			fmt.Println(err)
			// Could not reach read response
			return
		}
		response := fmt.Sprintf(owRank, us.Competitive.Rank)
		_, _ = s.ChannelMessageSend(m.ChannelID, response)
	}

	return
}

func addOWUser(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// check if the message is "!addOWUser"
	if strings.HasPrefix(m.Content, "!addowuser") {
		messageArray := strings.Split(m.Message.Content, " ")
		owMap[m.Author.ID] = strings.Replace(messageArray[1], "#", "-", 1)
		saveMapJSON("assets/owUsers.json", &owMap)
	}
}
