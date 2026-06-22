package routes

import (
	"time"

	"jobbin/backend/app/http/controllers"
	"jobbin/backend/app/http/middleware"
	"jobbin/backend/app/facades"
	"github.com/goravel/framework/contracts/route"
)

func Api() {
	authController := controllers.NewAuthController()
	applicationController := controllers.NewApplicationController()
	reminderController := controllers.NewReminderController()
	profileController := controllers.NewProfileController()
	jwtMiddleware := middleware.NewJwtMiddleware()
	loginRateLimit := middleware.NewRateLimitMiddleware(5, time.Minute)

	// Public auth routes
	facades.Route().Prefix("api/v1/auth").Group(func(router route.Router) {
		router.Post("/register", authController.Register)
		router.Post("/verify-email", authController.VerifyEmail)
		router.Post("/resend-verification", authController.ResendVerification)
		router.Middleware(loginRateLimit.Handle()).Post("/login", authController.Login)
	})

	// Protected auth routes
	facades.Route().Prefix("api/v1/auth").Middleware(jwtMiddleware.Handle()).Group(func(router route.Router) {
		router.Get("/me", authController.Me)
		router.Post("/logout", authController.Logout)
	})

	// Applications routes (semua protected)
	facades.Route().Prefix("api/v1").Middleware(jwtMiddleware.Handle()).Group(func(router route.Router) {
		router.Get("/applications", applicationController.Index)
		router.Get("/applications/{id}", applicationController.Show)
		router.Post("/applications", applicationController.Store)
		router.Put("/applications/{id}", applicationController.Update)
		router.Patch("/applications/{id}/position", applicationController.UpdatePosition)
		router.Patch("/applications/{id}/archive", applicationController.ToggleArchive)
		router.Delete("/applications/{id}", applicationController.Destroy)

		// Reminder routes
		router.Get("/reminders", reminderController.Index)
		router.Post("/reminders/test", reminderController.Test)

		// Profile routes
		router.Get("/profile", profileController.Show)
		router.Put("/profile", profileController.UpdateName)
		router.Put("/profile/password", profileController.UpdatePassword)
	})
}
