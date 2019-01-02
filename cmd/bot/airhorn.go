package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	audioBuffer = make([][]byte, 0)
)

// This function will be called (due to AddHandler above) every time a new
// airhorn is created on any channel that the authenticated bot has access to.
func airhorn(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Check if the message is "!airhorn"
	if strings.HasPrefix(m.Content, "!airhorn") {
		// Find the channel that the message came from.
		c, err := s.State.Channel(m.ChannelID)
		if err != nil {
			// Could not find channel.
			return
		}

		// Find the guild for that channel.
		g, err := s.State.Guild(c.GuildID)
		if err != nil {
			// Could not find guild.
			return
		}

		// Look for the message sender in that guild's current voice states.
		for _, vs := range g.VoiceStates {
			if vs.UserID == m.Author.ID {
				err := playSound(s, g.ID, vs.ChannelID)
				if err != nil {
					fmt.Println("Error playing sound:", err)
				}
				return
			}
		}
	}
}
