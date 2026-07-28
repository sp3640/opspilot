package utils

import (
	"regexp"
	"strings"
)

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

func GenerateSlug(name string) string {
	slug := strings.ToLower(strings.TrimSpace(name))
	slug = nonAlnum.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}
