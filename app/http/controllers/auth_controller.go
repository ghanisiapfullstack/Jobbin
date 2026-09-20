package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
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
		"password": "required|min_len:8|max_len:72",
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
		return internalError(ctx, "Gagal memproses registrasi", "AUTH_PASSWORD_HASH_FAILED", err)
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
		return internalError(ctx, "Gagal membuat akun", "AUTH_REGISTER_FAILED", err)
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
		return internalError(ctx, "Gagal memvalidasi permintaan", "AUTH_VALIDATION_FAILED", err)
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
		return internalError(ctx, "Gagal memvalidasi permintaan", "AUTH_VALIDATION_FAILED", err)
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
		return internalError(ctx, "Gagal membuat sesi login", "AUTH_TOKEN_FAILED", err)
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
		return internalError(ctx, "Gagal logout", "AUTH_LOGOUT_FAILED", err)
	}

	return ctx.Response().Cookie(services.ExpiredRefreshCookie()).Json(200, http.Json{"message": "Logout berhasil"})
}

// ForgotPassword POST /api/v1/auth/forgot-password
// The response is intentionally generic to prevent account enumeration.
func (r *AuthController) ForgotPassword(ctx http.Context) http.Response {
	validator, err := facades.Validation().Make(ctx, ctx.Request().All(), map[string]string{"email": "required|email"})
	if err != nil {
		return internalError(ctx, "Gagal memproses permintaan", "PASSWORD_RESET_VALIDATION_FAILED", err)
	}
	if validator.Fails() {
		return ctx.Response().Json(422, http.Json{"message": "Input tidak valid", "errors": validator.Errors().All()})
	}

	message := "Jika email tersebut terdaftar, kami telah mengirimkan tautan reset password."
	email := strings.ToLower(strings.TrimSpace(ctx.Request().Input("email")))
	var user models.User
	if err := facades.Orm().Query().Where("email", email).First(&user); err != nil {
		facades.Log().Errorf("PASSWORD_RESET_LOOKUP_FAILED: %v", err)
		return ctx.Response().Json(200, http.Json{"message": message})
	}
	if user.ID == 0 || user.EmailVerifiedAt == nil {
		return ctx.Response().Json(200, http.Json{"message": message})
	}

	token, err := services.NewPasswordResetService().Create(user.ID)
	if err != nil {
		facades.Log().Errorf("PASSWORD_RESET_CREATE_FAILED: %v", err)
		return ctx.Response().Json(200, http.Json{"message": message})
	}
	if err := services.NewEmailService().SendPasswordResetEmail(user.Email, user.Name, token); err != nil {
		facades.Log().Warningf("PASSWORD_RESET_EMAIL_FAILED: %v", err)
	}
	return ctx.Response().Json(200, http.Json{"message": message})
}

// ResetPassword POST /api/v1/auth/reset-password
func (r *AuthController) ResetPassword(ctx http.Context) http.Response {
	validator, err := facades.Validation().Make(ctx, ctx.Request().All(), map[string]string{
		"token":    "required",
		"password": "required|min_len:8|max_len:72",
	})
	if err != nil {
		return internalError(ctx, "Gagal memproses permintaan", "PASSWORD_RESET_VALIDATION_FAILED", err)
	}
	if validator.Fails() {
		return ctx.Response().Json(422, http.Json{"message": "Input tidak valid", "errors": validator.Errors().All()})
	}

	hashedPassword, err := facades.Hash().Make(ctx.Request().Input("password"))
	if err != nil {
		return internalError(ctx, "Gagal menyimpan password baru", "PASSWORD_RESET_HASH_FAILED", err)
	}
	user, err := services.NewPasswordResetService().Consume(ctx.Request().Input("token"), hashedPassword)
	if errors.Is(err, services.ErrInvalidPasswordResetToken) {
		return ctx.Response().Json(422, http.Json{"message": "Tautan reset tidak valid, sudah digunakan, atau kedaluwarsa"})
	}
	if err != nil {
		return internalError(ctx, "Gagal menyimpan password baru", "PASSWORD_RESET_FAILED", err)
	}

	services.NewSessionService().RevokeUser(user.ID)
	return ctx.Response().Cookie(services.ExpiredRefreshCookie()).Json(200, http.Json{
		"message": "Password berhasil diubah. Silakan login kembali.",
	})
}

func verificationToken() (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(tokenBytes), nil
}
