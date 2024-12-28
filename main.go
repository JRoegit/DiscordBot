package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

const Token string = "MTMyMjAzNDI3Njc1NzQwOTgwMw.GCkuAc.ruEupTtbtoNzbWNW9BeMJqSa8KDgqTG0Tduyw4"

func main() {
	dgSession, err := discordgo.New("Bot " + Token)
	if err != nil {
		panic(err)
	}

	dgSession.AddHandler(messageHandler)

	dgSession.Identify.Intents = discordgo.IntentsGuildMessages

	err = dgSession.Open()
	if err != nil {
		panic(err)
	}

	defer dgSession.Close()

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-signalChannel
}

func messageHandler(session *discordgo.Session, message *discordgo.MessageCreate) {

	// Bot ignores messages from iteself.
	if message.Author.ID == session.State.User.ID {
		return
	}

	if message.Content == "Hello" {
		session.ChannelMessageSend(message.ChannelID, "World!")
	}

	if message.Content == "FN" {
		session.ChannelMessageSend(message.ChannelID, "BR")
	}
}
