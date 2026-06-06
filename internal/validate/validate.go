package validate

import (
	"fmt"
	"strings"
	"unicode"
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

var commonPasswords = map[string]bool{
	"password": true, "123456": true, "12345678": true, "qwerty": true, "abc123": true,
	"monkey": true, "1234567": true, "letmein": true, "trustno1": true, "dragon": true,
	"baseball": true, "iloveyou": true, "master": true, "sunshine": true, "ashley": true,
	"bailey": true, "shadow": true, "123123": true, "654321": true, "superman": true,
	"qazwsx": true, "michael": true, "football": true, "password1": true, "123456789": true,
	"1234567890": true, "qwerty123": true, "1q2w3e4r": true, "123qwe": true, "login": true,
	"princess": true, "admin": true, "welcome": true, "1234": true, "password123": true,
	"passw0rd": true, "111111": true, "000000": true, "letmein1": true, "987654321": true,
	"jesus": true, "mustang": true, "access": true, "shadow1": true, "michael1": true,
	"whatever": true, "qwertyuiop": true, "harley": true, "hello": true, "access01": true,
	"starwars": true, "pepper": true, "123abc": true, "hello1": true, "qwerty1": true,
	"zxcvbnm": true, "123456a": true, "password12": true, "asdfghjkl": true, "flower": true,
	"batman": true, "hockey": true, "sunshine1": true, "iloveyou1": true, "ginger": true,
	"zxcvbn": true, "qwert": true, "tigger": true, "1q2w3e4r5t": true, "pass": true,
	"liverpool": true, "cheese": true, "summer": true, "charlie": true, "michelle": true,
	"andrew": true, "ashley1": true, "jessica": true, "joshua": true, "pepper1": true,
	"matthew": true, "hannah": true, "samantha": true, "password2": true, "12345678910": true,
	"1qaz2wsx": true, "pa$$word": true, "pass123": true, "charlie1": true, "computer": true,
	"daniel": true, "taylor": true, "freedom": true, "passwrod": true, "george": true,
	"thomas": true, "andrea": true, "asshole": true, "password3": true, "pepper12": true,
}

func Password(s string) error {
	if len(s) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if len(s) > 128 {
		return fmt.Errorf("password must be at most 128 characters")
	}

	if commonPasswords[strings.ToLower(s)] {
		return fmt.Errorf("password is too common")
	}

	hasLetter := false
	hasDigitOrSymbol := false
	for _, c := range s {
		if unicode.IsLetter(c) {
			hasLetter = true
		}
		if !unicode.IsLetter(c) {
			hasDigitOrSymbol = true
		}
		if hasLetter && hasDigitOrSymbol {
			break
		}
	}
	if !hasLetter {
		return fmt.Errorf("password must contain at least one letter")
	}
	if !hasDigitOrSymbol {
		return fmt.Errorf("password must contain at least one digit or symbol")
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
