package config

import (
	"github.com/goravel/framework/support/carbon"
	"jobbin/backend/app/facades"
)

func Boot() {}

func init() {
	config := facades.Config()
	config.Add("app", map[string]any{
		"name":            config.Env("APP_NAME", "Jobbin"),
		"env":             config.Env("APP_ENV", "production"),
		"debug":           config.Env("APP_DEBUG", false),
		"timezone":        carbon.UTC,
		"locale":          "en",
		"fallback_locale": "en",
		"key":             config.Env("APP_KEY", ""),

		// Frontend URL — dipakai untuk generate link di email
		"frontend_url": config.Env("FRONTEND_URL", "http://localhost:5173"),
	})
}
