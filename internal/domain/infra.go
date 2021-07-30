package domain

import (
	"context"
	"github.com/google/uuid"
	"time"
)

type Service interface {
	Register(request *RegistrationRequest) (*AuthResponse, error)
	Login(request *LoginRequest) (*AuthResponse, error)
	GetJWT(jwtRequest *JWTRequest) (string, uuid.UUID, error)
	ValidateRefreshToken(token string) (int, error)
	RefreshToken(ctx context.Context, fingerprint, userAgent, ip string) (*AuthResponse, error)
	Logout(ctx context.Context, everywhere bool) error
	GetUser(ctx context.Context) (*User, error)
	UpdateUser(ctx context.Context, r *ProfileUpdateRequest) (*User, error)
	UpdateEmail(ctx context.Context, email string) error
	ResendConfirmationEmail(ctx context.Context) error
	ConfirmEmail(token string) error
}

type Database interface {
	RegisterStart(phone string) (id int, err error)
	RegisterConfirm(id int) error
	RegisterFinish(id int, firstName, middleName, lastName, birthday, city string) error
	GetUser(id int) (*User, error)
	UpdateUser(id int, r *ProfileUpdateRequest) (*User, error)
	GetUserByPhone(phone string) (*User, error)
	CreateRefreshSession(id int, fingerprint, userAgent, ip string, expiresAt time.Time) (uuid.UUID, error)
	GetRefreshSessionByToken(token string) (*RefreshSession, error)
	RevokeSession(token string) error
	RevokeObsoleteSessions(userID int) error
	RevokeAllSessions(userID int) error
	UpdateEmail(userID int, email string) error
	ConfirmEmail(emailAddress string) error
}

type OTPStore interface {
	// OTP
	StoreID(requestID uuid.UUID, id int) error
	LoadID(requestID uuid.UUID) (int, error)

	Store(otpType OTPType, requestID uuid.UUID, phone, code string) error
	Verify(otpType OTPType, requestID uuid.UUID, phone, code string) error

	// Email confirmation
	StoreEmail(token, emailAddress string) error
	GetEmail(token string) (string, error)
}

type Security interface {
	GetRandomCode(length int) (string, error)
	GetAccessToken(userID int, duration time.Duration) (string, error)
	GetRandomToken() (string, error)
}

type Email interface {
	SendEmailConfirmation(address, name, token string) error
}

type SMSSender interface {
	SendSMS(phone, text string) error
}

type Delivery interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}