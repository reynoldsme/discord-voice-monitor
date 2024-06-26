package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

var discordToken = ""
var mxRoom = ""
var mxToken = ""
var friends []string
var activityInterval = time.Duration(time.Second * 0)

var lastActivityCheck = time.Now().Add(time.Second * time.Duration(-activityInterval))

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
	friends = viper.GetStringSlice("friends")
	activityInterval = time.Duration(viper.GetInt("activityinterval")) * time.Second

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
		fmt.Println("Error opening connection:", err)
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
	if event.ChannelID != "" && (event.BeforeUpdate == nil || (event.VoiceState.SelfMute == event.BeforeUpdate.SelfMute)) {
		friendStatus := getFriendSteamStatus(friends)
		msg := "'" + m.Nick + "' has entered '" + nc.Name + friendStatus
		sendMatrixMessage(mxRoom, mxToken, msg)
		fmt.Println(msg)
		//fmt.Println(friends)
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

// getFriendSteamStatus takes a slice of friend steam ids and returns a formatted list who is playing what
func getFriendSteamStatus(friends []string) string {

	games := ""
	delta := time.Now().Sub(lastActivityCheck)

	if delta >= activityInterval {
		lastActivityCheck = time.Now()
		for _, friend := range friends {

			response, err := http.Get("https://steamcommunity.com/id/" + friend)
			//response, err := http.Get("https://leptco.com/steam.html")
			if err != nil {
				fmt.Println("Failed to request steam in-game status for " + friend)
				continue
			}
			defer response.Body.Close()
			doc, _ := goquery.NewDocumentFromReader(response.Body)
			doc.Find(".profile_in_game_name").Each(func(i int, s *goquery.Selection) {
				game := strings.ReplaceAll(strings.ReplaceAll(s.Text(), "\n", ""), "\t", "")
				if game != "" {
					games += "\n* " + friend + " is playing: " + game
				}

			})

		}
	}
	fmt.Println(games)
	return games
}
