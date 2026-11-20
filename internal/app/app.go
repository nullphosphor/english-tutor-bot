package app

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"sync"
	"syscall"

	"github.com/nullphosphor/english-tutor-bot/internal/adapter/postgres"
	"github.com/nullphosphor/english-tutor-bot/internal/adapter/telegram"
	"github.com/nullphosphor/english-tutor-bot/internal/config"
	"github.com/nullphosphor/english-tutor-bot/internal/service"
)

func Run() error {
	shutdownCtx, shutdownCancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM, syscall.SIGINT,
	)
	defer shutdownCancel()

	var wg sync.WaitGroup
	cfg := config.LoadConfigMust()

	log.Println("initializing pool")
	pool, err := postgres.NewPool(
		shutdownCtx,
		cfg.DatabaseURL,
	)
	if err != nil {
		return fmt.Errorf("failed to init pool: %w", err)
	}
	defer pool.Close()

	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	log.Println("initializing feature lesson")
	lessonRepository := postgres.NewLessonRepository(pool)
	lessonService := service.NewLessonService(lessonRepository)
	lessonManager := telegram.NewLessonCreationManager()

	log.Println("initializing feature student")
	studentRepository := postgres.NewStudentRepository(pool)
	studentService := service.NewStudentService(studentRepository)

	log.Println("initializing feature invites")
	invitesRepository := postgres.NewStudentInviteRepository(pool)
	inviteStudentService := service.NewStudentInviteService(invitesRepository)

	log.Println("initializing feature homework")
	homeworkRepository := postgres.NewHomeworkRepository(pool)
	homeworkService := service.NewHomeworkService(homeworkRepository)
	homeworkManager := telegram.NewHomeworkCreationManager()

	log.Println("initializing telegram bot")
	bot, err := telegram.NewBot(cfg.Token)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}
	bot.RegisterHandlers(
		appCtx,
		lessonService,
		lessonManager,
		studentService,
		inviteStudentService,
		homeworkService,
		homeworkManager,
		cfg,
		cfg.AdminTelegramID,
	)

	wg.Go(func() {
		bot.Start()
	})

	<-shutdownCtx.Done()
	appCancel()
	bot.Stop()
	wg.Wait()

	return nil
}
