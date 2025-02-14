package main

import (
	"github.com/joho/godotenv"
	"github.com/smthjapanese/avito-merch/config"
	"github.com/smthjapanese/avito-merch/internal/app"
	"log"
)

func main() {
	// Configuration
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
