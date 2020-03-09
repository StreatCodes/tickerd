package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

//DB global access to the mysql DB
var DB *sqlx.DB

//Config for the ticker application
type Config struct {
	DBURL string
}

func loadConfig() (Config, error) {
	var config Config

	f, err := os.Open("config.json")
	defer f.Close()
	if err != nil {
		return config, err
	}

	dec := json.NewDecoder(f)
	err = dec.Decode(&config)

	return config, err
}

func dbSetup() error {
	var count []int
	err := DB.Select(&count, `SELECT COUNT(*) FROM User`)
	if err != nil {
		return err
	}

	if count[0] < 1 {
		encryptedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		//TODO don't print admin password to logs...
		fmt.Println("No users found, creating default Admin with default password")
		DB.MustExec(
			`INSERT INTO User (Name, Email, Admin, Password) VALUES (?, ?, ?, ?)`,
			"Admin", "admin@ticker.io", true, encryptedPassword,
		)

	}
	return nil
}

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Error reading config - %s\n", err)
	}

	DB, err = sqlx.Connect("mysql", config.DBURL)
	if err != nil {
		log.Fatalf("Failed to connect to the mysql DB - %s\n", err)
	}

	dbSetup()

	registerHandler("me", wshMe)
	registerHandler("echo", echo)

	initWeb()
}
