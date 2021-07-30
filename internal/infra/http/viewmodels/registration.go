package viewmodels

import (
	"database/sql"
	"encoding/json"
	"regexp"
	"strings"
	"trainee-assignment-backend/internal/domain"

	"github.com/go-ozzo/ozzo-validation/v3"
	"github.com/go-ozzo/ozzo-validation/v3/is"
	"github.com/google/uuid"
)

type RegistrationRequest struct {
	Type      string          `json:"type"`
	RequestID string          `json:"request_id"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

func (rr RegistrationRequest) Validate() error {
	switch rr.Type {
	case "start":
		return validation.ValidateStruct(
			&rr,
			validation.Field(&rr.Type, validation.Required),
			validation.Field(&rr.Payload, validation.By(validateRegistrationStart)),
		)
	case "resend":
		return validation.ValidateStruct(
			&rr,
			validation.Field(&rr.Type, validation.Required),
			validation.Field(&rr.RequestID, validation.Required, is.UUIDv4),
		)
	case "confirm":
		return validation.ValidateStruct(
			&rr,
			validation.Field(&rr.Type, validation.Required),
			validation.Field(&rr.RequestID, validation.Required, is.UUIDv4),
			validation.Field(&rr.Payload, validation.By(validateRegistrationConfirm)),
		)
	case "finish":
		return validation.ValidateStruct(
			&rr,
			validation.Field(&rr.Type, validation.Required),
			validation.Field(&rr.RequestID, validation.Required, is.UUIDv4),
			validation.Field(&rr.Payload, validation.By(validateRegistrationFinish)),
		)
	default:
		return sql.ErrNoRows
	}
}

type RegistrationRequestStartPayload struct {
	Phone string `json:"phone"`
}

func (p RegistrationRequestStartPayload) Domain() *domain.RegistrationRequestStartPayload {
	return &domain.RegistrationRequestStartPayload{
		Phone: p.Phone,
	}
}

func validateRegistrationStart(value interface{}) error {
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

type RegistrationRequestConfirmPayload struct {
	SMSCode string `json:"sms_code"`
}

func (p RegistrationRequestConfirmPayload) Domain() *domain.RegistrationRequestConfirmPayload {
	return &domain.RegistrationRequestConfirmPayload{
		SMSCode: p.SMSCode,
	}
}

func validateRegistrationConfirm(value interface{}) error {
	var p RegistrationRequestConfirmPayload
	if err := json.Unmarshal(value.(json.RawMessage), &p); err != nil {
		return err
	}

	return validation.ValidateStruct(
		&p,
		validation.Field(&p.SMSCode, validation.Required),
	)
}

type RegistrationRequestFinishPayload struct {
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name"`
	LastName   string `json:"last_name"`
	Birthday   string `json:"birthday"`
	City       string `json:"city"`

	Fingerprint string `json:"fingerprint"`
}

func (p *RegistrationRequestFinishPayload) Domain() *domain.RegistrationRequestFinishPayload {
	return &domain.RegistrationRequestFinishPayload{
		FirstName:   strings.Title(strings.ToLower(p.FirstName)),
		MiddleName:  strings.Title(strings.ToLower(p.MiddleName)),
		LastName:    strings.Title(strings.ToLower(p.LastName)),
		Birthday:    p.Birthday,
		City:        p.City,
		Fingerprint: p.Fingerprint,
	}
}

func validateRegistrationFinish(value interface{}) error {
	var p RegistrationRequestFinishPayload
	if err := json.Unmarshal(value.(json.RawMessage), &p); err != nil {
		return err
	}

	return validation.ValidateStruct(
		&p,
		validation.Field(&p.FirstName, validation.Required),
		validation.Field(&p.MiddleName, validation.Required),
		validation.Field(&p.LastName, validation.Required),
		validation.Field(&p.Birthday, validation.Required, validation.Date("2006-01-02")),
		validation.Field(&p.City, validation.Required),
		validation.Field(&p.Fingerprint, validation.Required),
	)
}

// Use only after validation
func (rr *RegistrationRequest) Domain() *domain.RegistrationRequest {
	d := &domain.RegistrationRequest{
		Type: domain.RegistrationRequestType(rr.Type),
	}

	requestID, err := uuid.Parse(rr.RequestID)
	if err == nil {
		d.RequestID = requestID
	}

	switch rr.Type {
	case "start":
		var p RegistrationRequestStartPayload
		_ = json.Unmarshal(rr.Payload, &p)
		d.Payload = p.Domain()
	case "confirm":
		var p RegistrationRequestConfirmPayload
		_ = json.Unmarshal(rr.Payload, &p)
		d.Payload = p.Domain()
	case "finish":
		var p RegistrationRequestFinishPayload
		_ = json.Unmarshal(rr.Payload, &p)
		d.Payload = p.Domain()
	}

	return d
}
