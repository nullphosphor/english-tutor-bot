package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/nullphosphor/english-tutor-bot/internal/adapter/postgres"
	"github.com/nullphosphor/english-tutor-bot/internal/adapter/telegram"
	"github.com/nullphosphor/english-tutor-bot/internal/config"
)

func main() {
	log.Println("start telegram bot")
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM, syscall.SIGINT,
	)
	defer cancel()

	cfg := config.LoadConfigMust()

	log.Println("initializing pool")
	pool, err := postgres.NewPool(
		ctx,
		cfg.DatabaseURL,
	)
	if err != nil {
		log.Printf("failed to init postgres pool: %v", err)
		return
	}
	defer pool.Close()

	log.Println("initializing telegram bot")
	bot, err := telegram.NewBot(cfg.Token)
	if err != nil {
		log.Printf("failed to init telegram bot: %v", err)
		return
	}

	go func() {
		bot.Start()
	}()

	bot.RegisterHandlers()
	<-ctx.Done()
	bot.Stop()
}
