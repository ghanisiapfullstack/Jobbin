package bootstrap

import (
	contractsconsole "github.com/goravel/framework/contracts/console"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	contractsschedule "github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/schedule"

	appconsole "jobbin/backend/app/console"
	"jobbin/backend/app/facades"
	"jobbin/backend/app/services"
	"jobbin/backend/config"
	"jobbin/backend/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithCommands(Commands).
		WithMigrations(Migrations).
		WithRouting(func() {
			routes.Web()
			routes.Grpc()
			routes.Api()
		}).
		WithSchedule(Schedules).
		WithProviders(Providers).
		WithConfig(config.Boot).
		Create()
}

func Commands() []contractsconsole.Command {
	return []contractsconsole.Command{
		appconsole.NewSendRemindersCommand(),
	}
}

func Schedules() []contractsschedule.Event {
	if !facades.Config().GetBool("app.in_process_jobs", true) {
		return nil
	}

	reminderSvc := services.NewReminderService()
	auditSvc := services.NewAuditService()
	sessionSvc := services.NewSessionService()
	return []contractsschedule.Event{
		// Kirim reminder email setiap hari jam 07.00
		schedule.NewCallbackEvent(func() {
			if err := reminderSvc.SendDailyReminders(); err != nil {
				facades.Log().Warningf("Daily reminders completed with errors: %v", err)
			}
		}).DailyAt("07:00"),
		// Cleanup audit logs lebih dari 90 hari — setiap hari jam 03.00
		schedule.NewCallbackEvent(func() {
			_ = auditSvc.CleanupOldLogs()
		}).DailyAt("03:00"),
		// Cleanup expired/revoked refresh sessions every day.
		schedule.NewCallbackEvent(func() {
			_ = sessionSvc.CleanupExpired()
		}).DailyAt("03:15"),
	}
}
