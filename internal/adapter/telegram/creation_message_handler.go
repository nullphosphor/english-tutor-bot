package telegram

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/nullphosphor/english-tutor-bot/internal/config"
	"github.com/nullphosphor/english-tutor-bot/internal/domain"
	"github.com/nullphosphor/english-tutor-bot/internal/service"
	tele "gopkg.in/telebot.v4"
)

func (b *Bot) CreationMessageHandler(
	ctx context.Context,
	lessonCreationManager *LessonCreationManager,
	homeworkCreationManager *HomeworkCreationManager,
	inviteManager *StudentInviteCreationManager,
	studentService *service.StudentService,
	studentInviteService *service.StudentInviteService,
	cfg config.Config,
	adminTelegramID int64,
) {
	b.Handle(tele.OnText, func(c tele.Context) error {
		id := c.Sender().ID

		if id != adminTelegramID {
			return c.Send("You aren't authorized to use this command.")
		}

		var lessonActive, homeworkActive bool = false, false

		lessonState, err := lessonCreationManager.GetState(id)
		if err != nil {
			if !errors.Is(err, ErrDialogNotFound) {
				return fmt.Errorf("get state: %w", err)
			}
			lessonActive = false
		}
		if err == nil {
			lessonActive = true
		}

		homeworkState, err := homeworkCreationManager.GetState(id)
		if err != nil {
			if !errors.Is(err, ErrDialogNotFound) {
				return fmt.Errorf("get state: %w", err)
			}
			homeworkActive = false
		}
		if err == nil {
			homeworkActive = true
		}

		ok := inviteManager.IsActive(id)

		switch {
		case (lessonActive && homeworkActive) ||
			(lessonActive && ok) ||
			(homeworkActive && ok):
			return c.Send("Conflict, both dialogues are active.")
		case lessonActive:
			if err := handleLessonCreationMessage(
				c, ctx, lessonCreationManager, lessonState,
				studentService, cfg, adminTelegramID,
			); err != nil {
				return fmt.Errorf("lesson creation: %w", err)
			}
		case homeworkActive:
			if err := handleHomeworkCreationMessage(
				c, ctx, homeworkCreationManager, homeworkState,
				studentService, cfg, adminTelegramID,
			); err != nil {
				return fmt.Errorf("homework creation: %w", err)
			}
		case ok:
			if err := handleStudentInviteCreationMessage(
				c, b, ctx, studentInviteService, inviteManager,
			); err != nil {
				return fmt.Errorf("invite creation: %w", err)
			}
		default:
			return c.Send("Please enter /create_lesson or /create_homework or /start")
		}
		return nil
	})
}

func handleLessonCreationMessage(
	c tele.Context,
	ctx context.Context,
	lessonCreationManager *LessonCreationManager,
	lessonState LessonCreationState,
	studentService *service.StudentService,
	cfg config.Config,
	adminTelegramID int64,
) error {
	switch lessonState.Step {
	case LessonStepWaitingStudentID:
		menu := &tele.ReplyMarkup{}
		cancelBtn := menu.Data("Cancel", "lesson_cancel")

		menu.Inline(
			menu.Row(cancelBtn),
		)

		text := c.Text()
		text = strings.TrimSpace(text)

		telegramID, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return c.Send("Invalid Telegram ID. Please enter a valid numeric ID.", menu)
		}

		if telegramID <= 0 {
			return c.Send("telegram ID must be more than 0", menu)
		}

		student, err := studentService.GetStudentByTelegramID(
			ctx,
			telegramID,
		)

		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return c.Send("Student not found. Please check the Telegram ID and try again.", menu)
			}
			return fmt.Errorf("get student: %w", err)
		}

		if !student.IsActive {
			return c.Send("This student is inactive.", menu)
		}

		if err := lessonCreationManager.UpdateState(
			adminTelegramID,
			student.ID,
			LessonStepWaitingDate,
		); err != nil {
			if errors.Is(err, ErrDialogNotFound) {
				return c.Send("Please start by running /create_lesson.")
			}
			return fmt.Errorf("update state: %w", err)
		}
		return c.Send("Please enter the lesson date and time (DD-MM-YYYY HH:MM).", menu)
	case LessonStepWaitingDate:
		menu := &tele.ReplyMarkup{}
		cancelBtn := menu.Data("Cancel", "lesson_cancel")
		menu.Inline(
			menu.Row(cancelBtn),
		)

		menuPlatform := &tele.ReplyMarkup{}
		zoomBtn := menuPlatform.Data("Zoom", "lesson_platform", "Zoom")
		meetBtn := menuPlatform.Data("Google Meet", "lesson_platform", "Google Meet")
		telemostBtn := menuPlatform.Data("Yandex Telemost", "lesson_platform", "Yandex Telemost")
		cancelPlatformButton := menuPlatform.Data("Cancel", "lesson_cancel")

		menuPlatform.Inline(
			menuPlatform.Row(zoomBtn),
			menuPlatform.Row(meetBtn),
			menuPlatform.Row(telemostBtn),
			menuPlatform.Row(cancelPlatformButton),
		)

		text := c.Text()

		date, err := time.ParseInLocation(
			"02-01-2006 15:04",
			text,
			cfg.TimeZone,
		)
		if err != nil {
			return c.Send("Please enter a valid date and time (DD-MM-YYYY HH:MM)", menu)
		}

		tz := time.Now().In(cfg.TimeZone)

		if !date.After(tz) {
			return c.Send("The lesson date must be in the future", menu)
		}

		if err := lessonCreationManager.UpdateDate(
			adminTelegramID,
			date,
			LessonStepWaitingPlatform,
		); err != nil {
			if errors.Is(err, ErrDialogNotFound) {
				return c.Send("Please start by running /create_lesson.")
			}
			return fmt.Errorf("update date: %w", err)
		}

		return c.Send("Please choose a meeting platform.", menuPlatform)
	case LessonStepWaitingPlatform:
		menuPlatform := &tele.ReplyMarkup{}
		zoomBtn := menuPlatform.Data("Zoom", "lesson_platform", "Zoom")
		meetBtn := menuPlatform.Data("Google Meet", "lesson_platform", "Google Meet")
		telemostBtn := menuPlatform.Data("Yandex Telemost", "lesson_platform", "Yandex Telemost")
		cancelPlatformButton := menuPlatform.Data("Cancel", "lesson_cancel")

		menuPlatform.Inline(
			menuPlatform.Row(zoomBtn),
			menuPlatform.Row(meetBtn),
			menuPlatform.Row(telemostBtn),
			menuPlatform.Row(cancelPlatformButton),
		)

		return c.Send("Please choose a meeting platform.", menuPlatform)
	case LessonStepWaitingURL:
		text := c.Text()

		text = strings.TrimSpace(text)
		if text == "" {
			return c.Send("Please enter a valid url.")
		}

		parsed, err := url.ParseRequestURI(text)
		if err != nil {
			return c.Send("Please enter a valid url.")
		}

		if parsed.User != nil {
			return c.Send("Invalid URL.")
		}

		if !strings.EqualFold(parsed.Scheme, "https") {
			return c.Send("Invalid scheme.")
		}

		if parsed.Hostname() == "" {
			return c.Send("Invalid hostname.")
		}

		host := strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))
		switch lessonState.Platform {
		case "Zoom":
			if host != "zoom.us" && !strings.HasSuffix(host, ".zoom.us") {
				return c.Send("Invalid hostname.")
			}
		case "Google Meet":
			if host != "meet.google.com" {
				return c.Send("Invalid hostname.")
			}
		case "Yandex Telemost":
			if host != "telemost.yandex.ru" {
				return c.Send("Invalid hostname.")
			}
		default:
			return c.Send("Invalid platform")
		}

		if err := lessonCreationManager.UpdateURL(
			adminTelegramID,
			text,
			LessonStepWaitingConfirm,
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

		message := fmt.Sprintf(
			"Please confirm the lesson:\n\n"+
				"Student database ID: %d\n"+
				"Date: %s\n"+
				"Platform: %s\n"+
				"Meeting URL: %s",
			lessonState.StudentID,
			lessonState.Date.Format("02-01-2006 15:04"),
			lessonState.Platform,
			text,
		)
		return c.Send(message, menu)
	default:
		return c.Send("Unknown step")
	}
}

func handleHomeworkCreationMessage(
	c tele.Context,
	ctx context.Context,
	homeworkCreationManager *HomeworkCreationManager,
	homeworkCreationState HomeworkCreationState,
	studentService *service.StudentService,
	cfg config.Config,
	adminTelegramID int64,
) error {
	switch homeworkCreationState.Step {
	case HomeworkStepWaitingStudentID:
		menu := &tele.ReplyMarkup{}
		cancelBtn := menu.Data("Cancel", "homework_cancel")

		menu.Inline(
			menu.Row(cancelBtn),
		)

		text := c.Text()
		text = strings.TrimSpace(text)

		telegramID, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return c.Send("Invalid Telegram ID. Please enter a valid numeric ID.", menu)
		}

		if telegramID <= 0 {
			return c.Send("telegram ID must be more than 0", menu)
		}

		student, err := studentService.GetStudentByTelegramID(
			ctx,
			telegramID,
		)

		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return c.Send("Student not found. Please check the Telegram ID and try again.", menu)
			}
			return fmt.Errorf("get student: %w", err)
		}

		if !student.IsActive {
			return c.Send("This student is inactive.", menu)
		}

		if err := homeworkCreationManager.UpdateStudentID(
			adminTelegramID,
			student.ID,
		); err != nil {
			if errors.Is(err, ErrDialogNotFound) {
				return c.Send("Please start by /create_homework")
			}
			return fmt.Errorf("update student id: %w", err)
		}
		return c.Send("Please enter the homework title", menu)
	case HomeworkStepWaitingTitle:
		menu := &tele.ReplyMarkup{}
		cancelBtn := menu.Data("Cancel", "homework_cancel")

		menu.Inline(
			menu.Row(cancelBtn),
		)

		text := c.Text()
		text = strings.TrimSpace(text)
		titleLen := len([]rune(text))
		if titleLen < 3 || titleLen > 30 {
			return c.Send("Please enter a title again.", menu)
		}

		if err := homeworkCreationManager.UpdateTitle(
			adminTelegramID,
			text,
		); err != nil {
			if errors.Is(err, ErrDialogNotFound) {
				return c.Send("Please start by /create_homework")
			}
			return fmt.Errorf("update title: %w", err)
		}

		state, err := homeworkCreationManager.GetState(adminTelegramID)
		if err != nil {
			return fmt.Errorf("get state: %w", err)
		}

		if state.IsEditing {
			msg, menu := buildAdminHomeworkEditMenu(cfg, state)
			return c.Send(msg, menu)
		} else {
			return c.Send("Please enter the homework description", menu)
		}
	case HomeworkStepWaitingDescription:
		menu := &tele.ReplyMarkup{}
		cancelBtn := menu.Data("Cancel", "homework_cancel")

		menu.Inline(
			menu.Row(cancelBtn),
		)

		text := c.Text()
		text = strings.TrimSpace(text)
		descriptionLen := len([]rune(text))
		if descriptionLen < 3 || descriptionLen > 3500 {
			return c.Send("Please enter a description again", menu)
		}

		if err := homeworkCreationManager.UpdateDescription(
			adminTelegramID,
			text,
		); err != nil {
			if errors.Is(err, ErrDialogNotFound) {
				return c.Send("Please start by /create_homework")
			}
			return fmt.Errorf("update description: %w", err)
		}

		state, err := homeworkCreationManager.GetState(adminTelegramID)
		if err != nil {
			return fmt.Errorf("get state: %w", err)
		}

		if state.IsEditing {
			msg, menu := buildAdminHomeworkEditMenu(cfg, state)
			return c.Send(msg, menu)
		} else {
			return c.Send("Please enter the homework deadline (DD-MM-YYYY HH:MM), or type skip.", menu)
		}
	case HomeworkStepWaitingDue:
		menu := &tele.ReplyMarkup{}
		menuCancel := &tele.ReplyMarkup{}
		mnCancelBtn := menuCancel.Data("Cancel", "homework_cancel")

		menuCancel.Inline(
			menuCancel.Row(mnCancelBtn),
		)

		text := c.Text()
		text = strings.TrimSpace(text)

		if text == "skip" {
			if err := homeworkCreationManager.UpdateDueAt(
				adminTelegramID,
				nil,
			); err != nil {
				if errors.Is(err, ErrDialogNotFound) {
					return c.Send("Please start by /create_homework")
				}
				return fmt.Errorf("update due: %w", err)
			}
		} else {
			date, err := time.ParseInLocation(
				"02-01-2006 15:04",
				text,
				cfg.TimeZone,
			)
			if err != nil {
				return c.Send("Please enter a valid date and time (DD-MM-YYYY HH:MM)", menuCancel)
			}

			tz := time.Now().In(cfg.TimeZone)

			if !date.After(tz) {
				return c.Send("The deadline date must be in the future", menuCancel)
			}
			if err := homeworkCreationManager.UpdateDueAt(
				adminTelegramID,
				&date,
			); err != nil {
				if errors.Is(err, ErrDialogNotFound) {
					return c.Send("Please start by /create_homework")
				}
				return fmt.Errorf("update due: %w", err)
			}
		}
		state, err := homeworkCreationManager.GetState(adminTelegramID)
		if err != nil {
			if errors.Is(err, ErrDialogNotFound) {
				return c.Send("Please start by /create_homework")
			}
			return fmt.Errorf("get state: %w", err)
		}

		if state.IsEditing {
			msg, menu := buildAdminHomeworkEditMenu(cfg, state)
			return c.Send(msg, menu)
		} else {
			deadlineText := ""
			if state.DueAt == nil {
				deadlineText = "Not specified"
			} else {
				in := state.DueAt.In(cfg.TimeZone)
				deadlineText = in.Format("02-01-2006 15:04")
			}

			confirmBtn := menu.Data("Confirm", "homework_confirm")
			cancelBtn := menu.Data("Cancel", "homework_cancel")

			menu.Inline(
				menu.Row(confirmBtn, cancelBtn),
			)

			message := fmt.Sprintf(
				"Please confirm the homework\n\n"+
					"Title: %s\n"+
					"Description: %s\n"+
					"Deadline: %s",
				state.Title,
				state.Description,
				deadlineText,
			)
			return c.Send(message, menu)
		}
	default:
		return c.Send("Please start by /create_lesson or /create_homework")
	}
}

func handleStudentInviteCreationMessage(
	c tele.Context,
	b *Bot,
	ctx context.Context,
	studentInviteService *service.StudentInviteService,
	inviteManager *StudentInviteCreationManager,
) error {
	menu := &tele.ReplyMarkup{}
	cancelBtn := menu.Data("Cancel", "invite_cancel")

	menu.Inline(
		menu.Row(cancelBtn),
	)

	text := c.Text()
	text = strings.TrimSpace(text)

	invName := domain.StudentInvite{StudentName: text}

	_, token, err := studentInviteService.CreateInvite(
		ctx,
		invName,
	)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidArgument) {
			return c.Send("Invalid student name.", menu)
		}

		if errors.Is(err, domain.ErrAlreadyExists) {
			return c.Send("Can't create a unique link. Please try again.", menu)
		}

		return fmt.Errorf("create invite: %w", err)
	}

	inviteManager.Finish(c.Sender().ID)

	botUsername := b.Me.Username
	response := fmt.Sprint("https://t.me/" + botUsername + "?start=" + token)

	return c.Send(response)
}
