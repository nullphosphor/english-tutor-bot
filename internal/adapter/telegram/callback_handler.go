package telegram

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/nullphosphor/english-tutor-bot/internal/config"
	"github.com/nullphosphor/english-tutor-bot/internal/domain"
	"github.com/nullphosphor/english-tutor-bot/internal/service"
	tele "gopkg.in/telebot.v4"
)

func (b *Bot) CallbackHandler(
	ctx context.Context,
	lessonCreationManager *LessonCreationManager,
	lessonService *service.LessonService,
	studentService *service.StudentService,
	studentInviteService *service.StudentInviteService,
	inviteManager *StudentInviteCreationManager,
	homeworkCreationManager *HomeworkCreationManager,
	homeworkService *service.HomeworkService,
	cfg config.Config,
	adminTelegramID int64,
) {
	b.Handle(tele.OnCallback, func(c tele.Context) error {
		callbackPayload := ""
		callbackPayload2 := ""
		callbackPayload3 := ""
		args := c.Args()

		if len(args) < 1 {
			return c.Send("cannot get args")
		}

		if len(args) >= 2 {
			callbackPayload = args[1]
		}
		if len(args) >= 3 {
			callbackPayload2 = args[2]
		}
		if len(args) >= 4 {
			callbackPayload3 = args[3]
		}

		args[0] = strings.TrimPrefix(args[0], "\f")
		args[0] = strings.TrimSpace(args[0])
		button := args[0]

		if err := c.Respond(); err != nil {
			return fmt.Errorf("respond to callback: %w", err)
		}

		id := c.Callback().Sender.ID

		if id == adminTelegramID {
			switch button {
			case "":
				return c.Send("invalid button.")
			case "homework_edit":
				homeworkID, err := strconv.ParseInt(callbackPayload, 10, 64)
				if err != nil {
					return c.Send("Invalid homework ID. Please enter a valid numeric ID.")
				}

				if homeworkID <= 0 {
					return c.Send("Homework ID must be a positive number.")
				}

				studentID, err := strconv.ParseInt(callbackPayload2, 10, 64)
				if err != nil {
					return c.Send("Invalid student ID. Please enter a valid numeric ID.")
				}

				if studentID <= 0 {
					return c.Send("student ID must be a positive number.")
				}

				version, err := strconv.ParseInt(callbackPayload3, 10, 64)
				if err != nil {
					return c.Send("Please try again.")
				}

				if version <= 0 {
					return c.Send("Please try again.")
				}

				homework, err := homeworkService.GetHomeworkByID(
					ctx,
					homeworkID,
				)
				if err != nil {
					if errors.Is(err, domain.ErrNotFound) {
						homeworks, errHomeworks := homeworkService.GetHomeworksByStudentID(ctx, studentID)
						if errHomeworks != nil {
							return fmt.Errorf("get student homeworks: %w", errHomeworks)
						}
						msg, menu := buildAdminStudentHomeworkMessage(cfg, homeworks)
						message := "This homework has changed or is no longer available. The list has been refreshed.\n\n"
						message += msg

						return c.Edit(message, menu)
					}
					return fmt.Errorf("get student homework: %w", err)
				}

				if homework.StudentID != studentID || homework.Version != version {
					homeworks, errHomeworks := homeworkService.GetHomeworksByStudentID(ctx, studentID)
					if errHomeworks != nil {
						return fmt.Errorf("get student homeworks: %w", errHomeworks)
					}
					msg, menu := buildAdminStudentHomeworkMessage(cfg, homeworks)
					message := "This homework has changed or is no longer available. The list has been refreshed.\n\n"
					message += msg

					return c.Edit(message, menu)
				}

				_, errHomeworkState := homeworkCreationManager.GetState(id)
				if errHomeworkState != nil {
					if !errors.Is(errHomeworkState, ErrDialogNotFound) {
						return fmt.Errorf("get homework state: %w", errHomeworkState)
					}
				}
				if errHomeworkState == nil {
					return c.Send("You must first complete or cancel the current homework operation.")
				}

				if ok := inviteManager.IsActive(id); ok {
					return c.Send("You must first complete or cancel the student invitation creation.")
				}

				_, errLessonState := lessonCreationManager.GetState(id)
				if errLessonState != nil {
					if !errors.Is(errLessonState, ErrDialogNotFound) {
						return fmt.Errorf("get lesson state: %w", errLessonState)
					}
				}

				if errLessonState == nil {
					return c.Send("You are already creating a lesson. Please complete or cancel it first.")
				}

				homeworkCreationManager.StartEdit(adminTelegramID, homework)

				state, err := homeworkCreationManager.GetState(adminTelegramID)
				if err != nil {
					if errors.Is(err, ErrDialogNotFound) {
						return c.Send("There is no active homework operation to cancel.")
					}
					return fmt.Errorf("get state: %w", err)
				}

				msg, menu := buildAdminHomeworkEditMenu(cfg, state)

				return c.Edit(msg, menu)
			case "homework_edit_title":
				menu := &tele.ReplyMarkup{}
				cancelBtn := menu.Data("Cancel", "homework_cancel")

				menu.Inline(
					menu.Row(cancelBtn),
				)

				if err := homeworkCreationManager.StartEditingField(
					adminTelegramID,
					HomeworkStepWaitingTitle,
				); err != nil {
					if errors.Is(err, ErrDialogNotFound) {
						return c.Send("This editing session has expired. Please open the homework again.")
					}

					if errors.Is(err, ErrNotEditing) {
						return c.Send("You are not currently editing a homework assignment.")
					}

					if errors.Is(err, ErrInvalidDialogStep) {
						return c.Send("Please choose a field from the editing menu.")
					}

					return fmt.Errorf("start editing field: %w", err)
				}

				return c.Edit("Please enter the new homework title.", menu)
			case "homework_edit_description":
				menu := &tele.ReplyMarkup{}
				cancelBtn := menu.Data("Cancel", "homework_cancel")

				menu.Inline(
					menu.Row(cancelBtn),
				)

				if err := homeworkCreationManager.StartEditingField(
					adminTelegramID,
					HomeworkStepWaitingDescription,
				); err != nil {
					if errors.Is(err, ErrDialogNotFound) {
						return c.Send("This editing session has expired. Please open the homework again.")
					}

					if errors.Is(err, ErrNotEditing) {
						return c.Send("You are not currently editing a homework assignment.")
					}

					if errors.Is(err, ErrInvalidDialogStep) {
						return c.Send("Please choose a field from the editing menu.")
					}

					return fmt.Errorf("start editing field: %w", err)
				}
				return c.Edit("Please enter the new homework description.", menu)
			case "homework_edit_due":
				menu := &tele.ReplyMarkup{}
				cancelBtn := menu.Data("Cancel", "homework_cancel")

				menu.Inline(
					menu.Row(cancelBtn),
				)

				if err := homeworkCreationManager.StartEditingField(
					adminTelegramID,
					HomeworkStepWaitingDue,
				); err != nil {
					if errors.Is(err, ErrDialogNotFound) {
						return c.Send("This editing session has expired. Please open the homework again.")
					}

					if errors.Is(err, ErrNotEditing) {
						return c.Send("You are not currently editing a homework assignment.")
					}

					if errors.Is(err, ErrInvalidDialogStep) {
						return c.Send("Please choose a field from the editing menu.")
					}

					return fmt.Errorf("start editing field: %w", err)
				}
				return c.Edit("Please enter the new homework deadline (DD-MM-YYYY HH:MM), or type skip.", menu)
			case "homework_review_changes":
				menu := &tele.ReplyMarkup{}

				confirmBtn := menu.Data("Confirm", "homework_update_confirm")
				cancelBtn := menu.Data("Cancel", "homework_cancel")

				menu.Inline(
					menu.Row(confirmBtn, cancelBtn),
				)

				if err := homeworkCreationManager.BeginEditConfirmation(adminTelegramID); err != nil {
					if errors.Is(err, ErrDialogNotFound) {
						return c.Send("This editing session has expired. Please open the homework again.")
					}

					if errors.Is(err, ErrNotEditing) {
						return c.Send("You are not currently editing a homework assignment.")
					}

					if errors.Is(err, ErrInvalidDialogStep) {
						return c.Send("Please choose a field from the editing menu.")
					}

					return fmt.Errorf("begin editing confirmation: %w", err)
				}

				state, err := homeworkCreationManager.GetState(adminTelegramID)
				if err != nil {
					return fmt.Errorf("get state: %w", err)
				}

				deadlineText := ""
				if state.DueAt == nil {
					deadlineText = "Not specified"
				} else {
					in := state.DueAt.In(cfg.TimeZone)
					deadlineText = in.Format("02-01-2006 15:04")
				}

				msg := fmt.Sprintf(
					"Please review the changes\n\n"+
						"Title: %s\n"+
						"Description: %s\n"+
						"Deadline: %s",
					state.Title,
					state.Description,
					deadlineText,
				)

				return c.Edit(msg, menu)
			case "homework_update_confirm":
				state, err := homeworkCreationManager.BeginCreation(adminTelegramID)
				if err != nil {
					if errors.Is(err, ErrDialogNotFound) {
						return c.Send("Please start by running /create_homework.")
					}
					if errors.Is(err, ErrInvalidDialogStep) {
						return c.Send("This homework is no longer awaiting confirmation.")
					}
					return fmt.Errorf("begin homework creation: %w", err)
				}

				homeworkDomain := domain.NewHomework(
					state.StudentID,
					state.Title,
					state.Description,
					state.DueAt,
				)

				homeworkDomain.ID = state.HomeworkID
				homeworkDomain.Version = state.Version

				if _, err := homeworkService.UpdateHomework(ctx, homeworkDomain); err != nil {
					homeworkCreationManager.FinishCreation(adminTelegramID)
					if errors.Is(err, domain.ErrConflict) {
						homeworks, errGetHomework := homeworkService.GetHomeworksByStudentID(
							ctx,
							state.StudentID,
						)
						if errGetHomework != nil {
							return fmt.Errorf("get homeworks by student id: %w", errGetHomework)
						}

						msg, menu := buildAdminStudentHomeworkMessage(cfg, homeworks)
						message := "This homework has changed or is no longer available. The list has been refreshed.\n\n"
						message += msg

						return c.Edit(message, menu)
					}
					return fmt.Errorf("update homework: %w", err)
				}

				homeworkCreationManager.FinishCreation(adminTelegramID)
				homeworks, errGetHomework := homeworkService.GetHomeworksByStudentID(
					ctx,
					state.StudentID,
				)
				if errGetHomework != nil {
					return fmt.Errorf("get homeworks by student id: %w", errGetHomework)
				}

				msg, menu := buildAdminStudentHomeworkMessage(cfg, homeworks)
				message := "Homework was successfully updated.\n\n"
				message += msg

				return c.Edit(message, menu)
			case "homework_delete":
				homeworkID, err := strconv.ParseInt(callbackPayload, 10, 64)
				if err != nil {
					return c.Send("Invalid homework ID. Please enter a valid numeric ID.")
				}

				if homeworkID <= 0 {
					return c.Send("Homework ID must be a positive number.")
				}

				studentID, err := strconv.ParseInt(callbackPayload2, 10, 64)
				if err != nil {
					return c.Send("Invalid student ID. Please enter a valid numeric ID.")
				}

				if studentID <= 0 {
					return c.Send("student ID must be a positive number.")
				}

				version, err := strconv.ParseInt(callbackPayload3, 10, 64)
				if err != nil {
					return c.Send("Please try again.")
				}

				if version <= 0 {
					return c.Send("Please try again.")
				}

				menu := &tele.ReplyMarkup{}
				deleteBtn := menu.Data("Yes, delete", "delete_homework", callbackPayload, callbackPayload2, callbackPayload3)
				noDeleteBtn := menu.Data("No, go back", "student_homework", callbackPayload2)

				menu.Inline(
					menu.Row(deleteBtn, noDeleteBtn),
				)

				return c.Edit("Are you sure you want to delete this homework?", menu)
			case "delete_homework":
				homeworkID, err := strconv.ParseInt(callbackPayload, 10, 64)
				if err != nil {
					return c.Send("Invalid homework ID. Please enter a valid numeric ID.")
				}

				if homeworkID <= 0 {
					return c.Send("Homework ID must be a positive number.")
				}

				studentID, err := strconv.ParseInt(callbackPayload2, 10, 64)
				if err != nil {
					return c.Send("Invalid student ID. Please enter a valid numeric ID.")
				}

				if studentID <= 0 {
					return c.Send("student ID must be a positive number.")
				}

				version, err := strconv.ParseInt(callbackPayload3, 10, 64)
				if err != nil {
					return c.Send("Please try again.")
				}

				if version <= 0 {
					return c.Send("Please try again.")
				}

				if err := homeworkService.DeleteHomework(
					ctx,
					homeworkID,
					version,
				); err != nil {
					if errors.Is(err, domain.ErrInvalidArgument) {
						return c.Send("Invalid homework ID. Please enter a valid numeric ID.")
					}

					if errors.Is(err, domain.ErrConflict) {
						homework, errGetHomework := homeworkService.GetHomeworksByStudentID(
							ctx,
							studentID,
						)
						if errGetHomework != nil {
							return fmt.Errorf("get homeworks by student id: %w", errGetHomework)
						}

						msg, menu := buildAdminStudentHomeworkMessage(cfg, homework)
						message := "This homework has changed or is no longer available. The list has been refreshed.\n\n"
						message += msg

						return c.Edit(message, menu)
					}
					return fmt.Errorf("delete homework: %w", err)
				}

				homeworks, err := homeworkService.GetHomeworksByStudentID(
					ctx,
					studentID,
				)
				if err != nil {
					return fmt.Errorf("get homework by student ID: %w", err)
				}

				msg, menu := buildAdminStudentHomeworkMessage(cfg, homeworks)
				message := "Homework was successfully deleted.\n\n"
				message += msg

				return c.Edit(message, menu)
			case "student_homework":
				studentID, err := strconv.ParseInt(callbackPayload, 10, 64)
				if err != nil {
					return c.Send("Invalid student ID. Please enter a valid numeric ID.")
				}

				if studentID <= 0 {
					return c.Send("Student ID must be a positive number.")
				}

				homework, err := homeworkService.GetHomeworksByStudentID(ctx, studentID)
				if err != nil {
					return fmt.Errorf("get homework by student ID: %w", err)
				}

				msg, menu := buildAdminStudentHomeworkMessage(cfg, homework)

				return c.Edit(msg, menu)
			case "manage_homework":
				students, err := studentService.GetStudents(ctx, true)
				if err != nil {
					return fmt.Errorf("get students: %w", err)
				}

				msg, menu := buildStudentsHomeworkMessage(students)

				return c.Edit(msg, menu)
			case "invite_cancel":
				if ok := inviteManager.IsActive(id); !ok {
					return c.Send("There is no active student invitation to cancel.")
				}

				inviteManager.Finish(id)

				return c.Edit("Student invitation creation was cancelled.")
			case "today_lessons":
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
			case "main_menu":
				firstName := c.Callback().Sender.FirstName
				if firstName == "" {
					firstName = "admin"
				}

				welcomeMsg := fmt.Sprintf(
					"Welcome, %s! Choose an action.",
					firstName,
				)

				menu := buildAdminMainMenu()
				return c.Edit(welcomeMsg, menu)
			case "add_student":
				menu := &tele.ReplyMarkup{}
				cancelBtn := menu.Data("Cancel", "invite_cancel")

				menu.Inline(
					menu.Row(cancelBtn),
				)

				_, errHomework := homeworkCreationManager.GetState(id)
				if errHomework != nil {
					if !errors.Is(errHomework, ErrDialogNotFound) {
						return fmt.Errorf("get homework state: %w", errHomework)
					}
				}
				if errHomework == nil {
					return c.Send("You must first complete or cancel the creation of your homework.")
				}

				_, errLesson := lessonCreationManager.GetState(id)
				if errLesson != nil {
					if !errors.Is(errLesson, ErrDialogNotFound) {
						return fmt.Errorf("get lesson state: %w", errLesson)
					}
				}
				if errLesson == nil {
					return c.Send("You must first complete or cancel the creation of your lesson.")
				}

				if ok := inviteManager.IsActive(id); ok {
					return c.Send("You are already creating a student invitation. Please complete or cancel it first.")
				}

				inviteManager.Start(id)
				return c.Send("Please enter a student name", menu)
			case "create_lesson":
				menu := &tele.ReplyMarkup{}
				cancelBtn := menu.Data("Cancel", "lesson_cancel")

				menu.Inline(
					menu.Row(cancelBtn),
				)

				_, errHomeworkState := homeworkCreationManager.GetState(id)
				if errHomeworkState != nil {
					if !errors.Is(errHomeworkState, ErrDialogNotFound) {
						return fmt.Errorf("get homework state: %w", errHomeworkState)
					}
				}
				if errHomeworkState == nil {
					return c.Send("You must first complete or cancel the creation of your homework.")
				}

				if ok := inviteManager.IsActive(id); ok {
					return c.Send("You must first complete or cancel the student invitation creation.")
				}

				_, errLessonState := lessonCreationManager.GetState(id)
				if errLessonState != nil {
					if !errors.Is(errLessonState, ErrDialogNotFound) {
						return fmt.Errorf("get lesson state: %w", errLessonState)
					}
				}

				if errLessonState == nil {
					return c.Send("You are already creating a lesson. Please complete or cancel it first.")
				}

				lessonCreationManager.Start(id)

				return c.Send("Please enter the student's Telegram ID.", menu)
			case "lesson_platform":
				menu := &tele.ReplyMarkup{}
				cancelBtn := menu.Data("Cancel", "lesson_cancel")

				menu.Inline(
					menu.Row(cancelBtn),
				)

				if callbackPayload != "Zoom" &&
					callbackPayload != "Google Meet" &&
					callbackPayload != "Yandex Telemost" {
					return c.Send("Invalid platform")
				}

				if err := lessonCreationManager.UpdatePlatform(
					id,
					callbackPayload,
					LessonStepWaitingURL,
				); err != nil {
					if errors.Is(err, ErrDialogNotFound) {
						return c.Send("Please start by running /create_lesson.")
					}
					return fmt.Errorf("update platform: %w", err)
				}

				return c.Edit("Please enter the meeting URL.", menu)
			case "create_homework":
				menu := &tele.ReplyMarkup{}
				cancelBtn := menu.Data("Cancel", "homework_cancel")

				menu.Inline(
					menu.Row(cancelBtn),
				)

				_, err := lessonCreationManager.GetState(id)
				if err != nil {
					if !errors.Is(err, ErrDialogNotFound) {
						return fmt.Errorf("get lesson state: %w", err)
					}
				}
				if err == nil {
					return c.Send("You must first complete or cancel the creation of your lesson.")
				}

				_, errHomeworkState := homeworkCreationManager.GetState(id)
				if errHomeworkState != nil {
					if !errors.Is(errHomeworkState, ErrDialogNotFound) {
						return fmt.Errorf("get homework state: %w", errHomeworkState)
					}
				}
				if errHomeworkState == nil {
					return c.Send("You are already creating homework. Please complete or cancel it first.")
				}

				if ok := inviteManager.IsActive(id); ok {
					return c.Send("You must first complete or cancel the student invitation creation.")
				}

				homeworkCreationManager.Start(id)
				return c.Send("Please enter the student's Telegram ID.", menu)
			case "homework_confirm":
				menu := &tele.ReplyMarkup{}
				state, err := homeworkCreationManager.BeginCreation(adminTelegramID)
				if err != nil {
					if errors.Is(err, ErrDialogNotFound) {
						return c.Send("Please start by running /create_homework.")
					}
					if errors.Is(err, ErrInvalidDialogStep) {
						return c.Send("This homework is no longer awaiting confirmation.")
					}
					return fmt.Errorf("begin homework creation: %w", err)
				}

				homeworkDomain := domain.NewHomework(
					state.StudentID,
					state.Title,
					state.Description,
					state.DueAt,
				)

				homework, err := homeworkService.CreateHomework(
					ctx,
					homeworkDomain,
				)
				if err != nil {
					errReturnToConfirm := homeworkCreationManager.ReturnToConfirmation(
						adminTelegramID,
					)

					if errReturnToConfirm != nil {
						return fmt.Errorf(
							"create homework: %w",
							errors.Join(err, errReturnToConfirm),
						)
					}
					return fmt.Errorf("create homework: %w", err)
				}
				homeworkCreationManager.FinishCreation(
					adminTelegramID,
				)

				deadlineText := ""
				if homework.DueAt == nil {
					deadlineText = "Not specified"
				} else {
					in := homework.DueAt.In(cfg.TimeZone)
					deadlineText = in.Format("02-01-2006 15:04")
				}

				message := fmt.Sprintf(
					"The homework was successfully created!\n\n"+
						"Title: %s\n"+
						"Description: %s\n"+
						"Deadline: %s",
					homework.Title,
					homework.Description,
					deadlineText,
				)

				return c.Edit(message, menu)
			case "homework_cancel":
				menu := &tele.ReplyMarkup{}

				state, err := homeworkCreationManager.GetState(adminTelegramID)
				if err != nil {
					if errors.Is(err, ErrDialogNotFound) {
						return c.Send("There is no active homework operation to cancel.", menu)
					}
					return fmt.Errorf("get state: %w", err)
				}

				if err := homeworkCreationManager.CancelDialog(
					adminTelegramID,
				); err != nil {
					if errors.Is(err, ErrDialogNotFound) {
						return c.Send("There is no active homework operation to cancel.")
					}
					if errors.Is(err, ErrInvalidDialogStep) {
						return c.Send("This homework operation is already being processed and cannot be cancelled.")
					}
					return fmt.Errorf("begin homework cancel: %w", err)
				}

				if state.IsEditing {
					return c.Edit("Homework editing was cancelled.", menu)
				} else {
					return c.Edit("The homework creation was cancelled.", menu)
				}
			case "lesson_confirm":
				menu := &tele.ReplyMarkup{}
				state, err := lessonCreationManager.BeginCreation(adminTelegramID)
				if err != nil {
					if errors.Is(err, ErrDialogNotFound) {
						return c.Send("Please start by running /create_lesson.")
					}
					if errors.Is(err, ErrInvalidDialogStep) {
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

				lessonCreationManager.LessonFinish(adminTelegramID)

				lesson.DateTime = lesson.DateTime.In(cfg.TimeZone)

				message := fmt.Sprintf(
					"The lesson was successfully created\n\n"+
						"Student database ID: %d\n"+
						"Date: %s\n"+
						"Platform: %s\n"+
						"Meeting URL: %s",
					lesson.StudentID,
					lesson.DateTime.Format("02-01-2006 15:04"),
					lesson.Platform,
					lesson.URL,
				)

				return c.Edit(message, menu)
			case "lesson_cancel":
				menu := &tele.ReplyMarkup{}
				if err := lessonCreationManager.CancelDialog(adminTelegramID); err != nil {
					if errors.Is(err, ErrDialogNotFound) {
						return c.Send("Please start by running /create_lesson.")
					}

					if errors.Is(err, ErrInvalidDialogStep) {
						return c.Send("invalid step")
					}
					log.Printf("cancel dialog: %v", err)
					return c.Send("invalid error")
				}

				return c.Edit("The lesson creation was cancelled.", menu)
			case "start_confirm":
				menu := &tele.ReplyMarkup{}
				idLesson, err := strconv.ParseInt(callbackPayload, 10, 64)
				if err != nil {
					return c.Send("Invalid lesson ID. Please enter a valid numeric ID.")
				}

				if idLesson <= 0 {
					return c.Send("Lesson ID must be a positive number.")
				}

				if _, err := lessonService.StartLesson(ctx, idLesson); err != nil {
					if errors.Is(err, domain.ErrNotFound) {
						return c.Edit("Lesson not found", menu)
					}

					if errors.Is(err, domain.ErrInvalidArgument) {
						return c.Edit("Cannot start lesson in its current status", menu)
					}

					if errors.Is(err, domain.ErrConflict) {
						return c.Edit("This lesson has changed or is no longer available. Please run /today to refresh.", menu)
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
				menu := &tele.ReplyMarkup{}
				idLesson, err := strconv.ParseInt(callbackPayload, 10, 64)
				if err != nil {
					return c.Send("Invalid lesson ID. Please enter a valid numeric ID.")
				}

				if idLesson <= 0 {
					return c.Send("Lesson ID must be a positive number.")
				}

				if _, err := lessonService.FinishLesson(ctx, idLesson); err != nil {
					if errors.Is(err, domain.ErrNotFound) {
						return c.Edit("Lesson not found", menu)
					}

					if errors.Is(err, domain.ErrInvalidArgument) {
						return c.Edit("Cannot finish lesson in its current status", menu)
					}

					if errors.Is(err, domain.ErrConflict) {
						return c.Edit("This lesson has changed or is no longer available. Please run /today to refresh.", menu)
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

					if errors.Is(err, domain.ErrConflict) {
						return c.Edit("This lesson has changed or is no longer available. Please run /today to refresh.", menu)
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

				if len(lessons) == 0 {
					return c.Edit("Enjoy your day off! :) You don't have any lessons today", menu)
				}

				msg, menu := buildLessonsMessage(lessons, cfg)

				return c.Edit(msg, menu)
			case "student_invite_approve":
				menu := &tele.ReplyMarkup{}
				idInvite, err := strconv.ParseInt(callbackPayload, 10, 64)
				if err != nil {
					return c.Send("Invalid invite ID. Please enter a valid numeric ID.")
				}

				if idInvite <= 0 {
					return c.Send("Invite ID must be positive.")
				}

				student, err := studentInviteService.ApproveInvite(
					ctx,
					idInvite,
				)
				if err != nil {
					if errors.Is(err, domain.ErrInviteUnavailable) {
						return c.Send("Application has already been processed or is no longer awaiting a decision.")
					}

					if errors.Is(err, domain.ErrAlreadyExists) {
						return c.Send("student with this Telegram ID already exists.")
					}
					return fmt.Errorf("approve invite: %w", err)
				}
				response := fmt.Sprintf(
					"The student has been successfully registered\n\n"+
						"Telegram ID: %d\n"+
						"Student name:  %s\n"+
						"Active: %t\n",
					student.TelegramID,
					student.Name,
					student.IsActive,
				)
				return c.Edit(response, menu)
			case "student_invite_reject":
				menu := &tele.ReplyMarkup{}
				idInvite, err := strconv.ParseInt(callbackPayload, 10, 64)
				if err != nil {
					return c.Send("Invalid invite ID. Please enter a valid numeric ID.")
				}

				if idInvite <= 0 {
					return c.Send("Invite ID must be positive.")
				}

				invite, err := studentInviteService.RejectInvite(
					ctx,
					idInvite,
				)
				if err != nil {
					if errors.Is(err, domain.ErrInviteUnavailable) {
						return c.Send("Application has already been processed or is no longer awaiting a decision.")
					}

					return fmt.Errorf("reject invite: %w", err)
				}

				response := fmt.Sprintf(
					"Registration request rejected\n\n"+
						"Invite ID: %d\n"+
						"Student name: %s\n",
					invite.ID,
					invite.StudentName,
				)

				return c.Edit(response, menu)
			case "students_list":
				students, err := studentService.GetStudents(ctx, true)
				if err != nil {
					return fmt.Errorf("get students: %w", err)
				}

				msg, menu := buildStudentsMessage(students)
				if len(students) == 0 {
					return c.Edit("You don't have any active students yet", menu)
				}

				return c.Edit(msg, menu)
			case "student_deactivate":
				studentID, err := strconv.ParseInt(callbackPayload, 10, 64)
				if err != nil {
					return c.Send("Invalid student ID. Please enter a valid numeric ID.")
				}

				if studentID <= 0 {
					return c.Send("student ID must be positive.")
				}

				if _, err := studentService.SetStudentActive(
					ctx,
					false,
					studentID,
				); err != nil {
					if errors.Is(err, domain.ErrNotFound) {
						return c.Send("Student not found")
					}
					return fmt.Errorf("set student active: %w", err)
				}

				students, err := studentService.GetStudents(ctx, true)
				if err != nil {
					return fmt.Errorf("get students: %w", err)
				}

				msg, menu := buildStudentsMessage(students)
				if len(students) == 0 {
					return c.Edit("You don't have any active students yet", menu)
				}

				return c.Edit(msg, menu)
			case "inactive_students_list":
				students, err := studentService.GetStudents(ctx, false)
				if err != nil {
					return fmt.Errorf("get students: %w", err)
				}

				msg, menu := buildStudentsActivateMessage(students)
				if len(students) == 0 {
					return c.Edit("You don't have any inactive students yet", menu)
				}

				return c.Edit(msg, menu)
			case "student_activate":
				studentID, err := strconv.ParseInt(callbackPayload, 10, 64)
				if err != nil {
					return c.Send("Invalid student ID. Please enter a valid numeric ID.")
				}

				if studentID <= 0 {
					return c.Send("student ID must be positive.")
				}

				if _, err := studentService.SetStudentActive(
					ctx,
					true,
					studentID,
				); err != nil {
					if errors.Is(err, domain.ErrNotFound) {
						return c.Send("Student not found")
					}
					return fmt.Errorf("set student active: %w", err)
				}

				students, err := studentService.GetStudents(ctx, false)
				if err != nil {
					return fmt.Errorf("get students: %w", err)
				}

				msg, menu := buildStudentsActivateMessage(students)

				if len(students) == 0 {
					return c.Edit("You don't have any inactive students yet", menu)
				}

				return c.Edit(msg, menu)
			default:
				return c.Send("Unknown operation")
			}
		} else {
			student, err := studentService.GetStudentByTelegramID(
				ctx,
				id,
			)
			if err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					return c.Send("Welcome! It seems you are not registered. Please contact the admin to get an invitation link.")
				}
				return fmt.Errorf("get student by telegram id: %w", err)
			}

			if !student.IsActive {
				return c.Send("Your account is inactive. Please contact the admin for assistance.")
			}

			switch button {
			case "my_lesson":
				now := time.Now().In(cfg.TimeZone)

				countLessons, err := lessonService.CountStudentLessons(
					ctx,
					student.ID,
				)
				if err != nil {
					return fmt.Errorf("Cannot count student lessons: %w", err)
				}

				soonLessons, err := lessonService.GetUpcomingStudentLessons(
					ctx,
					student.ID,
					now,
				)
				if err != nil {
					return fmt.Errorf("Cannot get upcoming lessons: %w", err)
				}

				message, menu := buildStudentLessonsMessage(
					countLessons,
					soonLessons,
					cfg,
				)

				return c.Edit(message, menu)
			case "my_homework":
				homeworks, err := homeworkService.GetHomeworksByStudentID(
					ctx,
					student.ID,
				)
				if err != nil {
					return fmt.Errorf("get homework: %w", err)
				}

				msg, menu := buildStudentHomeworkMessage(cfg, homeworks, "Back to main menu", "main_menu")

				return c.Edit(msg, menu)
			case "main_menu":
				menu := &tele.ReplyMarkup{}

				lessonBtn := menu.Data("My lessons", "my_lesson")
				homeworkBtn := menu.Data("My homework", "my_homework")
				menu.Inline(
					menu.Row(lessonBtn, homeworkBtn),
				)

				message := fmt.Sprintf(
					"Hello, %s. Choose an action",
					student.Name,
				)

				return c.Edit(message, menu)

			default:
				return c.Send("Unknown operation")
			}
		}
	})
}
