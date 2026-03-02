package utils

import (
	"strings"
)

func NormalizeTag(tag string) string {
	normalized := strings.ToLower(tag)
	normalized = strings.ReplaceAll(normalized, " ", "-")
	normalized = strings.ReplaceAll(normalized, "_", "-")

	for strings.Contains(normalized, "--") {
		normalized = strings.ReplaceAll(normalized, "--", "-")
	}
	normalized = strings.Trim(normalized, "-")
	return normalized
}
