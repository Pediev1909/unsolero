// Package domain holds the newsletter list's rules: what an acceptable address
// and requesting surface look like, and what the repository is handed when an
// address asks to join. Token material here is always a hash; raw tokens exist
// only in the email that carries them.
package domain

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

// ConsentTextVersion identifies the consent sentence shown beside the form.
// Bump it whenever that frontend copy changes so a stored row proves which
// wording the subscriber agreed to.
const ConsentTextVersion = "2026-09-02"

const maximumEmailLength = 254

type Status string

const (
	StatusPending      Status = "pending"
	StatusConfirmed    Status = "confirmed"
	StatusUnsubscribed Status = "unsubscribed"
)

var (
	ErrInvalidEmail  = errors.New("newsletter email address is invalid")
	ErrInvalidSource = errors.New("newsletter source is invalid")
)

// sourcePattern mirrors the database CHECK constraint. Sources name the
// surface that asked ("footer", "article:<slug>") and never carry free text.
var sourcePattern = regexp.MustCompile(`^[a-z][a-z0-9_.:-]{0,99}$`)

// NormalizeEmail returns the trimmed, lower-cased address or ErrInvalidEmail.
// The address must be a bare mailbox: no display name, one "@", 254 bytes at
// most, and it must survive a strict parse unchanged.
func NormalizeEmail(email string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" || len(normalized) > maximumEmailLength || strings.Count(normalized, "@") != 1 {
		return "", ErrInvalidEmail
	}
	parsed, err := mail.ParseAddress(normalized)
	if err != nil || parsed.Address != normalized {
		return "", ErrInvalidEmail
	}
	return normalized, nil
}

func ValidateSource(source string) error {
	if !sourcePattern.MatchString(source) {
		return ErrInvalidSource
	}
	return nil
}

// PendingSubscription is what the repository stores when an address asks to
// join or re-join the list.
type PendingSubscription struct {
	Email                string
	Source               string
	ConsentTextVersion   string
	ConfirmTokenHash     []byte
	ConfirmExpiresAt     time.Time
	UnsubscribeTokenHash []byte
	RequestedAt          time.Time
}
