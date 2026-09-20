package routes

import (
	"time"

	"jobbin/backend/app/facades"
	"jobbin/backend/app/http/controllers"
	"jobbin/backend/app/http/middleware"

	"github.com/goravel/framework/contracts/route"
)

func Api() {
	authController := controllers.NewAuthController()
	applicationController := controllers.NewApplicationController()
	reminderController := controllers.NewReminderController()
	profileController := controllers.NewProfileController()
	healthController := controllers.NewHealthController()

	jwtMiddleware := middleware.NewJwtMiddleware()
	loginRateLimit := middleware.NewRateLimitMiddleware(5, time.Minute)       // 5 req/menit — login
	publicRateLimit := middleware.NewRateLimitMiddleware(20, time.Minute)     // 20 req/menit — public endpoints
	refreshRateLimit := middleware.NewRateLimitMiddleware(30, time.Minute)    // 30 refresh/menit
	protectedRateLimit := middleware.NewRateLimitMiddleware(100, time.Minute) // 100 req/menit — protected endpoints
	mailRateLimit := middleware.NewRateLimitMiddleware(3, time.Minute)        // protect outbound email quota

	// Lightweight liveness and dependency-aware readiness probes.
	facades.Route().Get("api/v1/health/live", healthController.Live)
	facades.Route().Get("api/v1/health/ready", healthController.Ready)

	// Public auth routes
	facades.Route().Prefix("api/v1/auth").Middleware(publicRateLimit.Handle()).Group(func(router route.Router) {
		router.Post("/register", authController.Register)
		router.Post("/verify-email", authController.VerifyEmail)
		router.Post("/resend-verification", authController.ResendVerification)
		router.Middleware(mailRateLimit.Handle()).Post("/forgot-password", authController.ForgotPassword)
		router.Middleware(loginRateLimit.Handle()).Post("/reset-password", authController.ResetPassword)
		router.Post("/google", authController.GoogleAuth)
		router.Middleware(refreshRateLimit.Handle()).Post("/refresh", authController.Refresh)
		router.Middleware(loginRateLimit.Handle()).Post("/login", authController.Login)
	})

	// Protected auth routes
	facades.Route().Prefix("api/v1/auth").Middleware(jwtMiddleware.Handle(), protectedRateLimit.Handle()).Group(func(router route.Router) {
		router.Get("/me", authController.Me)
		router.Post("/logout", authController.Logout)
	})

	// Applications routes (semua protected)
	facades.Route().Prefix("api/v1").Middleware(jwtMiddleware.Handle(), protectedRateLimit.Handle()).Group(func(router route.Router) {
		router.Get("/applications", applicationController.Index)
		router.Get("/applications/{id}", applicationController.Show)
		router.Post("/applications", applicationController.Store)
		router.Put("/applications/{id}", applicationController.Update)
		router.Patch("/applications/{id}/position", applicationController.UpdatePosition)
		router.Patch("/applications/{id}/archive", applicationController.ToggleArchive)
		router.Delete("/applications/{id}", applicationController.Destroy)

		// Reminder routes
		router.Get("/reminders", reminderController.Index)
		router.Middleware(mailRateLimit.Handle()).Post("/reminders/test", reminderController.Test)

		// Profile routes
		router.Get("/profile", profileController.Show)
		router.Put("/profile", profileController.UpdateName)
		router.Put("/profile/password", profileController.UpdatePassword)
	})
}
