package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"jobbin/backend/app/models"
	"jobbin/backend/app/services"
)

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// GoogleAuth POST /api/v1/auth/google
func (r *AuthController) GoogleAuth(ctx contractshttp.Context) contractshttp.Response {
	auditSvc := services.NewAuditService()

	// Terima access_token dari FE
	accessToken := ctx.Request().Input("credential")
	if accessToken == "" {
		return ctx.Response().Json(422, contractshttp.Json{"message": "Google credential wajib diisi"})
	}

	// Verify access_token via Google userinfo endpoint
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return ctx.Response().Json(500, contractshttp.Json{"message": "Gagal membuat request ke Google"})
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return ctx.Response().Json(401, contractshttp.Json{"message": "Token Google tidak valid"})
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx.Response().Json(500, contractshttp.Json{"message": "Gagal membaca response Google"})
	}

	var googleUser googleUserInfo
	if err := json.Unmarshal(body, &googleUser); err != nil {
		return ctx.Response().Json(500, contractshttp.Json{"message": "Gagal parse response Google"})
	}

	if googleUser.Email == "" {
		return ctx.Response().Json(422, contractshttp.Json{"message": "Email tidak ditemukan dari akun Google"})
	}

	if !googleUser.EmailVerified {
		return ctx.Response().Json(422, contractshttp.Json{"message": "Email Google belum diverifikasi"})
	}

	// Normalisasi email
	email := strings.ToLower(strings.TrimSpace(googleUser.Email))

	// Cek user existing
	var user models.User
	facades.Orm().Query().Where("email", email).First(&user)

	if user.ID != 0 {
		// User sudah ada — merge google_id kalau belum ada
		if user.GoogleID == nil {
			user.GoogleID = &googleUser.Sub
			if user.Avatar == nil && googleUser.Picture != "" {
				user.Avatar = &googleUser.Picture
			}
			facades.Orm().Query().Save(&user)
		}
	} else {
		// User baru — buat akun otomatis
		now := carbon.NewDateTime(carbon.Now())
		user = models.User{
			Name:            googleUser.Name,
			Email:           email,
			GoogleID:        &googleUser.Sub,
			Avatar:          &googleUser.Picture,
			EmailVerifiedAt: now,
		}
		if err := facades.Orm().Query().Create(&user); err != nil {
			return ctx.Response().Json(500, contractshttp.Json{"message": "Gagal membuat akun", "error": err.Error()})
		}
	}

	// Generate JWT token
	token, err := facades.Auth(ctx).Login(&user)
	if err != nil {
		return ctx.Response().Json(500, contractshttp.Json{"message": "Gagal membuat token", "error": err.Error()})
	}

	// Audit log
	auditSvc.Log(ctx, &user.ID, services.ActionLogin, nil)

	return ctx.Response().Json(200, contractshttp.Json{
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
