package helpers

import (
	"html"
	"regexp"
	"strings"
)

var (
	// Strip semua HTML tags
	htmlTagRegex = regexp.MustCompile(`<[^>]*>`)
	// Strip script content
	scriptRegex = regexp.MustCompile(`(?i)<script[\s\S]*?</script>`)
	// Strip style content
	styleRegex = regexp.MustCompile(`(?i)<style[\s\S]*?</style>`)
)

// SanitizeString — strip HTML tags dan escape special chars
func SanitizeString(s string) string {
	// Strip script dan style blocks dulu
	s = scriptRegex.ReplaceAllString(s, "")
	s = styleRegex.ReplaceAllString(s, "")
	// Strip semua HTML tags
	s = htmlTagRegex.ReplaceAllString(s, "")
	// Unescape HTML entities dulu, lalu trim
	s = html.UnescapeString(s)
	s = strings.TrimSpace(s)
	return s
}

// SanitizeOptional — sanitize string, return empty string kalau nil
func SanitizeOptional(s string) string {
	if s == "" {
		return ""
	}
	return SanitizeString(s)
}
