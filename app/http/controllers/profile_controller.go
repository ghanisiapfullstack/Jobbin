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
			"avatar":            user.Avatar,
			"has_password":      user.Password != nil,
			"auth_methods":      authMethods(user),
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
		return internalError(ctx, "Gagal menyimpan profil", "PROFILE_UPDATE_FAILED", err)
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
		"new_password":     "required|min_len:8|max_len:72",
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
		return internalError(ctx, "Gagal memproses password baru", "PROFILE_PASSWORD_HASH_FAILED", err)
	}

	user.Password = &hashedPassword
	if err := facades.Orm().Query().Save(&user); err != nil {
		return internalError(ctx, "Gagal menyimpan password baru", "PROFILE_PASSWORD_UPDATE_FAILED", err)
	}

	// Audit log
	services.NewAuditService().Log(ctx, &user.ID, services.ActionChangePassword, nil)
	// Password changes invalidate every remembered device and the current JWT.
	services.NewSessionService().RevokeUser(user.ID)
	if err := facades.Auth(ctx).Logout(); err != nil {
		facades.Log().Warningf("Failed to blacklist access token after password change: %v", err)
	}

	return ctx.Response().Cookie(services.ExpiredRefreshCookie()).Json(200, http.Json{
		"message": "Password berhasil diupdate. Silakan login kembali.",
	})
}

func authMethods(user models.User) []string {
	methods := make([]string, 0, 2)
	if user.Password != nil {
		methods = append(methods, "password")
	}
	if user.GoogleID != nil {
		methods = append(methods, "google")
	}
	return methods
}
