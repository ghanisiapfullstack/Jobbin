package config

import (
	"fmt"
	"strings"

	"jobbin/backend/app/facades"
)

func allowedOrigins(value any) []string {
	parts := strings.Split(fmt.Sprint(value), ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		if origin := strings.TrimSpace(part); origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}

func init() {
	config := facades.Config()
	config.Add("cors", map[string]any{
		"paths":                []string{"*"},
		"allowed_methods":      []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		"allowed_origins":      allowedOrigins(config.Env("CORS_ALLOWED_ORIGINS", "https://www.jobbin.site,http://localhost:5173")),
		"allowed_headers":      []string{"Content-Type", "Authorization", "Accept", "X-Requested-With"},
		"exposed_headers":      []string{},
		"max_age":              86400,
		"supports_credentials": true,
	})
}
