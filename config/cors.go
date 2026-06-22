package config

import (
	"jobbin/backend/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("cors", map[string]any{
		"paths":                []string{"*"},
		"allowed_methods":      []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		"allowed_origins": []string{
			"http://localhost:5173",
			"https://jobbin-frontend.happygrass-ce5e6fc0.southeastasia.azurecontainerapps.io",
		},
		"allowed_headers":      []string{"Content-Type", "Authorization", "Accept", "X-Requested-With"},
		"exposed_headers":      []string{},
		"max_age":              86400,
		"supports_credentials": false,
	})
}
