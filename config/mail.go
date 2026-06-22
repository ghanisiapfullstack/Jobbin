package config

import (
	"jobbin/backend/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("mail", map[string]any{
		"host": config.Env("MAIL_HOST", ""),
		"port": config.Env("MAIL_PORT", 587),
		"from": map[string]any{
			"address": config.Env("MAIL_FROM_ADDRESS", "noreply@jobbin.app"),
			"name":    config.Env("MAIL_FROM_NAME", "Jobbin"),
		},
		"username":       config.Env("MAIL_USERNAME"),
		"password":       config.Env("MAIL_PASSWORD"),
		"resend_api_key": config.Env("RESEND_API_KEY", ""),
		"template": map[string]any{
			"default": config.Env("MAIL_TEMPLATE_ENGINE", "html"),
			"engines": map[string]any{
				"html": map[string]any{
					"driver": "html",
					"path":   config.Env("MAIL_VIEWS_PATH", "resources/views/mail"),
				},
			},
		},
	})
}
