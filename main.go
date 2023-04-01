package main

import (
	"flag"
	"log"

	"github.com/ashwinath/money-tracker-telegram/config"
	database "github.com/ashwinath/money-tracker-telegram/db"
	"github.com/ashwinath/money-tracker-telegram/telegram"
)

func main() {
	configPath := flag.String("config", "", "Path to config file")
	flag.Parse()

	c, err := config.New(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.New(
		c.DBConfig.Host,
		c.DBConfig.User,
		c.DBConfig.Password,
		c.DBConfig.DBName,
		c.DBConfig.Port,
	)
	defer db.Close()

	if err != nil {
		log.Fatal(err)
	}

	t, err := telegram.New(c.APIKey, c.Debug)
	if err != nil {
		log.Fatal(err)
	}

	t.Run(c.AllowedUser)
}
