package main

import (
	"github.com/GTedya/gophermart/internal/app"
	"log"

	_ "github.com/lib/pq"

	"github.com/GTedya/gophermart/config"
)

func main() {
	conf, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	app.Run(conf)
}
