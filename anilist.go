package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type anilistID struct {
	ID string `json:"id"`
}

type topLevel struct {
	Data aniData `json:"data"`
}

type aniData struct {
	Page page `json:"Page"`
	User user `json:"User"`
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
	CoverImage coverImage `json:"coverImage"`
	Title      title      `json:"title"`
	SiteURL    string     `json:"siteUrl"`
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

const aniListEndPoint string = "https://graphql.anilist.co"

func getRecommendationsByTitle(title string) topLevel {
	reqBody := strings.NewReader(`{
    "query": "query Query($page: Int, $mediaId: Int) {Page(page: $page) {recommendations(mediaId: $mediaId) {mediaRecommendation {coverImage{large}title {english native romaji}}}}}",
    "variables": {
        "page": 1,
        "mediaId": 105778
        }
	}`)

	response, err := http.Post(aniListEndPoint, "application/json", reqBody)
	if err != nil {
		return topLevel{}
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return topLevel{}
	}

	var data = topLevel{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return topLevel{}
	}
	fmt.Println(data.Data.Page.Recs[0].Media.Title.English)
	fmt.Println(data)

	return data
}

func searchUserIDByName(userName string) string {
	reqQuery := strings.NewReader(fmt.Sprintf(`{
		"query": "query Query($search: String) {User(search: $search) {id name}}",
		"variables": {
  			"search": "%s"
			}
		}`, userName))
	// response, err := http.Post(aniListEndPoint, "application/json", reqQuery)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// defer response.Body.Close()

	// body, err := io.ReadAll(response.Body)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// var data topLevel
	// err = json.Unmarshal(body, &data)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	data, err := fetchTopLevelFromQuery(reqQuery)
	if err != nil {
		fmt.Println(err)
	}

	return strconv.Itoa(data.Data.User.ID)
}

func getTopMediaByID(AnilistID string, MediaType string, Page int, PerPage int) []mediaListItem {
	var MediaListItems []mediaListItem

	reqQuery := strings.NewReader(fmt.Sprintf(`{
	"query": "query Query($userId: Int, $sort: [MediaListSort], $page: Int, $perPage: Int, $type: MediaType) { Page(page: $page, perPage: $perPage) { mediaList(userId: $userId, sort: $sort, type: $type) {media {title {english native romaji} siteUrl coverImage{medium}} score startedAt {day month year} completedAt {day month year}}}}",
	"variables": {
		"userId": %s,
		"sort": "SCORE_DESC",
		"page": %d,
		"perPage": %d,
		"type": "%s"
		}
	}`, AnilistID, Page, PerPage, MediaType))

	// response, err := http.Post(aniListEndPoint, "application/json", reqQuery)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// defer response.Body.Close()

	// body, err := io.ReadAll(response.Body)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// var data topLevel
	// err = json.Unmarshal(body, &data)
	// if err != nil {
	// 	fmt.Println(err)
	// }

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
