package telegram

import (
	"context"

	"github.com/nullphosphor/english-tutor-bot/internal/config"
	"github.com/nullphosphor/english-tutor-bot/internal/service"
)

func (b *Bot) RegisterHandlers(
	ctx context.Context,
	lessonService *service.LessonService,
	lessonManager *LessonCreationManager,
	studentService *service.StudentService,
	studentInviteService *service.StudentInviteService,
	inviteManager *StudentInviteCreationManager,
	homeworkService *service.HomeworkService,
	homeworkCreationManager *HomeworkCreationManager,
	cfg config.Config,
	adminTelegramID int64,
) {
	b.StartHandler(
		ctx,
		studentService,
		studentInviteService,
		adminTelegramID,
	)

	b.CreateLessonHandler(
		lessonManager,
		homeworkCreationManager,
		inviteManager,
		adminTelegramID,
	)

	b.CreationMessageHandler(
		ctx,
		lessonManager,
		homeworkCreationManager,
		inviteManager,
		studentService,
		studentInviteService,
		cfg,
		adminTelegramID,
	)

	b.CallbackHandler(
		ctx,
		lessonManager,
		lessonService,
		studentService,
		studentInviteService,
		inviteManager,
		homeworkCreationManager,
		homeworkService,
		cfg,
		adminTelegramID,
	)

	b.TodayLessonsHandler(
		ctx,
		lessonService,
		cfg,
		adminTelegramID,
	)

	b.CreateInviteHandler(
		ctx,
		studentInviteService,
		adminTelegramID,
	)

	b.CreateHomeworkHandler(
		homeworkCreationManager,
		lessonManager,
		inviteManager,
		adminTelegramID,
	)
}
