package controllers

import (
	"crypto/rand"
	"encoding/hex"

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

	// Cek email sudah ada
	var existing models.User
	facades.Orm().Query().Where("email", ctx.Request().Input("email")).First(&existing)
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
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	user := models.User{
		Name:     ctx.Request().Input("name"),
		Email:    ctx.Request().Input("email"),
		Password: &hashedPassword,
	}
	user.EmailVerifyToken = &token

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
	facades.Orm().Query().Where("email_verify_token", token).First(&user)
	if user.ID == 0 {
		return ctx.Response().Json(422, http.Json{"message": "Token tidak valid atau sudah digunakan"})
	}

	now := carbon.NewDateTime(carbon.Now())
	user.EmailVerifiedAt = now
	user.EmailVerifyToken = nil
	user.EmailVerifyExpires = nil
	facades.Orm().Query().Save(&user)

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
	facades.Orm().Query().Where("email", ctx.Request().Input("email")).First(&user)
	if user.ID == 0 {
		return ctx.Response().Json(200, http.Json{"message": "Email verifikasi telah dikirim ulang."})
	}
	if user.EmailVerifiedAt != nil {
		return ctx.Response().Json(422, http.Json{"message": "Email sudah diverifikasi."})
	}

	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)
	user.EmailVerifyToken = &token
	facades.Orm().Query().Save(&user)

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
	facades.Orm().Query().Where("email", ctx.Request().Input("email")).First(&user)
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

	// Audit log
	auditSvc.Log(ctx, &user.ID, services.ActionLogin, nil)

	return ctx.Response().Json(200, http.Json{
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
	facades.Auth(ctx).User(&user)
	if user.ID != 0 {
		auditSvc.Log(ctx, &user.ID, services.ActionLogout, nil)
	}

	if err := facades.Auth(ctx).Logout(); err != nil {
		return ctx.Response().Json(500, http.Json{"message": "Gagal logout", "error": err.Error()})
	}

	return ctx.Response().Json(200, http.Json{"message": "Logout berhasil"})
}
