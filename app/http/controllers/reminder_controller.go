package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"jobbin/backend/app/models"
	"jobbin/backend/app/services"
)

type ReminderController struct{}

func NewReminderController() *ReminderController {
	return &ReminderController{}
}

// Index GET /api/v1/reminders
func (r *ReminderController) Index(ctx http.Context) http.Response {
	userID, err := getAuthUserID(ctx)
	if err != nil {
		return ctx.Response().Json(401, http.Json{"message": "Unauthorized"})
	}

	today := carbon.Now().ToDateString()
	tomorrow := carbon.Now().AddDay().ToDateString()

	var todayApps []models.Application
	facades.Orm().Query().
		Where("user_id", userID).
		Where("reminder_date", today).
		Where("is_archived", false).
		Find(&todayApps)

	var tomorrowApps []models.Application
	facades.Orm().Query().
		Where("user_id", userID).
		Where("reminder_date", tomorrow).
		Where("is_archived", false).
		Find(&tomorrowApps)

	return ctx.Response().Json(200, http.Json{
		"message": "Berhasil",
		"data": map[string]interface{}{
			"today":    todayApps,
			"tomorrow": tomorrowApps,
		},
	})
}

// Test POST /api/v1/reminders/test
func (r *ReminderController) Test(ctx http.Context) http.Response {
	var user models.User
	if err := facades.Auth(ctx).User(&user); err != nil {
		return ctx.Response().Json(401, http.Json{"message": "Unauthorized"})
	}

	emailSvc := services.NewEmailService()
	if err := emailSvc.SendReminderEmail(
		user.Email, user.Name,
		"Frontend Developer", "Tokopedia",
		"day_of",
	); err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Gagal kirim test email", "error": err.Error()})
	}

	return ctx.Response().Json(200, http.Json{"message": "Test email reminder berhasil dikirim."})
}
