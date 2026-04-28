// Package validator provides simple value validators for HTTP input.
package validator

import (
	"errors"
	"regexp"
	"strings"
)

// ErrInvalidEmail is returned when an email fails validation.
var ErrInvalidEmail = errors.New("invalid email")

// ErrInvalidTrackingCode is returned when a tracking code fails validation.
var ErrInvalidTrackingCode = errors.New("invalid tracking code")

var (
	emailRE    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	trackingRE = regexp.MustCompile(`^[A-Z]{2}\d{9}[A-Z]{2}$`)
)

// Email validates an email address (basic RFC-ish check).
func Email(s string) error {
	if !emailRE.MatchString(strings.TrimSpace(s)) {
		return ErrInvalidEmail
	}
	return nil
}

// TrackingCode validates a Brazilian tracking code (e.g. AN123456789BR).
// Accepts spaced or unspaced variants — returns the unspaced form.
func TrackingCode(s string) (string, error) {
	clean := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(s), " ", ""))
	if !trackingRE.MatchString(clean) {
		return "", ErrInvalidTrackingCode
	}
	return clean, nil
}
