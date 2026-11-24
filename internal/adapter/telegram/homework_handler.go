package telegram

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nullphosphor/english-tutor-bot/internal/config"
	"github.com/nullphosphor/english-tutor-bot/internal/domain"
	tele "gopkg.in/telebot.v4"
)

func buildStudentHomeworkMessage(
	cfg config.Config,
	homework []domain.Homework,
	backButtonText, backCallback string,
) (string, *tele.ReplyMarkup) {
	menu := &tele.ReplyMarkup{}
	backBtn := menu.Data(backButtonText, backCallback)
	menu.Inline(
		tele.Row{backBtn},
	)

	if len(homework) == 0 {
		return "You don't have any homework yet.", menu
	}

	result := make([]string, 0, len(homework))
	for _, hw := range homework {
		deadlineText := ""
		if hw.DueAt == nil {
			deadlineText = "Not specified"
		} else {
			in := hw.DueAt.In(cfg.TimeZone)
			deadlineText = in.Format("02-01-2006 15:04")
		}

		msg := fmt.Sprintf(
			"Title: %s\n"+
				"Description: %s\n"+
				"Deadline: %s\n",
			hw.Title,
			hw.Description,
			deadlineText,
		)
		result = append(result, msg)
	}
	response := []string{"Here are your homework assignments\n"}
	response = append(response, result...)
	message := strings.Join(response, "\n")

	return message, menu
}

func buildAdminStudentHomeworkMessage(
	cfg config.Config,
	homework []domain.Homework,
) (string, *tele.ReplyMarkup) {
	menu := &tele.ReplyMarkup{}
	backBtn := menu.Data("Back to students", "manage_homework")

	if len(homework) == 0 {
		menu.Inline(
			menu.Row(backBtn),
		)
		return "This student doesn't have any homework yet.", menu
	}

	result := make([]string, 0, len(homework))
	var homeworkRows []tele.Row
	for _, hw := range homework {
		id := strconv.FormatInt(hw.ID, 10)
		studentID := strconv.FormatInt(hw.StudentID, 10)
		version := strconv.FormatInt(hw.Version, 10)

		deadlineText := ""
		if hw.DueAt == nil {
			deadlineText = "Not specified"
		} else {
			in := hw.DueAt.In(cfg.TimeZone)
			deadlineText = in.Format("02-01-2006 15:04")
		}

		editBtn := menu.Data("Edit", "homework_edit", id, studentID, version)
		deleteBtn := menu.Data("Delete", "homework_delete", id, studentID, version)

		row := menu.Row(
			editBtn, deleteBtn,
		)

		msg := fmt.Sprintf(
			"Title: %s\n"+
				"Description: %s\n"+
				"Deadline: %s\n",
			hw.Title,
			hw.Description,
			deadlineText,
		)
		result = append(result, msg)
		homeworkRows = append(homeworkRows, row)
	}
	homeworkRows = append(homeworkRows, menu.Row(backBtn))

	response := []string{"Here are your homework assignments\n"}
	response = append(response, result...)
	message := strings.Join(response, "\n")

	menu.Inline(
		homeworkRows...,
	)

	return message, menu
}

func buildAdminHomeworkEditMenu(
	cfg config.Config,
	state HomeworkCreationState,
) (string, *tele.ReplyMarkup) {
	menu := &tele.ReplyMarkup{}
	editTitleBtn := menu.Data("Change title", "homework_edit_title")
	editDescriptionBtn := menu.Data("Change description", "homework_edit_description")
	editDeadlineBtn := menu.Data("Change deadline", "homework_edit_due")
	reviewChangesBtn := menu.Data("Review changes", "homework_review_changes")
	cancelBtn := menu.Data("Cancel", "homework_cancel")

	menu.Inline(
		menu.Row(editTitleBtn),
		menu.Row(editDescriptionBtn),
		menu.Row(editDeadlineBtn),
		menu.Row(reviewChangesBtn),
		menu.Row(cancelBtn),
	)

	deadlineText := ""
	if state.DueAt == nil {
		deadlineText = "Not specified"
	} else {
		in := state.DueAt.In(cfg.TimeZone)
		deadlineText = in.Format("02-01-2006 15:04")
	}

	msg := fmt.Sprintf(
		"Edit homework\n\n"+
			"Title: %s\n"+
			"Description: %s\n"+
			"Deadline: %s",
		state.Title,
		state.Description,
		deadlineText,
	)

	return msg, menu
}
