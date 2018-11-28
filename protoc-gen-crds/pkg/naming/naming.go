package naming

import "strings"

// PascalCase formatting of a string.
func PascalCase(s string) string {
	if s == "" {
		return s
	}

	parts := strings.Split(s, "_")

	result := ""
	for _, p := range parts {
		if p == "" {
			continue
		}
		prefix := strings.ToUpper(p[:1])
		suffix := p[1:]
		result += prefix + suffix
	}

	return result
}

// CamelCase formatting of a string.
func CamelCase(s string) string {
	if s == "" {
		return s
	}

	parts := strings.Split(s, "_")

	result := ""
	for _, p := range parts {
		if p == "" {
			continue
		}
		prefix := strings.ToUpper(p[:1])
		if result == "" {
			prefix = strings.ToLower(p[:1])
		}
		suffix := p[1:]
		result += prefix + suffix
	}

	return result
}
