package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

var discordToken = ""
var mxRoom = ""
var mxToken = ""

func main() {

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	discordToken = viper.Get("discordToken").(string)
	mxRoom = viper.Get("mxRoom").(string)
	mxToken = viper.Get("mxToken").(string)

	sendMatrixMessage(mxRoom, mxToken, "Discord Voice Monitor booted.")

	// Using the matrix-bridge-bot token, enabled "Presence Intent" unclear if that is needed or not
	discord, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	// Required for voice state events to be sent to the client over the web socket
	discord.Identify.Intents = discordgo.IntentsGuildVoiceStates
	err = discord.Open()
	if err != nil {
		fmt.Println("error opening connection:", err)
		return
	}
	discord.AddHandler(voice)

	// Wait until CTRL-C or other term signal is received.
	fmt.Println("Running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	discord.Close()
}

// voice logs voice channel joins to stdout and matrix
func voice(s *discordgo.Session, event *discordgo.VoiceStateUpdate) {
	//fmt.Println("VOICE STATE EVENT DETECTED")

	m, _ := s.GuildMember(event.VoiceState.GuildID, event.UserID)
	nc, _ := s.Channel(event.ChannelID)

	// ChannelID is populated when joining a voice channel, but not when leaving.
	if event.ChannelID != "" {
		msg := "'" + m.Nick + "' has entered '" + nc.Name + "'"
		sendMatrixMessage(mxRoom, mxToken, msg)
		fmt.Println(msg)
	} else {

		// BeforeUpdate is not populated unless you were listening when the user originally connected.
		if event.BeforeUpdate != nil {
			oc, _ := s.Channel(event.BeforeUpdate.ChannelID)
			fmt.Println("'" + m.Nick + "' has left '" + oc.Name + "'")
		} else {
			fmt.Println("'" + m.Nick + "' has left 'UNKNOWN'")
		}
	}
}

// sendMatrixMessage sends an m.text message body to a specified non e2e encrypted matrix room
func sendMatrixMessage(mxRoom string, mxToken string, msg string) int {

	rand.Seed(time.Now().UnixNano())
	mxMsgId := fmt.Sprint(rand.Intn(1000000))

	type Mtext struct {
		Msgtype string `json:"msgtype"`
		Body    string `json:"body"`
	}

	mText := Mtext{
		Msgtype: "m.text",
		Body:    msg,
	}

	client := &http.Client{}

	json, err := json.Marshal(mText)
	if err != nil {
		panic(err)
	}

	// set the HTTP method, url, and request body
	req, err := http.NewRequest(http.MethodPut, mxRoom+"/send/m.room.message/"+mxMsgId+"?access_token="+mxToken, bytes.NewBuffer(json))
	if err != nil {
		panic(err)
	}

	// set the request header Content-Type for json
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	return resp.StatusCode
}
