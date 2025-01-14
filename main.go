package main

import (
	"discordBot/data"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

const Token string = "MTMyMjAzNDI3Njc1NzQwOTgwMw.GCkuAc.ruEupTtbtoNzbWNW9BeMJqSa8KDgqTG0Tduyw4"
const MainChannelID string = "631979185526800458"

var AudioBuffer = make([][]byte, 0)

func main() {
	DownloadImageFromURL()
	data.Init()
	dgSession, err := discordgo.New("Bot " + Token)
	if err != nil {
		panic(err)
	}

	dgSession.AddHandler(messageHandler)
	dgSession.AddHandler(ready)

	dgSession.Identify.Intents = discordgo.IntentsAll
	// dgSession.Identify.Intents = discordgo.IntentsGuildMessages

	err = dgSession.Open()
	if err != nil {
		panic(err)
	}

	defer dgSession.Close()

	// getRecommendationsByTitle("ABC")

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

	if after, isCommand := strings.CutPrefix(message.Content, "."); isCommand {
		split := strings.SplitAfterN(after, " ", -1)
		command := strings.ToLower(strings.Trim(split[0], " "))
		args := Map(split[1:], func(item string) string { return strings.ToUpper(strings.Trim(item, " ")) })
		ID, _ := data.GetUserByDiscordID(message.Author.ID)
		switch command {
		case "link":
			if len(args) == 1 {
				anilistID := searchUserIDByName(args[0])
				_, err := data.CreateUser(message.Author.ID, anilistID)
				if err != nil {
					session.ChannelMessageSend(message.ChannelID, err.Error())
				} else {
					session.ChannelMessageSend(message.ChannelID, "Account Linked!")
				}
			} else {
				session.ChannelMessageSend(message.ChannelID, "> Invalid command formatting, try: .link *Username*")
			}

		case "top":
			// ID, err := data.GetUserByDiscordID(message.Author.ID)
			// if err == sql.ErrNoRows {
			// 	fmt.Println("No attached user")
			// 	session.ChannelMessageSend(message.ChannelID, "> No Anilist profile linked, used .link *Username* to link your profile")
			// 	return
			// }
			// if len(split) > 1 {
			// 	switch strings.ToLower(split[1]) {
			// 	case "manga":
			// 		if len(split) > 2 {

			// 		}
			// 	case "anime":
			// 	default:
			// 		session.ChannelMessageSend(message.ChannelID, `> Invalid media type, try "Manga" or "Anime"`)
			// 	}
			// }
			fmt.Printf("ID: %s", ID)
			Data := getTopMediaByID(ID, "MANGA", 1, 10)
			embed := CreateTopMediaEmbed(Data)
			session.ChannelMessageSendEmbed(message.ChannelID, &embed)

		case "me":
			Data := getUserInfoByID(ID)
			embed := CreateProfileMediaEmbed(Data)
			session.ChannelMessageSendEmbed(message.ChannelID, &embed)

		case "wk":

		case "c":
			// THIS SHIT IS DRIVING ME INSANE...

			// var data []mediaListItem
			contentType := "ANIME"
			size := 9
			if len(args) != 0 {
				for _, arg := range args {
					if arg == "ANIME" {
						contentType = "ANIME"
						fmt.Println("Found ANIME")
					}
					if arg == "MANGA" {
						contentType = "MANGA"
						fmt.Println("Found MANGA")
					}
					if arg == "4X4" {
						size = 16
						fmt.Println("Found 4x4")
					}
					if arg == "5X5" {
						size = 25
						fmt.Println("Found 5x5")
					}
				}
				fmt.Println(size)
				fmt.Println(contentType)
			}
			data := getTopMediaByID(ID, contentType, 1, size)
			var URLs []string
			for _, item := range data {
				URLs = append(URLs, item.Media.CoverImage.Large)
			}
			CreateCollageFromImages(URLs)
			f, err := os.Open("collage.jpg")
			if err != nil {
				fmt.Println(err)
			}
			session.ChannelFileSend(message.ChannelID, "collage.jpg", f)
		case "help":

		default:
			session.ChannelMessageSend(message.ChannelID, "> **Not sure about that one... try .help for a list of commands**")
		}
	}
}

func CreateTopMediaEmbed(Data []mediaListItem) discordgo.MessageEmbed {

	Description := strings.Builder{}
	for _, Item := range Data {
		Description.WriteString(fmt.Sprintf("**%.1f/10.0** [%s](%s) \n", Item.Score, Item.Media.Title.English, Item.Media.SiteURL))
	}

	embed := discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       0x00ff00, // Green
		Description: Description.String(),
		Timestamp:   time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
		Title:       "Top 10",
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL:    Data[0].Media.CoverImage.Medium,
			Width:  128,
			Height: 128,
		},
	}
	return embed
}

func CreateProfileMediaEmbed(Data user) discordgo.MessageEmbed {
	var fields = []*discordgo.MessageEmbedField{
		{
			Name:   "Manga",
			Value:  fmt.Sprintf("%d Manga read\n%d Chapters read\nAverage score: %.1f/10\n", Data.UserStatistics.MangaStatistics.Count, Data.UserStatistics.MangaStatistics.ChaptersRead, Data.UserStatistics.MangaStatistics.MeanScore),
			Inline: true,
		},
		{
			Name:   "Anime",
			Value:  fmt.Sprintf("%d Anime watched\n%d Epidsodes watched\nAverage score: %.1f/10\n", Data.UserStatistics.AnimeStatistics.Count, Data.UserStatistics.AnimeStatistics.EpisodesWatched, Data.UserStatistics.AnimeStatistics.MeanScore),
			Inline: true,
		},
	}

	embed := discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       0x00ff00, // Green
		Description: Data.About,
		Timestamp:   time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
		Title:       fmt.Sprintf("%s's profile.", Data.Name),
		Fields:      fields,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL:    Data.Avatar.Medium,
			Width:  128,
			Height: 128,
		},
	}
	return embed
}
