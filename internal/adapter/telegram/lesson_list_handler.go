package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/nullphosphor/english-tutor-bot/internal/config"
	"github.com/nullphosphor/english-tutor-bot/internal/domain"
	"github.com/nullphosphor/english-tutor-bot/internal/service"
	tele "gopkg.in/telebot.v4"
)

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

		switch lesson.Status {
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
			lesson.Platform,
			lesson.URL,
			lesson.Status,
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
