package telegram

import (
	tele "gopkg.in/telebot.v4"
)

func (b *Bot) CreateLessonHandler(
	lessonCreationManager *LessonCreationManager,
	adminTelegramID int64,
) {
	b.Handle("/create_lesson", func(c tele.Context) error {
		id := c.Sender().ID

		if id != adminTelegramID {
			return c.Send("You aren't authorized to use this command.")
		}
		lessonCreationManager.Start(adminTelegramID)
		return c.Send("Please enter the student's Telegram ID.")
	})
}
