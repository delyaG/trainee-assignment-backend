package models

import (
	"time"
	"trainee-assignment-backend/internal/domain"
)

type RefreshSession struct {
	ID           int       `db:"id"`
	UserID       int       `db:"user_id"`
	RefreshToken string    `db:"refresh_token"`
	Fingerprint  string    `db:"fingerprint"`
	UserAgent    string    `db:"user_agent"`
	IP           string    `db:"ip"`
	ExpiresAt    time.Time `db:"expires_at"`
	CreatedAt    time.Time `db:"created_at"`
}

func (u *RefreshSession) Domain() *domain.RefreshSession {
	return &domain.RefreshSession{
		ID:           u.ID,
		UserID:       u.UserID,
		RefreshToken: u.RefreshToken,
		Fingerprint:  u.Fingerprint,
		UserAgent:    u.UserAgent,
		IP:           u.IP,
		ExpiresAt:    u.ExpiresAt,
		CreatedAt:    u.CreatedAt,
	}
}
