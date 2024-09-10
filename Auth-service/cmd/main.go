package main

import (
	"log"
	"log/slog"
	"os"

	"auth-service/cmd/api"
	"auth-service/internal/items/config"
)

func main() {
	config, err := config.New()
	if err != nil {
		log.Fatalln(err)
	}

	logFile, err := os.OpenFile("application.log", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	logger := slog.New(slog.NewJSONHandler(logFile, nil))

	log.Fatalln(api.Run(config, logger))

}
