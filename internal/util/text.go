package util

import (
	"regexp"
	"strings"
)

func CleanSpaces(s string) string {
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = strings.TrimSpace(s)
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	return s
}

func SplitQueryAndJS(query string) (string, string) {
	parts := strings.SplitN(query, "@js:", 2)
	if len(parts) == 1 {
		return strings.TrimSpace(query), ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

func ApplyInlineJS(jsCode, input string) string {
	if strings.TrimSpace(jsCode) == "" {
		return input
	}
	if out, err := RunInlineJS(jsCode, input); err == nil {
		return strings.TrimSpace(out)
	}
	res := input

	if strings.Contains(jsCode, `r='`) && strings.Contains(jsCode, `'+r`) {
		start := strings.Index(jsCode, `r='`)
		end := strings.Index(jsCode[start+3:], `'`)
		if end > -1 {
			prefix := jsCode[start+3 : start+3+end]
			res = prefix + res
		}
	}

	if strings.Contains(jsCode, "replaceAll(") || strings.Contains(jsCode, "replace(") {
		re := regexp.MustCompile(`replace(All)?\\(\\s*['\"](.*?)['\"]\\s*,\\s*['\"](.*?)['\"]\\s*\\)`)
		for _, m := range re.FindAllStringSubmatch(jsCode, -1) {
			oldV := m[2]
			newV := m[3]
			res = strings.ReplaceAll(res, oldV, newV)
		}
	}

	return strings.TrimSpace(res)
}
