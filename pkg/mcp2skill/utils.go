package mcp2skill

import (
	"regexp"
	"strings"
)

var nonAlnumRe = regexp.MustCompile(`[^\p{L}\p{N}-]+`)
var multiDashRe = regexp.MustCompile(`-{2,}`)
var trimDashRe = regexp.MustCompile(`^-+|-+$`)

// toFileName converts a name to a safe filename by replacing non-alphanumeric
// characters with hyphens.
func toFileName(name string) string {
	sanitized := nonAlnumRe.ReplaceAllString(name, "-")
	sanitized = multiDashRe.ReplaceAllString(sanitized, "-")
	sanitized = trimDashRe.ReplaceAllString(sanitized, "")
	if sanitized == "" {
		return "unnamed"
	}
	return strings.ToLower(sanitized)
}
