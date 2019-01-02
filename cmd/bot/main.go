package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	token string

	commandsList = "`!airhorn` - play an airhorn in the current voice channel\n" +
		"`!mystats` - display your Overwatch SR\n" +
		"`!myteam` - display your team's Overwatch average SR\n" +
		"`!teams` - display team rosters on the server\n" +
		"`!stats` - look up a player's Overwatch SR\n" +
		"`!updateteam` - gets the latest rankings for your team\n" +
		"`!adduser` - associate Discord userID with Blizzard username\n" +
		"`!addteam` - associate Blizzard username with OverWatch team"
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	if token == "" {
		fmt.Println("No token provided. Please run: discord-bot -t <bot token>")
		return
	}

	// Load persistent json file
	err := loadUserJSON("data/owUsers.json", &owPlayers)
	if err != nil {
		fmt.Println("Error loading map: ", err)
		return
	}
	err = loadTeamJSON("data/owTeams.json", &owTeams)
	if err != nil {
		fmt.Println("Error loading map: ", err)
		return
	}

	// Load the sound file.
	err = loadSound("assets/airhorn.dca")
	if err != nil {
		fmt.Println("Error loading sound: ", err)
		return
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	// Register each command as a callback for events.
	dg.AddHandler(commands)
	dg.AddHandler(airhorn)
	dg.AddHandler(owStats)
	dg.AddHandler(owTeamAvg)
	dg.AddHandler(owTeamList)
	dg.AddHandler(owPlayerStats)
	dg.AddHandler(owTeamUpdate)
	dg.AddHandler(addOWUser)
	dg.AddHandler(addOWTeam)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Discord bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// loadSound attempts to load an encoded sound file from disk.
func loadSound(audioFile string) error {
	var opuslen int16

	file, err := os.Open(audioFile)
	if err != nil {
		fmt.Println("Error opening dca file :", err)
		return err
	}

	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err := file.Close()
			if err != nil {
				return err
			}
			return nil
		}

		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return err
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return err
		}

		// Append encoded pcm data to the audioBuffer.
		audioBuffer = append(audioBuffer, InBuf)
	}
}

// playSound plays the current audioBuffer to the provided channel.
func playSound(s *discordgo.Session, guildID, channelID string) (err error) {
	// Return immediately if a voice connection already exists
	// This will prevent the byte audio channel from interjecting
	for _, v := range s.VoiceConnections {
		if v.ChannelID == channelID {
			return nil
		}
	}

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

	// Start speaking.
	vc.Speaking(true)

	// Send the audioBuffer data
	for _, buff := range audioBuffer {
		vc.OpusSend <- buff
	}

	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specificed amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	// Disconnect from the provided voice channel.
	vc.Disconnect()

	return nil
}

// commands displays the all the valid discord bot commands
func commands(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// check if the message is "!commands"
	if strings.HasPrefix(m.Content, "!commands") {
		s.ChannelMessageSend(m.ChannelID, commandsList)
	}

	return
}
