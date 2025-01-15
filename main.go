package main

import (
	"discordBot/data"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	paginator "github.com/topi314/dgo-paginator"
)

type Env struct {
	DiscordToken string `json:"discordToken"`
}

var env Env
var manager *paginator.Manager

func main() {
	f, err := os.Open(".env")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	envVariables, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(envVariables, &env)

	data.Init()
	dgSession, err := discordgo.New("Bot " + env.DiscordToken)
	if err != nil {
		panic(err)
	}

	dgSession.AddHandler(messageHandler)
	dgSession.AddHandler(ready)

	// Handles the button hellscape for pagination stuff.
	manager = paginator.NewManager()
	dgSession.AddHandler(manager.OnInteractionCreate)

	dgSession.Identify.Intents = discordgo.IntentsAll

	err = dgSession.Open()
	if err != nil {
		panic(err)
	}

	defer dgSession.Close()

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
			contentType := "ANIME"
			if len(args) != 0 {
				for _, arg := range args {
					switch arg {
					case "ANIME":
						contentType = "ANIME"
					case "MANGA":
						contentType = "MANGA"
					default:
						ID = searchUserIDByName(arg)
					}
				}
			}
			perPage := 10
			numPages := 10

			Description := strings.Builder{}
			pages := []string{}
			thumbnails := []string{}

			username := getUserInfoByID(ID).Name

			for i := range numPages {
				fmt.Println(i)
				Data := getTopMediaByID(ID, contentType, i+1, perPage)
				thumbnails = append(thumbnails, Data[0].Media.CoverImage.Large)
				for _, Item := range Data {
					score := Item.Score
					if score == math.Trunc(score) {
						intScore := int(score)
						if Item.Media.Title.English != "" {
							Description.WriteString(fmt.Sprintf("**%d/10** [%s](%s) \n", intScore, Item.Media.Title.English, Item.Media.SiteURL))
						} else if Item.Media.Title.Romaji != "" {
							Description.WriteString(fmt.Sprintf("**%d/10** [%s](%s) \n", intScore, Item.Media.Title.Romaji, Item.Media.SiteURL))
						} else {
							Description.WriteString(fmt.Sprintf("**%d/10** [%s](%s) \n", intScore, Item.Media.Title.Native, Item.Media.SiteURL))
						}
					} else {
						if Item.Media.Title.English != "" {
							Description.WriteString(fmt.Sprintf("**%.1f/10** [%s](%s) \n", score, Item.Media.Title.English, Item.Media.SiteURL))
						} else if Item.Media.Title.Romaji != "" {
							Description.WriteString(fmt.Sprintf("**%.1f/10** [%s](%s) \n", score, Item.Media.Title.Romaji, Item.Media.SiteURL))
						} else {
							Description.WriteString(fmt.Sprintf("**%.1f/10** [%s](%s) \n", score, Item.Media.Title.Native, Item.Media.SiteURL))
						}
					}
				}
				pages = append(pages, Description.String())
				Description.Reset()
				if len(Data) < perPage {
					perPage = i
					break
				}
			}

			if err := manager.CreateMessage(session, message.ChannelID, &paginator.Paginator{
				PageFunc: func(page int, embed *discordgo.MessageEmbed) {
					embed.Description = pages[page]
					embed.Color = 0x00ff00
					embed.Title = fmt.Sprintf("%s's top %s", username, strings.ToLower(contentType))
					embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
						URL:    thumbnails[page],
						Width:  128,
						Height: 128,
					}
				},
				MaxPages:        len(pages),
				Expiry:          time.Now(),
				ExpiryLastUsage: true,
			}); err != nil {
				fmt.Println(err)
			}

		case "me":
			Data := getUserInfoByID(ID)
			embed := CreateProfileMediaEmbed(Data)
			session.ChannelMessageSendEmbed(message.ChannelID, &embed)

		case "u":
			if len(args) == 1 {
				ID = searchUserIDByName(args[0])
				Data := getUserInfoByID(ID)
				embed := CreateProfileMediaEmbed(Data)
				session.ChannelMessageSendEmbed(message.ChannelID, &embed)
			} else {
				session.ChannelMessageSend(message.ChannelID, "> **.u takes one argument, try again with .u __username__**")
			}
		case "wk":

		case "c":
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
	var fields = []*discordgo.MessageEmbedField{}

	if float64(Data.UserStatistics.MangaStatistics.MeanScore) == math.Trunc(float64(Data.UserStatistics.MangaStatistics.MeanScore)) {
		intScore := int(Data.UserStatistics.MangaStatistics.MeanScore) / 10
		embed := discordgo.MessageEmbedField{
			Name:   "Manga",
			Value:  fmt.Sprintf("%d Manga read\n%d Chapters read\nAverage score: %d/10\n", Data.UserStatistics.MangaStatistics.Count, Data.UserStatistics.MangaStatistics.ChaptersRead, intScore),
			Inline: true,
		}
		fields = append(fields, &embed)
	} else {
		score := Data.UserStatistics.MangaStatistics.MeanScore / 10
		embed := discordgo.MessageEmbedField{
			Name:   "Manga",
			Value:  fmt.Sprintf("%d Manga read\n%d Chapters read\nAverage score: %.1f/10\n", Data.UserStatistics.MangaStatistics.Count, Data.UserStatistics.MangaStatistics.ChaptersRead, score),
			Inline: true,
		}
		fields = append(fields, &embed)
	}

	if float64(Data.UserStatistics.AnimeStatistics.MeanScore) == math.Trunc(float64(Data.UserStatistics.AnimeStatistics.MeanScore)) {
		intScore := int(Data.UserStatistics.AnimeStatistics.MeanScore) / 10
		embed := discordgo.MessageEmbedField{
			Name:   "Anime",
			Value:  fmt.Sprintf("%d Anime watched\n%d Episodes watched\nAverage score: %d/10\n", Data.UserStatistics.AnimeStatistics.Count, Data.UserStatistics.AnimeStatistics.EpisodesWatched, intScore),
			Inline: true,
		}
		fields = append(fields, &embed)
	} else {
		score := Data.UserStatistics.AnimeStatistics.MeanScore / 10
		embed := discordgo.MessageEmbedField{
			Name:   "Anime",
			Value:  fmt.Sprintf("%d Anime watched\n%d Episodes watched\nAverage score: %.1f/10\n", Data.UserStatistics.AnimeStatistics.Count, Data.UserStatistics.AnimeStatistics.EpisodesWatched, score),
			Inline: true,
		}
		fields = append(fields, &embed)
	}

	embed := discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{},
		Color:  0x00ff00, // Green
		// Description: Data.About //These are html for some reason??????????????? don't care to parse these atm.
		Timestamp: time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
		Title:     fmt.Sprintf("%s's profile.", Data.Name),
		Fields:    fields,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL:    Data.Avatar.Medium,
			Width:  128,
			Height: 128,
		},
	}
	return embed
}
