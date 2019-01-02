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
	url            = "http://overwatchy.com/profile/pc/us/%s"
	owPlayerRank   = "%s's SR is %d"
	owRosterHeader = "%s consists of the following players:\n"
	owRoster       = "\t%s - %d"
	owTeamRank     = "%s's SR is %d"
	owRank         = "Your SR is %d\n" +
		"Your endorsement level is %d\n" +
		"\tShot Caller: %d\n" +
		"\tTeammate: %d\n" +
		"\tSportsman: %d"

	owTeams   = make(map[string]teamStats)
	owPlayers = make(map[string]user)
)

type teamStats struct {
	Players []player `json:"players"`
	Rank    int      `json:"rank"`
}

type player struct {
	Username string `json:"username"`
	Rank     int    `json:"rank"`
}

type user struct {
	Username string `json:"username"`
	Team     string `json:"team"`
}

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

// loadUserJSON attempts to load users json file from disk into a map.
func loadUserJSON(jsonFile string, m *map[string]user) error {
	jsonByte, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		fmt.Println("Error opening json file :", err)
		return err
	}
	err = json.Unmarshal(jsonByte, m)
	if err != nil {
		fmt.Println("Error unmarshalling json file :", err)
		return err
	}
	return nil
}

// loadTeamJSON attempts to load teams json file from disk into a map.
func loadTeamJSON(jsonFile string, m *map[string]teamStats) error {
	jsonByte, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		fmt.Println("Error opening json file :", err)
		return err
	}
	err = json.Unmarshal(jsonByte, m)
	if err != nil {
		fmt.Println("Error unmarshalling json file :", err)
		return err
	}
	return nil
}

// saveMapJSON saves to map to json for persistence after the server closes
func saveUserJSON(jsonFile string, m *map[string]user) error {
	r, _ := json.MarshalIndent(m, "", "    ")
	err := ioutil.WriteFile(jsonFile, []byte(r), 0644)
	return err
}

// saveTeamJSON saves to map to json for persistence after the server closes
func saveTeamJSON(jsonFile string, m *map[string]teamStats) error {
	r, _ := json.MarshalIndent(m, "", "    ")
	err := ioutil.WriteFile(jsonFile, []byte(r), 0644)
	return err
}

// lookupStats queries the overwatch API
func lookupStats(blizzardID string) (userStats, error) {
	var us userStats

	urlString := fmt.Sprintf(url, blizzardID)
	resp, err := http.Get(urlString)
	if err != nil {
		// Could not reach API
		return us, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// Could not reach read response
		return us, err
	}
	err = json.Unmarshal(body, &us)
	// Could not reach read response
	return us, err
}

// owStats looks up the current user's OW competitive rank
func owStats(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// check if the message is "!mystats"
	if strings.HasPrefix(m.Content, "!mystats") {
		stats, err := lookupStats(owPlayers[m.Author.ID].Username)
		if err != nil {
			fmt.Println(err)
			return
		}
		response := fmt.Sprintf(
			owRank,
			stats.Competitive.Rank,
			stats.Endorsement.Level,
			stats.Endorsement.Shot.Rate,
			stats.Endorsement.Team.Rate,
			stats.Endorsement.Sports.Rate,
		)
		s.ChannelMessageSend(m.ChannelID, response)
	}
	return
}

// owTeamAvg looks up the current user's OW competitive rank
func owTeamAvg(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// check if the message is "!myteam"
	if strings.HasPrefix(m.Content, "!myteam") {
		team := owPlayers[m.Author.ID].Team
		if team != "" {
			response := fmt.Sprintf(owTeamRank, team, owTeams[team].Rank)
			s.ChannelMessageSend(m.ChannelID, response)
		} else {
			s.ChannelMessageSend(m.ChannelID, "You are not on a team or have not added your team name")
		}
	}
	return
}

// owTeamList lists all of the overwatch teams on the server
func owTeamList(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	// check if the message is "!teams"
	if strings.HasPrefix(m.Content, "!teams") {
		if len(owTeams) == 0 {
			s.ChannelMessageSend(m.ChannelID, "There are no Overwatch teams listed")
		} else {
			var roster string
			for name, team := range owTeams {
				roster = fmt.Sprintf(owRosterHeader, name)
				for _, player := range team.Players {
					playerName := strings.Replace(player.Username, "-", "#", 1)
					roster = roster + fmt.Sprintf(owRoster, playerName, player.Rank)
				}
			}
			s.ChannelMessageSend(m.ChannelID, roster)
		}
	}
	return
}

// owPlayerStats looks up a user's OW competitive rank
func owPlayerStats(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// check if the message is "!stats"
	if strings.HasPrefix(m.Content, "!stats") {
		messageArray := strings.Split(m.Message.Content, " ")
		username := strings.Replace(messageArray[1], "#", "-", 1)
		stats, err := lookupStats(username)
		if err != nil {
			fmt.Println(err)
			return
		}
		response := fmt.Sprintf(owPlayerRank, messageArray[1], stats.Competitive.Rank)
		s.ChannelMessageSend(m.ChannelID, response)
	}
	return
}

// owTeamUpdate updates up the current user's OW competitive rank
func owTeamUpdate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// check if the message is "!updateteam"
	if strings.HasPrefix(m.Content, "!updateteam") {
		var (
			tempPlayers []player
			totSR       int
		)

		teamName := owPlayers[m.Author.ID].Team
		owTeam := owTeams[teamName]
		for _, p := range owTeam.Players {
			stats, err := lookupStats(p.Username)
			if err != nil {
				fmt.Println(err)
				return
			}
			tempPlayers = append(tempPlayers, player{Username: p.Username, Rank: stats.Competitive.Rank})
			totSR += stats.Competitive.Rank
		}
		avg := totSR / len(owTeams[teamName].Players)
		owTeam.Players = tempPlayers
		owTeam.Rank = avg
		owTeams[teamName] = owTeam
		saveTeamJSON("data/owTeams.json", &owTeams)
	}
	return
}

func addOWUser(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// check if the message is "!addUser"
	if strings.HasPrefix(m.Content, "!adduser") {
		messageArray := strings.Split(m.Message.Content, " ")
		temp := owPlayers[m.Author.ID]
		temp.Username = strings.Replace(messageArray[1], "#", "-", 1)
		owPlayers[m.Author.ID] = temp
		saveUserJSON("data/owUsers.json", &owPlayers)
	}
}

func addOWTeam(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// check if the message is "!addOWUser"
	if strings.HasPrefix(m.Content, "!addteam") {
		messageArray := strings.Split(m.Message.Content, " ")
		owPlayer := owPlayers[m.Author.ID]
		owPlayer.Team = messageArray[1]
		owPlayers[m.Author.ID] = owPlayer

		team := owTeams[owPlayer.Team]
		newPlayer := player{Username: owPlayer.Username}
		team.Players = append(team.Players, newPlayer)
		owTeams[owPlayer.Team] = team
		saveUserJSON("data/owUsers.json", &owPlayers)
		saveTeamJSON("data/owTeams.json", &owTeams)
	}
}
