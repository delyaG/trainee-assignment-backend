package viewmodels

import (
	"database/sql"
	"encoding/json"
	"regexp"
	"trainee-assignment-backend/internal/domain"

	"github.com/go-ozzo/ozzo-validation/v3"
	"github.com/go-ozzo/ozzo-validation/v3/is"
	"github.com/google/uuid"
)

type LoginRequest struct {
	Type      string          `json:"type"`
	RequestID string          `json:"request_id"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

func (lr LoginRequest) Validate() error {
	switch lr.Type {
	case "start":
		return validation.ValidateStruct(
			&lr,
			validation.Field(&lr.Type, validation.Required),
			validation.Field(&lr.Payload, validation.By(validateLoginStart)),
		)
	case "resend":
		return validation.ValidateStruct(
			&lr,
			validation.Field(&lr.Type, validation.Required),
			validation.Field(&lr.RequestID, validation.Required, is.UUIDv4),
		)
	case "confirm":
		return validation.ValidateStruct(
			&lr,
			validation.Field(&lr.Type, validation.Required),
			validation.Field(&lr.RequestID, validation.Required, is.UUIDv4),
			validation.Field(&lr.Payload, validation.By(validateLoginConfirm)),
		)
	default:
		return sql.ErrNoRows
	}
}

type LoginRequestStartPayload struct {
	Phone string `json:"phone"`
}

func (p LoginRequestStartPayload) Domain() *domain.LoginRequestStartPayload {
	return &domain.LoginRequestStartPayload{
		Phone: p.Phone,
	}
}

func validateLoginStart(value interface{}) error {
	var p RegistrationRequestStartPayload
	if err := json.Unmarshal(value.(json.RawMessage), &p); err != nil {
		return err
	}

	return validation.ValidateStruct(
		&p,
		validation.Field(
			&p.Phone,
			validation.Required,
			validation.Match(regexp.MustCompile(`9\d{9}`)),
		),
	)
}

type LoginRequestConfirmPayload struct {
	SMSCode     string `json:"sms_code"`
	Fingerprint string `json:"fingerprint"`
}

func (p LoginRequestConfirmPayload) Domain() *domain.LoginRequestConfirmPayload {
	return &domain.LoginRequestConfirmPayload{
		SMSCode:     p.SMSCode,
		Fingerprint: p.Fingerprint,
	}
}

func validateLoginConfirm(value interface{}) error {
	var p LoginRequestConfirmPayload
	if err := json.Unmarshal(value.(json.RawMessage), &p); err != nil {
		return err
	}

	return validation.ValidateStruct(
		&p,
		validation.Field(&p.SMSCode, validation.Required),
		validation.Field(&p.Fingerprint, validation.Required),
	)
}

// Use only after validation
func (lr *LoginRequest) Domain() *domain.LoginRequest {
	d := &domain.LoginRequest{
		Type: domain.LoginRequestType(lr.Type),
	}

	requestID, err := uuid.Parse(lr.RequestID)
	if err == nil {
		d.RequestID = requestID
	}

	switch lr.Type {
	case "start":
		var p LoginRequestStartPayload
		_ = json.Unmarshal(lr.Payload, &p)
		d.Payload = p.Domain()
	case "confirm":
		var p LoginRequestConfirmPayload
		_ = json.Unmarshal(lr.Payload, &p)
		d.Payload = p.Domain()
	}

	return d
}
