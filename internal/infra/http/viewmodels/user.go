package viewmodels

import (
	"time"
	"trainee-assignment-backend/internal/domain"

	validation "github.com/go-ozzo/ozzo-validation/v3"
	"github.com/go-ozzo/ozzo-validation/v3/is"
)

type User struct {
	ID         int        `json:"id"`
	Phone      string     `json:"phone"`
	FirstName  string     `json:"first_name"`
	MiddleName string     `json:"middle_name"`
	LastName   string     `json:"last_name"`
	Birthday   string     `json:"birthday"`
	City       string     `json:"city"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at"`
}

func (m *User) Model(d *domain.User) {
	m.ID = d.ID
	m.Phone = d.Phone
	if d.FirstName != nil {
		m.FirstName = *d.FirstName
	}
	if d.MiddleName != nil {
		m.MiddleName = *d.MiddleName
	}
	if d.LastName != nil {
		m.LastName = *d.LastName
	}
	if d.Birthday != nil {
		m.Birthday = d.Birthday.Format("2006-01-02")
	}
	if d.City != nil {
		m.City = *d.City
	}
	m.CreatedAt = d.CreatedAt
	m.UpdatedAt = d.UpdatedAt
}

type ProfileUpdateRequest struct {
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name"`
	LastName   string `json:"last_name"`
	Birthday   string `json:"birthday"`
	City       string `json:"city"`
}

func (r *ProfileUpdateRequest) Domain() *domain.ProfileUpdateRequest {
	return &domain.ProfileUpdateRequest{
		FirstName:  r.FirstName,
		MiddleName: r.MiddleName,
		LastName:   r.LastName,
		Birthday:   r.Birthday,
		City:       r.City,
	}
}

func (r ProfileUpdateRequest) Validate() error {
	return validation.ValidateStruct(
		&r,
		validation.Field(&r.FirstName, validation.Required),
		validation.Field(&r.MiddleName, validation.Required),
		validation.Field(&r.LastName, validation.Required),
		validation.Field(&r.Birthday, validation.Required, validation.Date("2006-01-02")),
		validation.Field(&r.City, validation.Required),
	)
}

type EmailChangeRequest struct {
	Email string `json:"email"`
}

func (r EmailChangeRequest) Validate() error {
	return validation.ValidateStruct(
		&r,
		validation.Field(&r.Email, validation.Required, is.Email),
	)
}
