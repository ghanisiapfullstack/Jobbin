package console

import (
	"errors"

	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/facades"

	"jobbin/backend/app/services"
)

// DailyJobsCommand menjalankan seluruh pekerjaan harian Jobbin dalam satu
// invocation. Command ini dirancang untuk dipanggil oleh scheduler eksternal
// (Azure Container Apps Job / cron) sehingga tidak bergantung pada web replica
// yang bisa scale-to-zero.
//
// Setiap langkah berjalan independen: kegagalan satu langkah tidak membatalkan
// langkah lain, dan seluruh error digabung agar exit code non-nol saat ada
// masalah (berguna untuk alerting di job runner).
type DailyJobsCommand struct{}

func NewDailyJobsCommand() *DailyJobsCommand {
	return &DailyJobsCommand{}
}

func (r *DailyJobsCommand) Signature() string {
	return "jobs:daily"
}

func (r *DailyJobsCommand) Description() string {
	return "Run all daily Jobbin jobs: send reminders, cleanup audit logs, cleanup expired sessions"
}

func (r *DailyJobsCommand) Extend() command.Extend {
	return command.Extend{Category: "jobbin"}
}

func (r *DailyJobsCommand) Handle(ctx contractsconsole.Context) error {
	var failures []error

	// 1. Kirim reminder email yang jatuh tempo (H-1 dan hari-H).
	if err := services.NewReminderService().SendDailyReminders(); err != nil {
		facades.Log().Warningf("jobs:daily reminders completed with errors: %v", err)
		failures = append(failures, err)
	} else {
		ctx.Info("Reminders processed")
	}

	// 2. Bersihkan audit log yang lebih tua dari 90 hari.
	if err := services.NewAuditService().CleanupOldLogs(); err != nil {
		facades.Log().Warningf("jobs:daily audit cleanup failed: %v", err)
		failures = append(failures, err)
	} else {
		ctx.Info("Audit logs cleaned up")
	}

	// 3. Bersihkan refresh session yang sudah expired/revoked.
	if err := services.NewSessionService().CleanupExpired(); err != nil {
		facades.Log().Warningf("jobs:daily session cleanup failed: %v", err)
		failures = append(failures, err)
	} else {
		ctx.Info("Expired sessions cleaned up")
	}

	if len(failures) > 0 {
		ctx.Error("One or more daily jobs failed")
		return errors.Join(failures...)
	}

	ctx.Success("All daily jobs completed")
	return nil
}
