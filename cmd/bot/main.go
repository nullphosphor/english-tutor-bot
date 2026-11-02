package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/nullphosphor/english-tutor-bot/internal/adapter/postgres"
	"github.com/nullphosphor/english-tutor-bot/internal/adapter/telegram"
	"github.com/nullphosphor/english-tutor-bot/internal/config"
	"github.com/nullphosphor/english-tutor-bot/internal/service"
)

func main() {
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

	log.Println("initializing lesson feature")
	lessonRepository := postgres.NewLessonRepository(pool)
	lessonService := service.NewLessonService(lessonRepository)
	lessonManager := telegram.NewLessonCreationManager()

	log.Println("initializing student feature")
	studentRepository := postgres.NewStudentRepository(pool)
	studentService := service.NewStudentService(studentRepository)

	log.Println("initializing telegram bot")
	bot, err := telegram.NewBot(cfg.Token)
	if err != nil {
		log.Printf("failed to init telegram bot: %v", err)
		return
	}
	bot.RegisterHandlers(
		ctx,
		lessonService,
		studentService,
		lessonManager,
		cfg,
		cfg.AdminTelegramID,
	)

	go func() {
		bot.Start()
	}()

	<-ctx.Done()
	bot.Stop()
}
