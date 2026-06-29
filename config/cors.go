package config

import (
	"jobbin/backend/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("cors", map[string]any{
		// NOTE: Goravel Gin hardcode Cors() middleware di route.go:46
		// Tidak bisa di-override tanpa patch vendor.
		// Wildcard aman karena semua data endpoint protected by JWT.
		"paths":                []string{"*"},
		"allowed_methods":      []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		"allowed_origins":      []string{"*"},
		"allowed_headers":      []string{"Content-Type", "Authorization", "Accept", "X-Requested-With"},
		"exposed_headers":      []string{},
		"max_age":              86400,
		"supports_credentials": false,
	})
}
