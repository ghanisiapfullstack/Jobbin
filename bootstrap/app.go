package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	contractsconfiguration "github.com/goravel/framework/contracts/foundation/configuration"
	contractsschedule "github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/schedule"

	"jobbin/backend/app/http/middleware"
	"jobbin/backend/app/services"
	"jobbin/backend/config"
	"jobbin/backend/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMigrations(Migrations).
		WithMiddleware(func(m contractsconfiguration.Middleware) {
			m.Append(middleware.NewCorsMiddleware().Handle())
		}).
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
	return []contractsschedule.Event{
		// Kirim reminder email setiap hari jam 07.00
		schedule.NewCallbackEvent(func() {
			reminderSvc.SendDailyReminders()
		}).DailyAt("07:00"),
	}
}
