package viewmodels

import (
	validation "github.com/go-ozzo/ozzo-validation/v3"
	"github.com/google/uuid"
	"trainee-assignment-backend/internal/domain"
)

type AuthResponse struct {
	Status       string `json:"status"`
	RequestID    string `json:"request_id,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

func (ar *AuthResponse) Model(d *domain.AuthResponse) {
	ar.Status = d.Status
	if d.RequestID != uuid.Nil {
		ar.RequestID = d.RequestID.String()
	}
	ar.AccessToken = d.AccessToken
	if d.RefreshToken != uuid.Nil {
		ar.RefreshToken = d.RefreshToken.String()
	}
}

type RefreshRequest struct {
	Fingerprint string `json:"fingerprint"`
}

func (rr RefreshRequest) Validate() error {
	return validation.ValidateStruct(
		&rr,
		validation.Field(&rr.Fingerprint, validation.Required),
	)
}

type JWTRequest struct {
	UserID int    `json:"user_id"`
}

func (jr JWTRequest) Validate() error {
	return validation.ValidateStruct(
		&jr,
		validation.Field(&jr.UserID, validation.Required),
	)
}

func (jr JWTRequest) Domain(userAgent, ip string) *domain.JWTRequest {
	return &domain.JWTRequest{
		UserID:    jr.UserID,
		UserAgent: userAgent,
		IP:        ip,
	}
}
