package main

import (
	"fmt"
	"log"

	"github.com/asdine/storm/v3"
)

//global access to databases
var tickerDB *storm.DB
var attachmentDB *storm.DB

func adminSetup() error {
	count, err := tickerDB.Count(&User{})
	if err != nil {
		return err
	}

	if count < 1 {
		fmt.Println("Creating user") //TODO improve
		_, err := createUser("Admin", "admin@ticker.io", []byte("password"), true)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	var err error
	tickerDB, err = storm.Open("ticker.db")
	if err != nil {
		log.Fatalf("Failed to connect to ticker.db: %s\n", err)
	}
	defer tickerDB.Close()

	err = adminSetup()
	if err != nil {
		log.Fatalf("Error setting up DB: %s\n", err)
	}

	attachmentDB, err = storm.Open("attachments.db")
	if err != nil {
		log.Fatalf("Failed to connect to attachments.db: %s\n", err)
	}
	defer attachmentDB.Close()

	registerHandler("echo", echoHandler)

	initWeb()
}

func echoHandler(userID int64, req []byte) ([]byte, error) {
	return req, nil
}
