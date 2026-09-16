package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	contractsschedule "github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/schedule"

	"jobbin/backend/app/services"
	"jobbin/backend/config"
	"jobbin/backend/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
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

func Schedules() []contractsschedule.Event {
	reminderSvc := services.NewReminderService()
	auditSvc := services.NewAuditService()
	sessionSvc := services.NewSessionService()
	return []contractsschedule.Event{
		// Kirim reminder email setiap hari jam 07.00
		schedule.NewCallbackEvent(func() {
			reminderSvc.SendDailyReminders()
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
