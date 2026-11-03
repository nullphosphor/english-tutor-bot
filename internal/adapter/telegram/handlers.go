package telegram

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/nullphosphor/english-tutor-bot/internal/config"
	"github.com/nullphosphor/english-tutor-bot/internal/domain"
	"github.com/nullphosphor/english-tutor-bot/internal/service"
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

func (b *Bot) LessonCreationMessageHandler(
	ctx context.Context,
	lessonCreationManager *LessonCreationManager,
	studentService *service.StudentService,
	cfg config.Config,
	adminTelegramID int64,
) {
	b.Handle(tele.OnText, func(c tele.Context) error {
		id := c.Sender().ID

		if id != adminTelegramID {
			return c.Send("You aren't authorized to use this command.")
		}

		state, err := lessonCreationManager.GetState(id)
		if err != nil {
			if errors.Is(err, ErrDialogNotFound) {
				return c.Send("Please start by running /create_lesson.")
			}
			return fmt.Errorf("get state: %w", err)
		}
		switch state.Step {
		case StepWaitingStudentID:
			text := c.Text()

			telegramID, err := strconv.ParseInt(text, 10, 64)
			if err != nil {
				return c.Send("Invalid Telegram ID. Please enter a valid numeric ID.")
			}

			student, err := studentService.GetStudentByTelegramID(
				ctx,
				telegramID,
			)

			if err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					return c.Send("Student not found. Please check the Telegram ID and try again.")
				}
				return fmt.Errorf("get student: %w", err)
			}

			if !student.IsActive {
				return c.Send("This student is inactive.")
			}

			if err := lessonCreationManager.UpdateState(
				adminTelegramID,
				student.ID,
				StepWaitingDate,
			); err != nil {
				if errors.Is(err, ErrDialogNotFound) {
					return c.Send("Please start by running /create_lesson.")
				}
				return fmt.Errorf("update state: %w", err)
			}
			return c.Send("Please enter the lesson date and time (DD-MM-YYYY HH:MM).")

		case StepWaitingDate:
			text := c.Text()

			date, err := time.ParseInLocation(
				"02-01-2006 15:04",
				text,
				cfg.TimeZone,
			)
			if err != nil {
				return c.Send("Please enter a valid date and time (DD-MM-YYYY HH:MM)")
			}
			tz := time.Now().In(cfg.TimeZone)

			if !date.After(tz) {
				return c.Send("The lesson date must be in the future")
			}

			if err := lessonCreationManager.UpdateDate(
				id,
				date,
				StepWaitingPlatform,
			); err != nil {
				if errors.Is(err, ErrDialogNotFound) {
					return c.Send("Please start by running /create_lesson.")
				}
				return fmt.Errorf("update date: %w", err)
			}

			return c.Send("Please enter the meeting platform.")

		case StepWaitingPlatform:
			text := c.Text()

			if text != "Zoom" && text != "Google Meet" {
				return c.Send("Invalid platform. Please choose Zoom or Google Meet.")
			}
			if err := lessonCreationManager.UpdatePlatform(
				id,
				text,
				StepWaitingURL,
			); err != nil {
				if errors.Is(err, ErrDialogNotFound) {
					return c.Send("Please start by running /create_lesson.")
				}
				return fmt.Errorf("update platform: %w", err)
			}
			return c.Send("Please enter the meeting URL.")

		case StepWaitingURL:
			text := c.Text()

			if strings.TrimSpace(text) == "" {
				return c.Send("Please enter a valid url")
			}

			urlParse, err := url.ParseRequestURI(text)
			if err != nil {
				return c.Send("Please enter a valid url")
			}

			if !strings.EqualFold(urlParse.Scheme, "https") {
				return c.Send("Invalid scheme")
			}

			if urlParse.Host == "" {
				return c.Send("Invalid host")
			}

			if err := lessonCreationManager.UpdateURL(
				id,
				text,
				StepWaitingConfirm,
			); err != nil {
				if errors.Is(err, ErrDialogNotFound) {
					return c.Send("Please start by running /create_lesson.")
				}
				return fmt.Errorf("update url: %w", err)
			}

			menu := &tele.ReplyMarkup{}

			confirmBtn := menu.Data("Confirm", "lesson_confirm", "confirm_lesson")
			cancelBtn := menu.Data("Cancel", "lesson_cancel", "cancel_lesson")

			menu.Inline(
				menu.Row(confirmBtn, cancelBtn),
			)

			stateFinish, err := lessonCreationManager.GetState(id)
			if err != nil {
				return fmt.Errorf("get state: %w", err)
			}

			message := fmt.Sprintf(
				"Please confirm the lesson:\n\n"+
					"Student database ID: %d\n"+
					"Date: %s\n"+
					"Platform: %s\n"+
					"Meeting URL: %s",
				stateFinish.StudentID,
				stateFinish.Date.Format("02-01-2006 15:04"),
				stateFinish.Platform,
				stateFinish.URL,
			)

			return c.Send(message, menu)
		}
		return nil
	})
}

func (b *Bot) ConfirmCreationHandler(
	ctx context.Context,
	lessonCreationManager *LessonCreationManager,
	lessonService *service.LessonService,
	cfg config.Config,
	adminTelegramID int64,
) {
	b.Handle(tele.OnCallback, func(c tele.Context) error {
		if err := c.Respond(); err != nil {
			return fmt.Errorf("respond to callback: %w", err)
		}

		log.Println(c.Callback().Unique)
		log.Println(c.Callback().Data)
		id := c.Callback().Sender.ID

		if id != adminTelegramID {
			return c.Send("You aren't authorized to use this command.")
		}

		button := c.Args()[1]
		log.Println(c.Args()[1])
		switch button {
		case "confirm_lesson":
			state, err := lessonCreationManager.BeginCreation(adminTelegramID)
			if err != nil {
				if errors.Is(err, ErrDialogNotFound) {
					return c.Send("Please start by running /create_lesson.")
				}
				if errors.Is(err, ErrInvalidScriptStep) {
					return c.Send("This lesson is no longer awaiting confirmation.")
				}
				return fmt.Errorf("begin lesson creation: %w", err)
			}

			lessonDomain := domain.NewLesson(
				state.StudentID,
				state.Date,
				state.Platform,
				state.URL,
			)

			lesson, err := lessonService.CreateLesson(ctx, lessonDomain)
			if err != nil {
				errReturnToConfirm := lessonCreationManager.ReturnToConfirmation(adminTelegramID)

				return fmt.Errorf(
					"create lesson: %w",
					errors.Join(err, errReturnToConfirm),
				)
			}

			lessonCreationManager.Finish(adminTelegramID)

			lesson.DateTime = lesson.DateTime.In(cfg.TimeZone)

			message := fmt.Sprintf(
				"The lesson was successfully created!\n\n"+
					"Student database ID: %d\n"+
					"Date: %s\n"+
					"Platform: %s\n"+
					"Meeting URL: %s",
				lesson.StudentID,
				lesson.DateTime.Format("02-01-2006 15:04"),
				lesson.MeetingApp,
				lesson.MeetingURL,
			)

			return c.Send(message)

		case "cancel_lesson":
			if err := lessonCreationManager.CancelDialog(adminTelegramID); err != nil {
				if errors.Is(err, ErrDialogNotFound) {
					return c.Send("Please start by running /create_lesson.")
				}

				if errors.Is(err, ErrInvalidScriptStep) {
					return c.Send("invalid step")
				}
				log.Printf("cancel dialog: %v", err)
				return c.Send("invalid error")
			}

			return c.Send("The lesson creation was cancelled.")
		default:
			return nil
		}
	})
}
