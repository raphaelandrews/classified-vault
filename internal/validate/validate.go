package validate

import (
	"fmt"
	"strings"
)

func Username(s string) error {
	s = strings.TrimSpace(s)
	if len(s) < 3 || len(s) > 32 {
		return fmt.Errorf("username must be between 3 and 32 characters")
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-') {
			return fmt.Errorf("username can only contain alphanumeric characters, underscores, and hyphens")
		}
	}
	return nil
}

func Password(s string) error {
	if len(s) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if len(s) > 128 {
		return fmt.Errorf("password must be at most 128 characters")
	}
	return nil
}

func Email(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if len(s) < 3 || len(s) > 254 || !strings.Contains(s, "@") {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func DocumentTitle(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("document title is required")
	}
	if len(s) > 256 {
		return fmt.Errorf("document title must be at most 256 characters")
	}
	return nil
}

func DocumentContent(s string) error {
	if len(s) > 65536 {
		return fmt.Errorf("document content must be at most 64 KB")
	}
	return nil
}
