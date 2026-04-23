package util

import (
	"regexp"
	"strings"
)

var invalidName = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)

func SanitizeName(s string) string {
	s = invalidName.ReplaceAllString(s, "_")
	s = strings.TrimSpace(s)
	if s == "" {
		return "untitled"
	}
	return s
}
