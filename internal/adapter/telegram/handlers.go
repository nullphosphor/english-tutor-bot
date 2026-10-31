package telegram

import tele "gopkg.in/telebot.v4"

func (b *Bot) EchoHandler() {
	b.Handle("/hello", func(c tele.Context) error {
		return c.Send("hello")
	})

}
