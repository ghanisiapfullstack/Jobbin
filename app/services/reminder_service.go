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

	r.sendReminders(today, "reminder_sent_day_of", "day_of")
	r.sendReminders(tomorrow, "reminder_sent_day_before", "day_before")
}

func (r *ReminderService) sendReminders(date, sentColumn, reminderType string) {
	var applications []models.Application
	if err := facades.Orm().Query().
		With("User").
		Where("reminder_date", date).
		Where(sentColumn, false).
		Where("is_archived", false).
		Find(&applications); err != nil {
		facades.Log().Warningf("Failed to load %s reminders: %v", reminderType, err)
		return
	}

	for _, app := range applications {
		if app.User.ID == 0 {
			continue
		}

		if err := r.emailService.SendReminderEmail(
			app.User.Email, app.User.Name,
			app.JobTitle, app.Company,
			reminderType,
		); err != nil {
			facades.Log().Warningf("Failed to send %s reminder for app %d: %v", reminderType, app.ID, err)
			continue
		}

		if _, err := facades.Orm().Query().Model(&models.Application{}).Where("id", app.ID).Update(sentColumn, true); err != nil {
			facades.Log().Warningf("Failed to mark %s reminder for app %d as sent: %v", reminderType, app.ID, err)
			continue
		}
		facades.Log().Infof("Reminder %s sent for app %d (%s @ %s)", reminderType, app.ID, app.JobTitle, app.Company)
	}
}
