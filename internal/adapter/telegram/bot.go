package telegram

import (
	"fmt"
	"time"

	tele "gopkg.in/telebot.v4"
)

type Bot struct {
	*tele.Bot
}

func NewBot(token string) (*Bot, error) {
	b, err := tele.NewBot(tele.Settings{
		Token:       token,
		Poller:      &tele.LongPoller{Timeout: 10 * time.Second},
		Synchronous: true,
	})
	if err != nil {
		return nil, fmt.Errorf("create bot: %w", err)
	}

	return &Bot{Bot: b}, nil
}
