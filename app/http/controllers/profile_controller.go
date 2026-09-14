package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"jobbin/backend/app/models"
	"jobbin/backend/app/services"
)

type ProfileController struct{}

func NewProfileController() *ProfileController {
	return &ProfileController{}
}

// Show GET /api/v1/profile
func (r *ProfileController) Show(ctx http.Context) http.Response {
	var user models.User
	if err := facades.Auth(ctx).User(&user); err != nil {
		return ctx.Response().Json(401, http.Json{"message": "Unauthorized"})
	}

	return ctx.Response().Json(200, http.Json{
		"message": "Berhasil",
		"data": map[string]interface{}{
			"id":                user.ID,
			"name":              user.Name,
			"email":             user.Email,
			"email_verified_at": user.EmailVerifiedAt,
			"created_at":        user.CreatedAt,
		},
	})
}

// UpdateName PUT /api/v1/profile
func (r *ProfileController) UpdateName(ctx http.Context) http.Response {
	var user models.User
	if err := facades.Auth(ctx).User(&user); err != nil {
		return ctx.Response().Json(401, http.Json{"message": "Unauthorized"})
	}

	validator, err := facades.Validation().Make(ctx, ctx.Request().All(), map[string]string{
		"name": "required|min_len:2|max_len:100",
	})
	if err != nil {
		return ctx.Response().Json(422, http.Json{"message": "Input tidak valid", "errors": map[string]string{
			"name": "name wajib diisi",
		}})
	}
	if validator.Fails() {
		return ctx.Response().Json(422, http.Json{"message": "Input tidak valid", "errors": validator.Errors().All()})
	}

	user.Name = ctx.Request().Input("name")
	if err := facades.Orm().Query().Save(&user); err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Gagal menyimpan", "error": err.Error()})
	}

	// Audit log
	services.NewAuditService().Log(ctx, &user.ID, services.ActionUpdateProfile, nil)

	return ctx.Response().Json(200, http.Json{
		"message": "Profil berhasil diupdate",
		"data": map[string]interface{}{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

// UpdatePassword PUT /api/v1/profile/password
func (r *ProfileController) UpdatePassword(ctx http.Context) http.Response {
	var user models.User
	if err := facades.Auth(ctx).User(&user); err != nil {
		return ctx.Response().Json(401, http.Json{"message": "Unauthorized"})
	}

	validator, err := facades.Validation().Make(ctx, ctx.Request().All(), map[string]string{
		"current_password": "required",
		"new_password":     "required|min_len:6",
	})
	if err != nil {
		return ctx.Response().Json(422, http.Json{"message": "Input tidak valid", "errors": map[string]string{
			"current_password": "password lama wajib diisi",
			"new_password":     "password baru wajib diisi",
		}})
	}
	if validator.Fails() {
		return ctx.Response().Json(422, http.Json{"message": "Input tidak valid", "errors": validator.Errors().All()})
	}

	// Verifikasi password lama
	if user.Password == nil || !facades.Hash().Check(ctx.Request().Input("current_password"), *user.Password) {
		return ctx.Response().Json(422, http.Json{
			"message": "Input tidak valid",
			"errors":  map[string]string{"current_password": "Password lama tidak sesuai"},
		})
	}

	// Hash password baru
	hashedPassword, err := facades.Hash().Make(ctx.Request().Input("new_password"))
	if err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Terjadi kesalahan", "error": err.Error()})
	}

	user.Password = &hashedPassword
	if err := facades.Orm().Query().Save(&user); err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Gagal menyimpan", "error": err.Error()})
	}

	// Audit log
	services.NewAuditService().Log(ctx, &user.ID, services.ActionChangePassword, nil)

	return ctx.Response().Json(200, http.Json{"message": "Password berhasil diupdate"})
}
