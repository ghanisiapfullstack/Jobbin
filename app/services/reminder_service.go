package services

import (
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"jobbin/backend/app/models"
)

type ReminderService struct {
	emailService *EmailService
}

func NewReminderService() *ReminderService {
	return &ReminderService{
		emailService: NewEmailService(),
	}
}

// SendDailyReminders cron job harian jam 07.00
func (r *ReminderService) SendDailyReminders() {
	today := carbon.Now().ToDateString()
	tomorrow := carbon.Now().AddDay().ToDateString()

	// Kirim reminder hari H
	r.sendReminderDayOf(today)

	// Kirim reminder H-1
	r.sendReminderDayBefore(tomorrow)
}

func (r *ReminderService) sendReminderDayOf(today string) {
	var applications []models.Application
	facades.Orm().Query().
		Where("reminder_date", today).
		Where("reminder_sent_day_of", false).
		Where("is_archived", false).
		Find(&applications)

	for _, app := range applications {
		// Ambil user
		var user models.User
		if err := facades.Orm().Query().Find(&user, app.UserID); err != nil || user.ID == 0 {
			continue
		}

		if err := r.emailService.SendReminderEmail(
			user.Email, user.Name,
			app.JobTitle, app.Company,
			"day_of",
		); err != nil {
			facades.Log().Warningf("Failed to send day_of reminder for app %d: %v", app.ID, err)
			continue
		}

		// Update flag
		app.ReminderSentDayOf = true
		facades.Orm().Query().Save(&app)
		facades.Log().Infof("Reminder day_of sent for app %d (%s @ %s)", app.ID, app.JobTitle, app.Company)
	}
}

func (r *ReminderService) sendReminderDayBefore(tomorrow string) {
	var applications []models.Application
	facades.Orm().Query().
		Where("reminder_date", tomorrow).
		Where("reminder_sent_day_before", false).
		Where("is_archived", false).
		Find(&applications)

	for _, app := range applications {
		var user models.User
		if err := facades.Orm().Query().Find(&user, app.UserID); err != nil || user.ID == 0 {
			continue
		}

		if err := r.emailService.SendReminderEmail(
			user.Email, user.Name,
			app.JobTitle, app.Company,
			"day_before",
		); err != nil {
			facades.Log().Warningf("Failed to send day_before reminder for app %d: %v", app.ID, err)
			continue
		}

		app.ReminderSentDayBefore = true
		facades.Orm().Query().Save(&app)
		facades.Log().Infof("Reminder day_before sent for app %d (%s @ %s)", app.ID, app.JobTitle, app.Company)
	}
}
