package data

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type User struct {
	DiscordID string
	AniListID string
}

func Init() {
	db, err := sql.Open("sqlite3", "./users.db")
	if err != nil {
		fmt.Println(err)
	}

	err = db.Ping()
	if err != nil {
		fmt.Println("Something went wrong..")
		fmt.Println(err)
	}
	fmt.Println("CONNECTED")
}

func CreateUser(discordID string, aniListID string) (int, error) {
	_, err := GetUserByDiscordID(discordID)
	if err == nil {
		return -1, fmt.Errorf("account already linked to this discord user")
	}
	result, err := db.Exec("insert into users (discordID, anilistID) values (?, ?)", discordID, aniListID)
	if err != nil {
		return -1, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}
	return int(id), nil
}

func GetUserByDiscordID(discordID string) (User, error) {
	var user User
	err := db.QueryRow(`SELECT * FROM users WHERE discordID ="?"`, discordID).Scan(&user.DiscordID, &user.AniListID)
	if err != sql.ErrNoRows {
		return User{}, err
	}
	fmt.Printf("Found user %s", user)
	return user, nil
}

func UpdateUser(discordID string, anilistID string) error {
	_, err := db.Exec(`UPDATE users SET anilistID = "?" WHERE discordID = "?"`, anilistID, discordID)
	if err != nil {
		return err
	}
	return nil
}
