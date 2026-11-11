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
			text = strings.TrimSpace(text)

			telegramID, err := strconv.ParseInt(text, 10, 64)
			if err != nil {
				return c.Send("Invalid Telegram ID. Please enter a valid numeric ID.")
			}

			if telegramID <= 0 {
				return c.Send("telegram ID must be more than 0")
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

func (b *Bot) CallbackHandler(
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

		id := c.Callback().Sender.ID

		if id != adminTelegramID {
			return c.Send("You aren't authorized to use this command.")
		}

		args := c.Args()

		if len(args) < 2 {
			return c.Send("cannot get args")
		}

		args[0] = strings.TrimPrefix(args[0], "\f")
		args[0] = strings.TrimSpace(args[0])
		button := args[0]
		callbackPayload := args[1]

		switch button {
		case "":
			return c.Send("invalid button")
		case "lesson_confirm":
			menu := &tele.ReplyMarkup{}
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

				if errReturnToConfirm != nil {
					return fmt.Errorf(
						"create lesson: %w",
						errors.Join(err, errReturnToConfirm),
					)
				}
				if errors.Is(err, domain.ErrDuplicateLesson) {
					return c.Send("You already have an active lesson scheduled for this student at this time.")
				}

				return fmt.Errorf("create lesson: %w", err)
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

			return c.Edit(message, menu)
		case "lesson_cancel":
			menu := &tele.ReplyMarkup{}
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

			return c.Edit("The lesson creation was cancelled.", menu)
		case "start_confirm":
			idLesson, err := strconv.ParseInt(callbackPayload, 10, 64)
			if err != nil {
				return c.Send("Invalid lesson ID. Please enter a valid numeric ID.")
			}

			if idLesson <= 0 {
				return c.Send("Lesson ID must be a positive number.")
			}

			if _, err := lessonService.StartLesson(ctx, idLesson); err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					return c.Send("Lesson not found")
				}

				if errors.Is(err, domain.ErrInvalidArgument) {
					return c.Send("Cannot start lesson in its current status")
				}

				return fmt.Errorf("lesson start: %w", err)
			}

			in := time.Now().In(cfg.TimeZone)
			year, month, day := in.Year(), in.Month(), in.Day()
			today := time.Date(year, month, day, 0, 0, 0, 0, cfg.TimeZone)
			tomorrow := today.AddDate(0, 0, 1)

			lessons, err := lessonService.GetLessonsByPeriod(
				ctx,
				today,
				tomorrow,
			)
			if err != nil {
				return fmt.Errorf("get lessons: %w", err)
			}

			msg, menu := buildLessonsMessage(lessons, cfg)

			return c.Edit(msg, menu)
		case "finish_confirm":
			idLesson, err := strconv.ParseInt(callbackPayload, 10, 64)
			if err != nil {
				return c.Send("Invalid lesson ID. Please enter a valid numeric ID.")
			}

			if idLesson <= 0 {
				return c.Send("Lesson ID must be a positive number.")
			}

			if _, err := lessonService.FinishLesson(ctx, idLesson); err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					return c.Send("Lesson not found")
				}

				if errors.Is(err, domain.ErrInvalidArgument) {
					return c.Send("Cannot finish lesson in its current status")
				}

				return fmt.Errorf("lesson finish: %w", err)
			}

			in := time.Now().In(cfg.TimeZone)
			year, month, day := in.Year(), in.Month(), in.Day()
			today := time.Date(year, month, day, 0, 0, 0, 0, cfg.TimeZone)
			tomorrow := today.AddDate(0, 0, 1)

			lessons, err := lessonService.GetLessonsByPeriod(
				ctx,
				today,
				tomorrow,
			)
			if err != nil {
				return fmt.Errorf("get lessons: %w", err)
			}

			msg, menu := buildLessonsMessage(lessons, cfg)

			return c.Edit(msg, menu)
		case "cancel_confirm":
			menu := &tele.ReplyMarkup{}

			cancelYesBtn := menu.Data("Yes, cancel", "cancel_yes", callbackPayload)
			cancelNoBtn := menu.Data("No, go back", "cancel_no", callbackPayload)

			menu.Inline(
				menu.Row(cancelYesBtn, cancelNoBtn),
			)

			return c.Edit("Are you sure you want to cancel the lesson?", menu)
		case "cancel_yes":
			menu := &tele.ReplyMarkup{}
			idLesson, err := strconv.ParseInt(callbackPayload, 10, 64)
			if err != nil {
				return c.Edit("Invalid lesson ID. Please enter a valid numeric ID.", menu)
			}

			if idLesson <= 0 {
				return c.Edit("Lesson ID must be a positive number.", menu)
			}

			if _, err := lessonService.CancelLesson(ctx, idLesson); err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					return c.Edit("Lesson not found", menu)
				}

				if errors.Is(err, domain.ErrInvalidArgument) {
					return c.Edit("Cannot cancel lesson in its current status", menu)
				}

				return fmt.Errorf("lesson cancel: %w", err)
			}

			in := time.Now().In(cfg.TimeZone)
			year, month, day := in.Year(), in.Month(), in.Day()
			today := time.Date(year, month, day, 0, 0, 0, 0, cfg.TimeZone)
			tomorrow := today.AddDate(0, 0, 1)

			lessons, err := lessonService.GetLessonsByPeriod(
				ctx,
				today,
				tomorrow,
			)
			if err != nil {
				return fmt.Errorf("get lessons: %w", err)
			}

			if len(lessons) == 0 {
				return c.Edit("Enjoy your day off! :) You don't have any lessons today", menu)
			}

			msg, menu := buildLessonsMessage(lessons, cfg)

			return c.Edit(msg, menu)
		case "cancel_no":
			menu := &tele.ReplyMarkup{}

			in := time.Now().In(cfg.TimeZone)
			year, month, day := in.Year(), in.Month(), in.Day()
			today := time.Date(year, month, day, 0, 0, 0, 0, cfg.TimeZone)
			tomorrow := today.AddDate(0, 0, 1)

			lessons, err := lessonService.GetLessonsByPeriod(
				ctx,
				today,
				tomorrow,
			)
			if err != nil {
				return fmt.Errorf("get lessons: %w", err)
			}

			msg, menu := buildLessonsMessage(lessons, cfg)

			return c.Edit(msg, menu)
		default:
			return c.Send("Unknown operation")
		}
	})
}

func (b *Bot) TodayLessonsHandler(
	ctx context.Context,
	lessonService *service.LessonService,
	cfg config.Config,
	adminTelegramID int64,
) {
	b.Handle("/today", func(c tele.Context) error {
		id := c.Sender().ID

		if id != adminTelegramID {
			return c.Send("You aren't authorized to use this command.")
		}

		in := time.Now().In(cfg.TimeZone)
		year, month, day := in.Year(), in.Month(), in.Day()
		today := time.Date(year, month, day, 0, 0, 0, 0, cfg.TimeZone)
		tomorrow := today.AddDate(0, 0, 1)

		lessons, err := lessonService.GetLessonsByPeriod(
			ctx,
			today,
			tomorrow,
		)
		if err != nil {
			return fmt.Errorf("get lessons: %w", err)
		}

		if len(lessons) == 0 {
			return c.Send("Enjoy your day off! :) You don't have any lessons today")
		}

		msg, menu := buildLessonsMessage(lessons, cfg)
		return c.Send(msg, menu)
	})
}
func buildLessonsMessage(
	lessons []domain.Lesson,
	cfg config.Config,
) (string, *tele.ReplyMarkup) {
	menu := &tele.ReplyMarkup{}
	result := make([]string, 0, len(lessons))
	lessonRows := make([]tele.Row, 0, len(lessons))
	for _, lesson := range lessons {
		start := lesson.StartedAt
		finish := lesson.FinishedAt

		var duration time.Duration
		var durationText string = "Not started"

		if start != nil && finish != nil {
			duration = finish.Sub(*start)
			durationText = duration.String()
		} else if start != nil && finish == nil {
			now := time.Now()
			duration = now.Sub(*start)
			durationText = duration.String()
		}

		lesson.DateTime = lesson.DateTime.In(cfg.TimeZone)

		switch lesson.LessonStatus {
		case domain.LessonStatusScheduled:
			id := lesson.ID
			lessonID := strconv.FormatInt(id, 10)

			startBtn := menu.Data("Start", "start_confirm", lessonID)
			cancelBtn := menu.Data("Cancel", "cancel_confirm", lessonID)

			row := menu.Row(
				startBtn, cancelBtn,
			)

			lessonRows = append(lessonRows, row)
		case domain.LessonStatusInProgress:
			id := lesson.ID
			lessonID := strconv.FormatInt(id, 10)

			finishBtn := menu.Data("Finish", "finish_confirm", lessonID)
			cancelBtn := menu.Data("Cancel", "cancel_confirm", lessonID)

			row := menu.Row(
				finishBtn, cancelBtn,
			)

			lessonRows = append(lessonRows, row)
		}
		msg := fmt.Sprintf(

			"Student ID: %d\n"+
				"Date: %s\n"+
				"Duration: %s\n"+
				"Platform: %s\n"+
				"URL: %s\n"+
				"Status: %s\n",
			lesson.StudentID,
			lesson.DateTime.Format("02-01-2006 15:04"),
			durationText,
			lesson.MeetingApp,
			lesson.MeetingURL,
			lesson.LessonStatus,
		)
		result = append(result, msg)
	}

	response := []string{"There are your lessons today:\n"}
	response = append(response, result...)
	message := strings.Join(response, "\n")

	menu.Inline(
		lessonRows...,
	)

	return message, menu
}
