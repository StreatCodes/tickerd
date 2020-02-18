package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

//DB global access to the mysql DB
var DB *sqlx.DB

type Tickets struct {
	Name string
}

func (h Tickets) Handle(userID int) ([]byte, error) {
	return json.Marshal(fmt.Sprintf("Tickets %s with ID of %d", h.Name, userID))
}

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

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Error reading config - %s\n", err)
	}

	DB, err = sqlx.Connect("mysql", config.DBURL)
	if err != nil {
		log.Fatalf("Failed to connect to the mysql DB - %s\n", err)
	}
	initWeb()
}
