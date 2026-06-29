package controllers

import (
	"strconv"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"jobbin/backend/app/helpers"
	"jobbin/backend/app/models"
)

type ApplicationController struct{}

func NewApplicationController() *ApplicationController {
	return &ApplicationController{}
}

// getAuthUserID helper — ambil user ID dari JWT
func getAuthUserID(ctx http.Context) (uint, error) {
	var user models.User
	if err := facades.Auth(ctx).User(&user); err != nil {
		return 0, err
	}
	return user.ID, nil
}

// Index GET /api/v1/applications
func (r *ApplicationController) Index(ctx http.Context) http.Response {
	userID, err := getAuthUserID(ctx)
	if err != nil {
		return ctx.Response().Json(401, http.Json{"message": "Unauthorized"})
	}

	archived := ctx.Request().QueryBool("archived", false)
	status := ctx.Request().Query("status", "")

	query := facades.Orm().Query().
		Where("user_id", userID).
		Where("is_archived", archived).
		Order("position asc")

	if status != "" {
		query = query.Where("status", status)
	}

	var applications []models.Application
	if err := query.Find(&applications); err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Gagal mengambil data", "error": err.Error()})
	}

	return ctx.Response().Json(200, http.Json{
		"message": "Berhasil",
		"data":    applications,
	})
}

// Show GET /api/v1/applications/:id
func (r *ApplicationController) Show(ctx http.Context) http.Response {
	userID, err := getAuthUserID(ctx)
	if err != nil {
		return ctx.Response().Json(401, http.Json{"message": "Unauthorized"})
	}

	id := ctx.Request().RouteInt("id")
	var application models.Application
	if err := facades.Orm().Query().Find(&application, id); err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Terjadi kesalahan", "error": err.Error()})
	}
	if application.ID == 0 {
		return ctx.Response().Json(404, http.Json{"message": "Data tidak ditemukan"})
	}
	if application.UserID != userID {
		return ctx.Response().Json(403, http.Json{"message": "Forbidden"})
	}

	return ctx.Response().Json(200, http.Json{"message": "Berhasil", "data": application})
}

// Store POST /api/v1/applications
func (r *ApplicationController) Store(ctx http.Context) http.Response {
	userID, err := getAuthUserID(ctx)
	if err != nil {
		return ctx.Response().Json(401, http.Json{"message": "Unauthorized"})
	}

	validator, err := facades.Validation().Make(ctx, ctx.Request().All(), map[string]string{
		"job_title": "required|max_len:255",
		"company":   "required|max_len:255",
		"url":       "url",
		"status":    "in:wishlist,applied,interview,offer,rejected",
	})
	if err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Kesalahan validasi", "error": err.Error()})
	}
	if validator.Fails() {
		return ctx.Response().Json(422, http.Json{"message": "Input tidak valid", "errors": validator.Errors().All()})
	}

	// Hitung posisi terakhir di kolom status ini
	status := ctx.Request().Input("status", "wishlist")
	var lastApp models.Application
	facades.Orm().Query().
		Where("user_id", userID).
		Where("status", status).
		Where("is_archived", false).
		Order("position desc").
		First(&lastApp)

	position := lastApp.Position + 1.0

	application := models.Application{
		UserID:   userID,
		JobTitle: helpers.SanitizeString(ctx.Request().Input("job_title")),
		Company:  helpers.SanitizeString(ctx.Request().Input("company")),
		Status:   status,
		Position: position,
	}

	// Optional fields
	if url := ctx.Request().Input("url"); url != "" {
		sanitizedURL := helpers.SanitizeString(url)
		application.URL = &sanitizedURL
	}
	if notes := ctx.Request().Input("notes"); notes != "" {
		sanitizedNotes := helpers.SanitizeString(notes)
		application.Notes = &sanitizedNotes
	}
	if appliedDate := ctx.Request().Input("applied_date"); appliedDate != "" {
		application.AppliedDate = &appliedDate
	}
	if reminderDate := ctx.Request().Input("reminder_date"); reminderDate != "" {
		application.ReminderDate = &reminderDate
	}

	if err := facades.Orm().Query().Create(&application); err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Gagal menyimpan", "error": err.Error()})
	}

	return ctx.Response().Json(201, http.Json{"message": "Lamaran berhasil ditambahkan", "data": application})
}

// Update PUT /api/v1/applications/:id
func (r *ApplicationController) Update(ctx http.Context) http.Response {
	userID, err := getAuthUserID(ctx)
	if err != nil {
		return ctx.Response().Json(401, http.Json{"message": "Unauthorized"})
	}

	id := ctx.Request().RouteInt("id")
	var application models.Application
	if err := facades.Orm().Query().Find(&application, id); err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Terjadi kesalahan", "error": err.Error()})
	}
	if application.ID == 0 {
		return ctx.Response().Json(404, http.Json{"message": "Data tidak ditemukan"})
	}
	if application.UserID != userID {
		return ctx.Response().Json(403, http.Json{"message": "Forbidden"})
	}

	validator, err := facades.Validation().Make(ctx, ctx.Request().All(), map[string]string{
		"job_title": "required|max_len:255",
		"company":   "required|max_len:255",
		"url":       "url",
		"status":    "in:wishlist,applied,interview,offer,rejected",
	})
	if err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Kesalahan validasi", "error": err.Error()})
	}
	if validator.Fails() {
		return ctx.Response().Json(422, http.Json{"message": "Input tidak valid", "errors": validator.Errors().All()})
	}

	// Update fields
	application.JobTitle = helpers.SanitizeString(ctx.Request().Input("job_title", application.JobTitle))
	application.Company = helpers.SanitizeString(ctx.Request().Input("company", application.Company))
	application.Status = ctx.Request().Input("status", application.Status)

	if url := ctx.Request().Input("url"); url != "" {
		sanitizedURL := helpers.SanitizeString(url)
		application.URL = &sanitizedURL
	} else {
		application.URL = nil
	}
	if notes := ctx.Request().Input("notes"); notes != "" {
		sanitizedNotes := helpers.SanitizeString(notes)
		application.Notes = &sanitizedNotes
	} else {
		application.Notes = nil
	}
	if appliedDate := ctx.Request().Input("applied_date"); appliedDate != "" {
		application.AppliedDate = &appliedDate
	} else {
		application.AppliedDate = nil
	}
	if reminderDate := ctx.Request().Input("reminder_date"); reminderDate != "" {
		// Reset reminder flags kalau tanggal berubah
		if application.ReminderDate == nil || *application.ReminderDate != reminderDate {
			application.ReminderSentDayBefore = false
			application.ReminderSentDayOf = false
		}
		application.ReminderDate = &reminderDate
	} else {
		application.ReminderDate = nil
	}

	if err := facades.Orm().Query().Save(&application); err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Gagal mengupdate", "error": err.Error()})
	}

	return ctx.Response().Json(200, http.Json{"message": "Lamaran berhasil diupdate", "data": application})
}

// UpdatePosition PATCH /api/v1/applications/:id/position
func (r *ApplicationController) UpdatePosition(ctx http.Context) http.Response {
	userID, err := getAuthUserID(ctx)
	if err != nil {
		return ctx.Response().Json(401, http.Json{"message": "Unauthorized"})
	}

	id := ctx.Request().RouteInt("id")
	var application models.Application
	if err := facades.Orm().Query().Find(&application, id); err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Terjadi kesalahan", "error": err.Error()})
	}
	if application.ID == 0 {
		return ctx.Response().Json(404, http.Json{"message": "Data tidak ditemukan"})
	}
	if application.UserID != userID {
		return ctx.Response().Json(403, http.Json{"message": "Forbidden"})
	}

	validator, err := facades.Validation().Make(ctx, ctx.Request().All(), map[string]string{
		"position": "required",
		"status":   "in:wishlist,applied,interview,offer,rejected",
	})
	if err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Kesalahan validasi", "error": err.Error()})
	}
	if validator.Fails() {
		return ctx.Response().Json(422, http.Json{"message": "Input tidak valid", "errors": validator.Errors().All()})
	}

	positionStr := ctx.Request().Input("position")
	position, err2 := strconv.ParseFloat(positionStr, 64)
	if err2 != nil {
		return ctx.Response().Json(422, http.Json{"message": "Position harus berupa angka"})
	}
	application.Position = position
	if status := ctx.Request().Input("status"); status != "" {
		application.Status = status
	}

	if err := facades.Orm().Query().Save(&application); err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Gagal mengupdate posisi", "error": err.Error()})
	}

	return ctx.Response().Json(200, http.Json{
		"message": "Posisi diperbarui",
		"data":    map[string]interface{}{"position": application.Position, "status": application.Status},
	})
}

// ToggleArchive PATCH /api/v1/applications/:id/archive
func (r *ApplicationController) ToggleArchive(ctx http.Context) http.Response {
	userID, err := getAuthUserID(ctx)
	if err != nil {
		return ctx.Response().Json(401, http.Json{"message": "Unauthorized"})
	}

	id := ctx.Request().RouteInt("id")
	var application models.Application
	if err := facades.Orm().Query().Find(&application, id); err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Terjadi kesalahan", "error": err.Error()})
	}
	if application.ID == 0 {
		return ctx.Response().Json(404, http.Json{"message": "Data tidak ditemukan"})
	}
	if application.UserID != userID {
		return ctx.Response().Json(403, http.Json{"message": "Forbidden"})
	}

	application.IsArchived = !application.IsArchived
	if err := facades.Orm().Query().Save(&application); err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Gagal mengupdate", "error": err.Error()})
	}

	message := "Lamaran diarsipkan"
	if !application.IsArchived {
		message = "Lamaran dipulihkan dari arsip"
	}

	return ctx.Response().Json(200, http.Json{
		"message": message,
		"data":    map[string]interface{}{"is_archived": application.IsArchived},
	})
}

// Destroy DELETE /api/v1/applications/:id
func (r *ApplicationController) Destroy(ctx http.Context) http.Response {
	userID, err := getAuthUserID(ctx)
	if err != nil {
		return ctx.Response().Json(401, http.Json{"message": "Unauthorized"})
	}

	id := ctx.Request().RouteInt("id")
	var application models.Application
	if err := facades.Orm().Query().Find(&application, id); err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Terjadi kesalahan", "error": err.Error()})
	}
	if application.ID == 0 {
		return ctx.Response().Json(404, http.Json{"message": "Data tidak ditemukan"})
	}
	if application.UserID != userID {
		return ctx.Response().Json(403, http.Json{"message": "Forbidden"})
	}

	if _, err := facades.Orm().Query().Delete(&application); err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Gagal menghapus", "error": err.Error()})
	}

	return ctx.Response().Json(200, http.Json{"message": "Lamaran berhasil dihapus"})
}
