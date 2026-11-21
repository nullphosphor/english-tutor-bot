package telegram

import (
	"errors"
	"fmt"

	tele "gopkg.in/telebot.v4"
)

func (b *Bot) CreateLessonHandler(
	lessonCreationManager *LessonCreationManager,
	homeworkCreationManager *HomeworkCreationManager,
	adminTelegramID int64,
) {
	b.Handle("/create_lesson", func(c tele.Context) error {
		id := c.Sender().ID

		if id != adminTelegramID {
			return c.Send("You aren't authorized to use this command.")
		}
		_, err := homeworkCreationManager.GetState(id)
		if err != nil {
			if !errors.Is(err, ErrDialogNotFound) {
				return fmt.Errorf("get homework state: %w", err)
			}
		}
		if err == nil {
			return c.Send("You must first complete or cancel the creation of your homework.")
		}

		lessonCreationManager.Start(adminTelegramID)
		return c.Send("Please enter the student's Telegram ID.")
	})
}
