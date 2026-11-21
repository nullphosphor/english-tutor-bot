package telegram

import (
	"errors"
	"fmt"

	tele "gopkg.in/telebot.v4"
)

func (b *Bot) CreateHomeworkHandler(
	homeworkCreationManager *HomeworkCreationManager,
	lessonCreationManager *LessonCreationManager,
	adminTelegramID int64,
) {
	b.Handle("/create_homework", func(c tele.Context) error {
		id := c.Sender().ID

		if id != adminTelegramID {
			return c.Send("You aren't authorized to use this command.")
		}

		_, err := lessonCreationManager.GetState(id)
		if err != nil {
			if !errors.Is(err, ErrDialogNotFound) {
				return fmt.Errorf("get lesson state: %w", err)
			}
		}
		if err == nil {
			return c.Send("You must first complete or cancel the creation of your lesson.")
		}
		homeworkCreationManager.Start(id)
		return c.Send("Please enter the student's Telegram ID.")
	})
}
