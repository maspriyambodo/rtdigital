package auth

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
)

var (
	ErrInvalidEmail = errors.New("invalid email address")
	ErrInvalidPhone = errors.New("invalid phone number")
	e164Phone       = regexp.MustCompile(`^\+[1-9][0-9]{6,14}$`)
)

func NormalizeEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	addr, err := mail.ParseAddress(email)
	if email == "" || err != nil || addr.Address != email {
		return "", ErrInvalidEmail
	}

	at := strings.LastIndexByte(email, '@')
	if at <= 0 || at == len(email)-1 || !strings.Contains(email[at+1:], ".") {
		return "", ErrInvalidEmail
	}
	return email, nil
}

func NormalizePhone(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidPhone
	}

	var digits strings.Builder
	for _, r := range raw {
		switch {
		case r >= '0' && r <= '9':
			digits.WriteRune(r)
		case r == '+' || r == ' ' || r == '-' || r == '(' || r == ')':
		default:
			return "", ErrInvalidPhone
		}
	}

	number := digits.String()
	switch {
	case strings.HasPrefix(raw, "+"):
		number = "+" + number
	case strings.HasPrefix(number, "62"):
		number = "+" + number
	case strings.HasPrefix(number, "0"):
		number = "+62" + number[1:]
	default:
		number = "+62" + number
	}

	if !e164Phone.MatchString(number) {
		return "", ErrInvalidPhone
	}
	return number, nil
}