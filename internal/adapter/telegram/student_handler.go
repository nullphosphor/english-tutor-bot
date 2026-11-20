package telegram

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nullphosphor/english-tutor-bot/internal/config"
	"github.com/nullphosphor/english-tutor-bot/internal/domain"
	tele "gopkg.in/telebot.v4"
)

func buildStudentsMessage(
	students []domain.Student,
) (string, *tele.ReplyMarkup) {
	menu := &tele.ReplyMarkup{}
	result := make([]string, 0, len(students))
	var studentRows []tele.Row
	for _, student := range students {
		id := strconv.FormatInt(student.ID, 10)
		deactivateBtn := menu.Data("Deactivate", "student_deactivate", id)

		row := menu.Row(
			deactivateBtn,
		)
		msg := fmt.Sprintf(
			"Student ID: %d\n"+
				"Telegram ID: %d\n"+
				"Student name: %s\n"+
				"Active: %t",
			student.ID,
			student.TelegramID,
			student.Name,
			student.IsActive,
		)
		studentRows = append(studentRows, row)
		result = append(result, msg)
	}

	response := []string{"Here are your active students\n"}
	response = append(response, result...)
	message := strings.Join(response, "\n")

	menu.Inline(
		studentRows...,
	)

	return message, menu
}

func buildStudentsActivateMessage(
	students []domain.Student,
) (string, *tele.ReplyMarkup) {
	menu := &tele.ReplyMarkup{}
	result := make([]string, 0, len(students))
	var studentRows []tele.Row
	for _, student := range students {
		id := strconv.FormatInt(student.ID, 10)
		activateBtn := menu.Data("Activate", "student_activate", id)

		row := menu.Row(
			activateBtn,
		)
		msg := fmt.Sprintf(
			"Student ID: %d\n"+
				"Telegram ID: %d\n"+
				"Student name: %s\n"+
				"Active: %t",
			student.ID,
			student.TelegramID,
			student.Name,
			student.IsActive,
		)
		studentRows = append(studentRows, row)
		result = append(result, msg)
	}

	response := []string{"Here are your inactive students\n"}
	response = append(response, result...)
	message := strings.Join(response, "\n")

	menu.Inline(
		studentRows...,
	)

	return message, menu
}

func buildStudentLessonsMessage(
	count int64,
	lessons []domain.Lesson,
	cfg config.Config,
) string {
	var builder strings.Builder

	total := fmt.Sprintf("Total lessons: %d\n\n", count)
	builder.WriteString("My lessons\n\n")
	builder.WriteString(total)
	builder.WriteString("Upcoming lessons\n\n")

	if len(lessons) == 0 {
		builder.WriteString("You don't have any upcoming lessons")
	}

	for _, lesson := range lessons {
		lesson.DateTime = lesson.DateTime.In(cfg.TimeZone)

		_, _ = fmt.Fprintf(
			&builder,
			"Date: %s\n"+
				"Platform: %s\n"+
				"URL: %s\n\n",
			lesson.DateTime.Format("02-01-2006 15:04"),
			lesson.Platform,
			lesson.URL,
		)
	}

	return builder.String()
}
