package main

import (
	"log"

	"github.com/nullphosphor/english-tutor-bot/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("start telegram bot: %v", err)
	}
}
