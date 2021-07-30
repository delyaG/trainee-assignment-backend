package domain

import "fmt"

var (
	// Internal database error
	ErrInternalDatabase = fmt.Errorf("internal database error")

	// User is unauthorized
	ErrUnauthorized = fmt.Errorf("unauthorized")
	// Bad request
	ErrInvalidInputData = fmt.Errorf("invalid input data")
	// Validation Failed
	ErrValidationFailed = fmt.Errorf("validation failed")
	// User already exists
	ErrUserAlreadyExists = fmt.Errorf("user already exists")
	// Invalid registration order
	ErrInvalidRegistrationOrder = fmt.Errorf("invalid registration order")
	// Same email received
	ErrSameEmail = fmt.Errorf("old and new emails are the same")

	// Internal security module error
	ErrInternalSecurity           = fmt.Errorf("internal security module error")

	// Internal OTPStore
	ErrInternalOTPStore         = fmt.Errorf("internal otp store error")
	ErrNonexistentOrExpiredCode = fmt.Errorf("nonexistent or expired code")
	ErrInvalidOTPCode           = fmt.Errorf("invalid otp code")
	ErrOTPSendingExceeded       = fmt.Errorf("otp sending exceeded")
	ErrOTPRateLimitReached      = fmt.Errorf("otp rate limit reached")
	ErrOTPAttemptsExceeded      = fmt.Errorf("otp attempts exceeded")

	ErrNonexistentOrExpiredToken = fmt.Errorf("nonexistent or expired email confirmation token")

	// Internal Email
	ErrInternalEmail         = fmt.Errorf("internal email error")
	ErrEmailAlreadyConfirmed = fmt.Errorf("email is already confirmed")
)

