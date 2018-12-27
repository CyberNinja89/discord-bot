package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

var (
	token       string
	audioBuffer = make([][]byte, 0)
	owMap       = make(map[string]string)
)

func main() {
	if token == "" {
		fmt.Println("No token provided. Please run: discord-bot -t <bot token>")
		return
	}

	// Load persistent json file
	err := loadMapJSON("data/owUsers.json", &owMap)
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

	// Register airhorn as a callback for the airhorn events.
	dg.AddHandler(airhorn)
	dg.AddHandler(addOWUser)
	dg.AddHandler(owStats)

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

// loadMapJSON attempts to load json file from disk into a map.
func loadMapJSON(jsonFile string, m *map[string]string) error {
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
func saveMapJSON(jsonFile string, m *map[string]string) error {
	r, _ := json.MarshalIndent(m, "", "    ")
	err := ioutil.WriteFile(jsonFile, []byte(r), 0644)
	return err
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
