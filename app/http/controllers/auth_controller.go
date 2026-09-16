package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"jobbin/backend/app/models"
	"jobbin/backend/app/services"
)

type AuthController struct{}

func NewAuthController() *AuthController {
	return &AuthController{}
}

// Register POST /api/v1/auth/register
func (r *AuthController) Register(ctx http.Context) http.Response {
	validator, err := facades.Validation().Make(ctx, ctx.Request().All(), map[string]string{
		"name":     "required|min_len:2|max_len:100",
		"email":    "required|email",
		"password": "required|min_len:6",
	})
	if err != nil {
		return ctx.Response().Json(422, http.Json{"message": "Input tidak valid", "errors": map[string]string{
			"name":     "name wajib diisi",
			"email":    "email wajib diisi",
			"password": "password wajib diisi",
		}})
	}
	if validator.Fails() {
		return ctx.Response().Json(422, http.Json{"message": "Input tidak valid", "errors": validator.Errors().All()})
	}

	email := strings.ToLower(strings.TrimSpace(ctx.Request().Input("email")))
	name := strings.TrimSpace(ctx.Request().Input("name"))

	// Cek email sudah ada
	var existing models.User
	if err := facades.Orm().Query().Where("email", email).First(&existing); err != nil {
		facades.Log().Errorf("Failed to check existing email: %v", err)
		return ctx.Response().Json(500, http.Json{"message": "Gagal memeriksa email"})
	}
	if existing.ID != 0 {
		return ctx.Response().Json(422, http.Json{
			"message": "Input tidak valid",
			"errors":  map[string]string{"email": "email sudah digunakan"},
		})
	}

	// Hash password
	hashedPassword, err := facades.Hash().Make(ctx.Request().Input("password"))
	if err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Terjadi kesalahan", "error": err.Error()})
	}

	// Generate verify token
	token, err := verificationToken()
	if err != nil {
		facades.Log().Errorf("Failed to generate verification token: %v", err)
		return ctx.Response().Json(500, http.Json{"message": "Gagal membuat token verifikasi"})
	}

	user := models.User{
		Name:     name,
		Email:    email,
		Password: &hashedPassword,
	}
	user.EmailVerifyToken = &token
	user.EmailVerifyExpires = carbon.NewDateTime(carbon.Now().AddHours(24))

	if err := facades.Orm().Query().Create(&user); err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Gagal membuat akun", "error": err.Error()})
	}

	// Kirim email verifikasi via Resend
	emailSvc := services.NewEmailService()
	if err := emailSvc.SendVerificationEmail(user.Email, user.Name, token); err != nil {
		facades.Log().Warningf("Failed to send verification email: %v", err)
	}

	return ctx.Response().Json(201, http.Json{
		"message": "Registrasi berhasil. Cek email untuk verifikasi.",
		"data":    map[string]string{"email": user.Email},
	})
}

// VerifyEmail POST /api/v1/auth/verify-email
func (r *AuthController) VerifyEmail(ctx http.Context) http.Response {
	validator, err := facades.Validation().Make(ctx, ctx.Request().All(), map[string]string{
		"token": "required",
	})
	if err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Kesalahan validasi", "error": err.Error()})
	}
	if validator.Fails() {
		return ctx.Response().Json(422, http.Json{"message": "Input tidak valid", "errors": validator.Errors().All()})
	}

	token := ctx.Request().Input("token")
	var user models.User
	if err := facades.Orm().Query().Where("email_verify_token", token).First(&user); err != nil {
		facades.Log().Errorf("Failed to load verification token: %v", err)
		return ctx.Response().Json(500, http.Json{"message": "Gagal memverifikasi email"})
	}
	if user.ID == 0 {
		return ctx.Response().Json(422, http.Json{"message": "Token tidak valid atau sudah digunakan"})
	}
	if user.EmailVerifyExpires == nil || user.EmailVerifyExpires.IsPast() {
		return ctx.Response().Json(422, http.Json{"message": "Token verifikasi sudah kedaluwarsa. Minta token baru."})
	}

	now := carbon.NewDateTime(carbon.Now())
	user.EmailVerifiedAt = now
	user.EmailVerifyToken = nil
	user.EmailVerifyExpires = nil
	if err := facades.Orm().Query().Save(&user); err != nil {
		facades.Log().Errorf("Failed to save email verification: %v", err)
		return ctx.Response().Json(500, http.Json{"message": "Gagal memverifikasi email"})
	}

	return ctx.Response().Json(200, http.Json{"message": "Email berhasil diverifikasi. Silakan login."})
}

// ResendVerification POST /api/v1/auth/resend-verification
func (r *AuthController) ResendVerification(ctx http.Context) http.Response {
	validator, err := facades.Validation().Make(ctx, ctx.Request().All(), map[string]string{
		"email": "required|email",
	})
	if err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Kesalahan validasi", "error": err.Error()})
	}
	if validator.Fails() {
		return ctx.Response().Json(422, http.Json{"message": "Input tidak valid", "errors": validator.Errors().All()})
	}

	var user models.User
	email := strings.ToLower(strings.TrimSpace(ctx.Request().Input("email")))
	if err := facades.Orm().Query().Where("email", email).First(&user); err != nil {
		facades.Log().Errorf("Failed to load user for verification resend: %v", err)
		return ctx.Response().Json(500, http.Json{"message": "Gagal memproses permintaan"})
	}
	if user.ID == 0 {
		return ctx.Response().Json(200, http.Json{"message": "Email verifikasi telah dikirim ulang."})
	}
	if user.EmailVerifiedAt != nil {
		return ctx.Response().Json(422, http.Json{"message": "Email sudah diverifikasi."})
	}

	token, err := verificationToken()
	if err != nil {
		facades.Log().Errorf("Failed to generate verification token: %v", err)
		return ctx.Response().Json(500, http.Json{"message": "Gagal membuat token verifikasi"})
	}
	user.EmailVerifyToken = &token
	user.EmailVerifyExpires = carbon.NewDateTime(carbon.Now().AddHours(24))
	if err := facades.Orm().Query().Save(&user); err != nil {
		facades.Log().Errorf("Failed to save verification token: %v", err)
		return ctx.Response().Json(500, http.Json{"message": "Gagal menyimpan token verifikasi"})
	}

	// Kirim email via Resend
	emailSvc := services.NewEmailService()
	if err := emailSvc.SendVerificationEmail(user.Email, user.Name, token); err != nil {
		facades.Log().Warningf("Failed to resend verification email: %v", err)
	}

	return ctx.Response().Json(200, http.Json{
		"message": "Email verifikasi telah dikirim ulang.",
	})
}

// Login POST /api/v1/auth/login
func (r *AuthController) Login(ctx http.Context) http.Response {
	auditSvc := services.NewAuditService()

	validator, err := facades.Validation().Make(ctx, ctx.Request().All(), map[string]string{
		"email":    "required|email",
		"password": "required",
	})
	if err != nil {
		return ctx.Response().Json(422, http.Json{"message": "Input tidak valid", "errors": map[string]string{
			"email":    "email wajib diisi",
			"password": "password wajib diisi",
		}})
	}
	if validator.Fails() {
		return ctx.Response().Json(422, http.Json{"message": "Input tidak valid", "errors": validator.Errors().All()})
	}

	var user models.User
	email := strings.ToLower(strings.TrimSpace(ctx.Request().Input("email")))
	if err := facades.Orm().Query().Where("email", email).First(&user); err != nil {
		facades.Log().Errorf("Failed to load user during login: %v", err)
		return ctx.Response().Json(500, http.Json{"message": "Gagal memproses login"})
	}
	if user.ID == 0 {
		return ctx.Response().Json(401, http.Json{"message": "Email atau password salah"})
	}

	if user.Password == nil || !facades.Hash().Check(ctx.Request().Input("password"), *user.Password) {
		return ctx.Response().Json(401, http.Json{"message": "Email atau password salah"})
	}

	if user.EmailVerifiedAt == nil {
		return ctx.Response().Json(403, http.Json{"message": "Verifikasi email dulu sebelum login."})
	}

	token, err := facades.Auth(ctx).Login(&user)
	if err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Gagal membuat token", "error": err.Error()})
	}

	session, refreshToken, err := services.NewSessionService().Create(user.ID, ctx.Request().InputBool("remember_me", false))
	if err != nil {
		facades.Log().Errorf("Failed to create refresh session: %v", err)
		return ctx.Response().Json(500, http.Json{"message": "Gagal membuat sesi login"})
	}

	// Audit log
	auditSvc.Log(ctx, &user.ID, services.ActionLogin, nil)

	return ctx.Response().Cookie(services.RefreshCookie(refreshToken, session)).Json(200, http.Json{
		"message": "Login berhasil",
		"data": map[string]interface{}{
			"token": token,
			"user": map[string]interface{}{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
			},
		},
	})
}

// Refresh POST /api/v1/auth/refresh
func (r *AuthController) Refresh(ctx http.Context) http.Response {
	user, session, refreshToken, err := services.NewSessionService().Rotate(ctx.Request().Cookie(services.RefreshCookieName))
	if err != nil {
		return ctx.Response().Cookie(services.ExpiredRefreshCookie()).Json(401, http.Json{"message": "Sesi login sudah berakhir"})
	}

	token, err := facades.Auth(ctx).Login(&user)
	if err != nil {
		facades.Log().Errorf("Failed to refresh access token: %v", err)
		return ctx.Response().Json(500, http.Json{"message": "Gagal memperbarui sesi"})
	}

	return ctx.Response().Cookie(services.RefreshCookie(refreshToken, session)).Json(200, http.Json{
		"message": "Sesi berhasil diperbarui",
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

// Me GET /api/v1/auth/me
func (r *AuthController) Me(ctx http.Context) http.Response {
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

// Logout POST /api/v1/auth/logout
func (r *AuthController) Logout(ctx http.Context) http.Response {
	auditSvc := services.NewAuditService()

	// Ambil user ID sebelum logout
	var user models.User
	_ = facades.Auth(ctx).User(&user)
	if user.ID != 0 {
		auditSvc.Log(ctx, &user.ID, services.ActionLogout, nil)
	}
	services.NewSessionService().Revoke(ctx.Request().Cookie(services.RefreshCookieName))

	if err := facades.Auth(ctx).Logout(); err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Gagal logout", "error": err.Error()})
	}

	return ctx.Response().Cookie(services.ExpiredRefreshCookie()).Json(200, http.Json{"message": "Logout berhasil"})
}

func verificationToken() (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(tokenBytes), nil
}
