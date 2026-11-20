package telegram

import tele "gopkg.in/telebot.v4"

func (b *Bot) CreateHomeworkHandler(
	homeworkCreationManager *HomeworkCreationManager,
	adminTelegramID int64,
) {
	b.Handle("/create_homework", func(c tele.Context) error {
		id := c.Sender().ID

		if id != adminTelegramID {
			return c.Send("You aren't authorized to use this command.")
		}
		homeworkCreationManager.Start(id)
		return c.Send("Please enter the student's Telegram ID.")
	})
}
