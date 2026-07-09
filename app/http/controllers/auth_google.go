package controllers

import (
	"context"
	"strings"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"google.golang.org/api/idtoken"
	"jobbin/backend/app/models"
	"jobbin/backend/app/services"
)

// GoogleAuth POST /api/v1/auth/google
func (r *AuthController) GoogleAuth(ctx http.Context) http.Response {
	auditSvc := services.NewAuditService()

	// Validasi input
	credential := ctx.Request().Input("credential")
	if credential == "" {
		return ctx.Response().Json(422, http.Json{"message": "Google credential wajib diisi"})
	}

	// Verify Google token
	clientID := facades.Config().Env("GOOGLE_CLIENT_ID", "").(string)
	payload, err := idtoken.Validate(context.Background(), credential, clientID)
	if err != nil {
		return ctx.Response().Json(401, http.Json{"message": "Token Google tidak valid", "error": err.Error()})
	}

	// Ambil data dari Google payload
	googleID := payload.Subject
	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	avatar, _ := payload.Claims["picture"].(string)
	emailVerified, _ := payload.Claims["email_verified"].(bool)

	if email == "" {
		return ctx.Response().Json(422, http.Json{"message": "Email tidak ditemukan dari akun Google"})
	}

	if !emailVerified {
		return ctx.Response().Json(422, http.Json{"message": "Email Google belum diverifikasi"})
	}

	// Normalisasi email
	email = strings.ToLower(strings.TrimSpace(email))

	// Cek user existing
	var user models.User
	facades.Orm().Query().Where("email", email).First(&user)

	if user.ID != 0 {
		// User sudah ada — merge google_id kalau belum ada
		if user.GoogleID == nil {
			user.GoogleID = &googleID
			if user.Avatar == nil && avatar != "" {
				user.Avatar = &avatar
			}
			facades.Orm().Query().Save(&user)
		}
	} else {
		// User baru — buat akun otomatis
		now := carbon.NewDateTime(carbon.Now())
		user = models.User{
			Name:            name,
			Email:           email,
			GoogleID:        &googleID,
			Avatar:          &avatar,
			EmailVerifiedAt: now,
		}
		if err := facades.Orm().Query().Create(&user); err != nil {
			return ctx.Response().Json(500, http.Json{"message": "Gagal membuat akun", "error": err.Error()})
		}
	}

	// Generate JWT token
	token, err := facades.Auth(ctx).Login(&user)
	if err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Gagal membuat token", "error": err.Error()})
	}

	// Audit log
	auditSvc.Log(ctx, &user.ID, services.ActionLogin, nil)

	return ctx.Response().Json(200, http.Json{
		"message": "Login berhasil",
		"data": map[string]interface{}{
			"token": token,
			"user": map[string]interface{}{
				"id":     user.ID,
				"name":   user.Name,
				"email":  user.Email,
				"avatar": user.Avatar,
			},
		},
	})
}
