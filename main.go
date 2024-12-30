package main

import (
	"discordBot/data"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

const Token string = "MTMyMjAzNDI3Njc1NzQwOTgwMw.GCkuAc.ruEupTtbtoNzbWNW9BeMJqSa8KDgqTG0Tduyw4"
const MainChannelID string = "631979185526800458"

var AudioBuffer = make([][]byte, 0)

func main() {
	data.Init()
	dgSession, err := discordgo.New("Bot " + Token)
	if err != nil {
		panic(err)
	}

	loadAudio()

	dgSession.AddHandler(messageHandler)
	dgSession.AddHandler(userHandler)
	dgSession.AddHandler(presenceHandler)
	dgSession.AddHandler(ready)

	dgSession.Identify.Intents = discordgo.IntentsAll
	// dgSession.Identify.Intents = discordgo.IntentsGuildMessages

	err = dgSession.Open()
	if err != nil {
		panic(err)
	}

	defer dgSession.Close()

	getRecommendationsByTitle("ABC")

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-signalChannel
}

func ready(session *discordgo.Session, event *discordgo.Ready) {
	session.UpdateCustomStatus("Writen in Go")
}

func messageHandler(session *discordgo.Session, message *discordgo.MessageCreate) {

	// Bot ignores messages from iteself.
	if message.Author.ID == session.State.User.ID {
		return
	}

	if message.Content == "Hello" {
		session.ChannelMessageSend(message.ChannelID, "World!")
		fmt.Println(message.ChannelID)
	}

	if message.Content == "FN" {
		session.ChannelMessageSend(message.ChannelID, "BR")
	}

	if message.Content == "Join" {
		channel, err := session.State.Channel(message.ChannelID)
		if err != nil {
			panic(err)
		}

		guild, err := session.State.Guild(channel.GuildID)
		if err != nil {
			panic(err)
		}

		for _, vs := range guild.VoiceStates {
			if vs.UserID == message.Author.ID {
				err = playAudio(session, guild.ID, vs.ChannelID)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func userHandler(session *discordgo.Session, user *discordgo.GuildMemberAdd) {
	session.ChannelMessageSend(MainChannelID, fmt.Sprintf("Welcome %s", user.DisplayName()))
}

func presenceHandler(session *discordgo.Session, presence *discordgo.PresenceUpdate) {
	session.ChannelMessageSend(MainChannelID, fmt.Sprintf("%s %s", presence.User.Username, presence.Status))
}

func loadAudio() error {

	file, err := os.Open("airhorn.dca")
	if err != nil {
		fmt.Println("Error opening sound file.")
		return err
	}

	var opusLen int16
	for {
		err = binary.Read(file, binary.LittleEndian, &opusLen)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err := file.Close()
			if err != nil {
				return err
			}
			return nil
		}

		if err != nil {
			fmt.Println("Error reading dca file.")
			return err
		}

		inBuf := make([]byte, opusLen)
		err := binary.Read(file, binary.LittleEndian, &inBuf)
		if err != nil {
			fmt.Println("Error reading dca file.")
			return err
		}
		AudioBuffer = append(AudioBuffer, inBuf)
	}
}

func playAudio(session *discordgo.Session, guildID string, channelID string) error {
	vc, err := session.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

	time.Sleep(250 * time.Millisecond)

	vc.Speaking(true)

	for _, buff := range AudioBuffer {
		vc.OpusSend <- buff
	}

	vc.Speaking(false)

	time.Sleep(250 * time.Millisecond)

	vc.Disconnect()

	return nil
}
