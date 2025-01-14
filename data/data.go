package data

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var Db *sql.DB

type User struct {
	DiscordID string
	AniListID string
}

func Init() {
	var err error
	Db, err = sql.Open("sqlite3", "./users.db")
	if err != nil {
		fmt.Println(err)
	}

	err = Db.Ping()
	if err != nil {
		fmt.Println("Something went wrong..")
		fmt.Println(err)
	}
	fmt.Println("CONNECTED")
}

func CreateUser(discordID string, aniListID string) (int, error) {
	user, err := GetUserByDiscordID(discordID)
	if user != "" {
		UpdateUser(discordID, aniListID)
		return 0, nil
	}
	result, err := Db.Exec(`insert into users (discordID, anilistID) values (?, ?)`, discordID, aniListID)
	if err != nil {
		fmt.Println(err)
		return -1, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		fmt.Println(err)

		return -1, err
	}
	return int(id), nil
}

func GetUserByDiscordID(discordID string) (string, error) {
	var anilistID string
	err := Db.QueryRow(`SELECT anilistID FROM users WHERE discordID = ?`, discordID).Scan(&anilistID)
	if err == sql.ErrNoRows {
		return "", err
	}
	fmt.Printf("Found user %s", anilistID)
	return anilistID, nil
}

func UpdateUser(discordID string, anilistID string) error {
	_, err := Db.Exec(`UPDATE users SET anilistID = ? WHERE discordID = ?`, anilistID, discordID)
	if err != nil {
		return err
	}
	return nil
}
