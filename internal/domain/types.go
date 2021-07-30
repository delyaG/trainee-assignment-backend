package domain

import (
	"github.com/google/uuid"
	"time"
)

type ContextKey string

const (
	ContextUserID       ContextKey = "ctx_user_id"
	ContextRefreshToken ContextKey = "ctx_refresh_token"
)

type RegistrationRequestType string

const (
	RegistrationRequestTypeStart   RegistrationRequestType = "start"
	RegistrationRequestTypeResend  RegistrationRequestType = "resend"
	RegistrationRequestTypeConfirm RegistrationRequestType = "confirm"
	RegistrationRequestTypeFinish  RegistrationRequestType = "finish"
)

type RegistrationRequest struct {
	Type      RegistrationRequestType
	RequestID uuid.UUID
	Payload   interface{}
}

type RegistrationRequestStartPayload struct {
	Phone string
}

type RegistrationRequestConfirmPayload struct {
	SMSCode string
}

type RegistrationRequestFinishPayload struct {
	FirstName  string
	MiddleName string
	LastName   string
	Birthday   string
	City       string

	Fingerprint string
	UserAgent   string
	IP          string
}

type LoginRequestType string

const (
	LoginRequestTypeStart   LoginRequestType = "start"
	LoginRequestTypeResend  LoginRequestType = "resend"
	LoginRequestTypeConfirm LoginRequestType = "confirm"
)

type LoginRequest struct {
	Type      LoginRequestType
	RequestID uuid.UUID
	Payload   interface{}
}

type LoginRequestStartPayload struct {
	Phone string
}

type LoginRequestConfirmPayload struct {
	SMSCode string

	Fingerprint string
	UserAgent   string
	IP          string
}

type AuthResponse struct {
	Status       string
	RequestID    uuid.UUID
	AccessToken  string
	RefreshToken uuid.UUID
}

type OTPType string

const (
	OTPTypeRegistration OTPType = "registration"
	OTPTypeLogin        OTPType = "login"
)

type User struct {
	ID         int
	Status     UserStatus
	Phone      string
	FirstName  *string
	MiddleName *string
	LastName   *string
	Birthday   *time.Time
	City       *string
	Email      *string
	CreatedAt  time.Time
	UpdatedAt  *time.Time
}

type UserStatus int

func (s UserStatus) IsStarted() bool {
	return s&0b00000001 == 0b00000001
}

func (s UserStatus) IsConfirmed() bool {
	return s&0b00000010 == 0b00000010
}

func (s UserStatus) IsFinished() bool {
	return s&0b00000100 == 0b00000100
}

func (s UserStatus) IsEmailReceived() bool {
	return s&0b00001000 == 0b00001000
}

func (s UserStatus) IsEmailConfirmed() bool {
	return s&0b00010000 == 0b00010000
}

type RefreshSession struct {
	ID           int
	UserID       int
	RefreshToken string
	Fingerprint  string
	UserAgent    string
	IP           string
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

type ProfileUpdateRequest struct {
	FirstName  string
	MiddleName string
	LastName   string
	Birthday   string
	City       string
	Email      string
}

type JWTRequest struct {
	UserID      int
	UserAgent   string
	IP          string
	Fingerprint string
}
