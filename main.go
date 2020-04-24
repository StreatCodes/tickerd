package main

import (
	"fmt"
	"log"

	"github.com/asdine/storm/v3"
)

//DB global access to the DB
var DB *storm.DB

func adminSetup() error {
	count, err := DB.Count(&User{})
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
	DB, err = storm.Open("ticker.db")
	if err != nil {
		log.Fatalf("Failed to connect to ticker.db: %s\n", err)
	}
	defer DB.Close()

	err = adminSetup()
	if err != nil {
		log.Fatalf("Error setting up DB: %s\n", err)
	}

	registerHandler("echo", echoHandler)

	initWeb()
}

func echoHandler(reqJSON []byte) ([]byte, error) {
	return reqJSON, nil
}
