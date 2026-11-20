package telegram

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/nullphosphor/english-tutor-bot/internal/domain"
	"github.com/nullphosphor/english-tutor-bot/internal/service"
	tele "gopkg.in/telebot.v4"
)

func (b *Bot) StartHandler(
	ctx context.Context,
	studentService *service.StudentService,
	studentInviteService *service.StudentInviteService,
	adminTelegramID int64,
) {
	b.Handle("/start", func(c tele.Context) error {
		id := c.Sender().ID

		menu := &tele.ReplyMarkup{}

		if id == adminTelegramID {
			studentsBtn := menu.Data("Students", "students_list")
			inactiveStudentsBtn := menu.Data("Inactive students", "inactive_students_list")
			menu.Inline(
				menu.Row(studentsBtn, inactiveStudentsBtn),
			)
			return c.Send("Welcome, admin! You can manage lessons and students using the available commands.", menu)
		}

		student, err := studentService.GetStudentByTelegramID(ctx, id)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				payload := c.Message().Payload
				if payload == "" {
					return c.Send("Welcome! It seems you are not registered. Please contact the admin to get an invitation link.")
				}

				studentInvite, errInvite := studentInviteService.RequestRegistration(
					ctx,
					payload,
					id,
				)
				if errInvite != nil {
					if errors.Is(errInvite, domain.ErrInviteUnavailable) {
						return c.Send("Link invite unavailable. Please contact the admin to get an invitation link.")
					}
					return fmt.Errorf("request registration: %w", errInvite)
				}

				studentInviteString := strconv.FormatInt(studentInvite.ID, 10)

				menu := &tele.ReplyMarkup{}

				approveBtn := menu.Data("Approve", "student_invite_approve", studentInviteString)
				rejectBtn := menu.Data("Reject", "student_invite_reject", studentInviteString)

				menu.Inline(
					menu.Row(approveBtn, rejectBtn),
				)

				if _, errSend := b.Send(tele.ChatID(adminTelegramID), "New student registration request:\n\n"+
					fmt.Sprintf("Telegram ID: %d\nInvite ID: %d\nName: %s\nUsername: @%s",
						id, studentInvite.ID, studentInvite.StudentName, c.Sender().Username), menu); errSend != nil {
					return fmt.Errorf("cannot send info about student: %w", errSend)
				}

				return c.Send("The registration application has been successfully submitted and is awaiting approval.")
			}

			return fmt.Errorf("get student by telegram id: %w", err)
		}

		if !student.IsActive {
			return c.Send("Your account is inactive. Please contact the admin for assistance.")
		}

		studentLessonsMenu := &tele.ReplyMarkup{}

		studentLessonBtn := studentLessonsMenu.Data("My lessons", "my_lesson")

		studentLessonsMenu.Inline(
			studentLessonsMenu.Row(studentLessonBtn),
		)

		return c.Send("Welcome! You are registered as a student. You can view your upcoming lessons and manage your account using the available commands.", studentLessonsMenu)

	})
}
