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
	Recs []mediaRec `json:"recommendations"`
}

type user struct {
	ID int `json:"id"`
}

type mediaRec struct {
	Media media `json:"mediaRecommendation"`
}

type media struct {
	CoverImage coverImage `json:"coverImage"`
	Title      title      `json:"title"`
}

type coverImage struct {
	Large string `json:"large"`
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
	response, err := http.Post(aniListEndPoint, "application/json", reqQuery)
	if err != nil {
		fmt.Println(err)
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}
	var data topLevel
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println(err)
	}

	return strconv.Itoa(data.Data.User.ID)
}
