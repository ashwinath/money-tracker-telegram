package main

import (
	"flag"
	"log"

	"github.com/ashwinath/money-tracker-telegram/config"
	database "github.com/ashwinath/money-tracker-telegram/db"
	"github.com/ashwinath/money-tracker-telegram/processor"
	"github.com/ashwinath/money-tracker-telegram/telegram"
	"github.com/ashwinath/money-tracker-telegram/webhandler"
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

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	processorManager, err := processor.NewManager(db)
	if err != nil {
		log.Fatal(err)
	}

	t, err := telegram.New(
		c.APIKey,
		c.Debug,
		c.AllowedUser,
		processorManager,
	)
	if err != nil {
		log.Fatal(err)
	}

	// start telegram bot
	go t.Run() // fire and forget

	// start http server
	handler := webhandler.NewDataDumpHandler(db)
	handler.Serve(8080)
}
