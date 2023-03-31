package main

import (
	"flag"
	"log"

	"github.com/ashwinath/money-tracker-telegram/config"
	"github.com/ashwinath/money-tracker-telegram/telegram"
)

func main() {
	configPath := flag.String("config", "", "Path to config file")
	flag.Parse()

	c, err := config.New(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	t, err := telegram.New(c.APIKey, c.Debug)
	if err != nil {
		log.Fatal(err)
	}

	t.Run()
}
