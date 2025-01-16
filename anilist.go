package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

type anilistID struct {
	ID string `json:"id"`
}

type topLevel struct {
	Data aniData `json:"data"`
}

type aniData struct {
	Page  page  `json:"Page"`
	User  user  `json:"User"`
	Media media `json:"Media"`
}

type page struct {
	Recs      []mediaRecItem  `json:"recommendations"`
	MediaList []mediaListItem `json:"mediaList"`
}

type mediaListItem struct {
	Media       media     `json:"media"`
	Score       float64   `json:"score"`
	StartedAt   fuzzyDate `json:"startedAt"`
	CompletedAt fuzzyDate `json:"completedAt"`
}

type fuzzyDate struct {
	Day   int `json:"day"`
	Month int `json:"month"`
	Year  int `json:"year"`
}

type user struct {
	ID             int           `json:"id"`
	Avatar         avatar        `json:"avatar"`
	About          string        `json:"about"`
	Name           string        `json:"name"`
	SiteURL        string        `json:"siteUrl"`
	UserStatistics userStatistic `json:"statistics"`
}

type userStatistic struct {
	AnimeStatistics animeStatistic `json:"anime"`
	MangaStatistics mangaStatistic `json:"manga"`
}

type mangaStatistic struct {
	ChaptersRead int     `json:"chaptersRead"`
	Count        int     `json:"count"`
	MeanScore    float32 `json:"meanScore"`
}

type animeStatistic struct {
	EpisodesWatched int     `json:"episodesWatched"`
	Count           int     `json:"count"`
	MeanScore       float32 `json:"meanScore"`
}

type avatar struct {
	Large  string `json:"large"`
	Medium string `json:"medium"`
}

type mediaRecItem struct {
	Media media `json:"mediaRecommendation"`
}

type media struct {
	Id           int        `json:"id"`
	Description  string     `json:"description"`
	CoverImage   coverImage `json:"coverImage"`
	Title        title      `json:"title"`
	SiteURL      string     `json:"siteUrl"`
	AverageScore int        `json:"averageScore"`
	Genres       []string   `json:"genres"`
}

type coverImage struct {
	Large  string `json:"large"`
	Medium string `json:"medium"`
}

type title struct {
	English string `json:"english"`
	Native  string `json:"native"`
	Romaji  string `json:"romaji"`
}

func (Media *media) ToCleanString() string {
	return bluemonday.UGCPolicy().Sanitize(Media.Description)
}

func (Title title) ToString() string {
	if Title.English != "" {
		return Title.English
	} else if Title.Romaji != "" {
		return Title.Romaji
	} else {
		return Title.Native
	}
}

func (Item *mediaListItem) ToRatingString() string {
	score := Item.Score
	if score == math.Trunc(score) {
		intScore := int(score)
		if Item.Media.Title.English != "" {
			return fmt.Sprintf("**%d/10** [%s](%s) \n", intScore, Item.Media.Title.English, Item.Media.SiteURL)
		} else if Item.Media.Title.Romaji != "" {
			return fmt.Sprintf("**%d/10** [%s](%s) \n", intScore, Item.Media.Title.Romaji, Item.Media.SiteURL)
		} else {
			return fmt.Sprintf("**%d/10** [%s](%s) \n", intScore, Item.Media.Title.Native, Item.Media.SiteURL)
		}
	} else {
		if Item.Media.Title.English != "" {
			return fmt.Sprintf("**%.1f/10** [%s](%s) \n", score, Item.Media.Title.English, Item.Media.SiteURL)
		} else if Item.Media.Title.Romaji != "" {
			return fmt.Sprintf("**%.1f/10** [%s](%s) \n", score, Item.Media.Title.Romaji, Item.Media.SiteURL)
		} else {
			return fmt.Sprintf("**%.1f/10** [%s](%s) \n", score, Item.Media.Title.Native, Item.Media.SiteURL)
		}
	}
}

const aniListEndPoint string = "https://graphql.anilist.co"

func searchUserIDByName(userName string) string {
	reqQuery := strings.NewReader(fmt.Sprintf(`{
		"query": "query Query($search: String) {User(search: $search) {id name}}",
		"variables": {
  			"search": "%s"
			}
		}`, userName))

	data, err := fetchTopLevelFromQuery(reqQuery)
	if err != nil {
		fmt.Println(err)
	}

	return strconv.Itoa(data.Data.User.ID)
}

func getTopMediaByID(AnilistID string, MediaType string, Page int, PerPage int) []mediaListItem {
	var MediaListItems []mediaListItem

	reqQuery := strings.NewReader(fmt.Sprintf(`{
	"query": "query Query($userId: Int, $sort: [MediaListSort], $page: Int, $perPage: Int, $type: MediaType) { Page(page: $page, perPage: $perPage) { mediaList(userId: $userId, sort: $sort, type: $type) {media {title {english native romaji} siteUrl coverImage{medium large}} score startedAt {day month year} completedAt {day month year}}}}",
	"variables": {
		"userId": %s,
		"sort": "SCORE_DESC",
		"page": %d,
		"perPage": %d,
		"type": "%s"
		}
	}`, AnilistID, Page, PerPage, MediaType))

	data, err := fetchTopLevelFromQuery(reqQuery)
	if err != nil {
		fmt.Println(err)
	}

	for _, Item := range data.Data.Page.MediaList {
		MediaListItems = append(MediaListItems, Item)
	}

	return MediaListItems
}

func fetchTopLevelFromQuery(QueryString *strings.Reader) (topLevel, error) {
	response, err := http.Post(aniListEndPoint, "application/json", QueryString)
	if err != nil {
		return topLevel{}, err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return topLevel{}, err
	}
	var data topLevel
	err = json.Unmarshal(body, &data)
	if err != nil {
		return topLevel{}, err
	}
	return data, nil
}

func getMediaIdByName(Name string) int {
	reqQuery := strings.NewReader(fmt.Sprintf(`{
	"query": "query Query($search: String) { Media(search: $search) { id }}",
	"variables": {
		"search": "%s"
		}
	}`, Name))
	data, err := fetchTopLevelFromQuery(reqQuery)
	if err != nil {
		fmt.Println(err)
	}
	return data.Data.Media.Id
}

func getRecommendationsByMediaId(Id int) []mediaRecItem {
	var MediaRecItems []mediaRecItem

	reqQuery := strings.NewReader(fmt.Sprintf(`{
	"query": "query Query($page: Int, $mediaId: Int, $perPage: Int) { Page(page: $page, perPage: $perPage) { recommendations(mediaId: $mediaId) { mediaRecommendation { coverImage { medium large } title { english native romaji } description averageScore genres siteUrl }}}}",
	"variables": {
		"page": 1,
		"perPage": 10,
		"mediaId": %d
		}
	}`, Id))
	fmt.Println(reqQuery)

	data, err := fetchTopLevelFromQuery(reqQuery)
	if err != nil {
		fmt.Println(err)
	}

	for _, Item := range data.Data.Page.Recs {
		MediaRecItems = append(MediaRecItems, Item)
	}

	return MediaRecItems
}

func getUserInfoByID(AnilistID string) user {
	reqQuery := strings.NewReader(fmt.Sprintf(`{
	"query": "query User($userId: Int) {User(id: $userId) {avatar {large medium }about name siteUrl statistics {anime {count episodesWatched meanScore}manga {chaptersRead count meanScore}}}}",
	"variables": {
 		"userId": %s
		}
	}`, AnilistID))

	data, err := fetchTopLevelFromQuery(reqQuery)
	if err != nil {
		fmt.Println(err)
	}
	return data.Data.User
}
