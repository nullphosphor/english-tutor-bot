package telegram

import (
	"context"
	"errors"
	"fmt"

	"github.com/nullphosphor/english-tutor-bot/internal/domain"
	"github.com/nullphosphor/english-tutor-bot/internal/service"
	tele "gopkg.in/telebot.v4"
)

func (b *Bot) CreateInviteHandler(
	ctx context.Context,
	studentInviteService *service.StudentInviteService,
	adminTelegramID int64,
) {
	b.Handle("/create_invite", func(c tele.Context) error {
		id := c.Sender().ID

		if id != adminTelegramID {
			return c.Send("You aren't authorized to use this command.")
		}

		payload := c.Message().Payload
		if payload == "" {
			return c.Send("Empty student name.")
		}

		invName := domain.StudentInvite{StudentName: payload}

		_, token, err := studentInviteService.CreateInvite(
			ctx,
			invName,
		)
		if err != nil {
			if errors.Is(err, domain.ErrInvalidArgument) {
				return c.Send("Invalid student name.")
			}

			if errors.Is(err, domain.ErrAlreadyExists) {
				return c.Send("Can't create a unique link. Please try again.")
			}

			return fmt.Errorf("create invite: %w", err)
		}

		botUsername := b.Me.Username
		response := fmt.Sprint("https://t.me/" + botUsername + "?start=" + token)

		return c.Send(response)
	})
}
