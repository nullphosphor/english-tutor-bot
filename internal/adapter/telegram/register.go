package telegram

import (
	"context"

	"github.com/nullphosphor/english-tutor-bot/internal/config"
	"github.com/nullphosphor/english-tutor-bot/internal/service"
)

func (b *Bot) RegisterHandlers(
	ctx context.Context,
	lessonService *service.LessonService,
	studentService *service.StudentService,
	lessonManager *LessonCreationManager,
	cfg config.Config,
	adminTelegramID int64,
) {
	b.CreateLessonHandler(
		lessonManager,
		adminTelegramID,
	)

	b.LessonCreationMessageHandler(
		ctx,
		lessonManager,
		studentService,
		cfg,
		adminTelegramID,
	)
}
