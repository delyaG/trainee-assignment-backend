package models

import (
	"database/sql"
	"time"
	"trainee-assignment-backend/internal/domain"

	"github.com/jackc/pgtype"
)

type User struct {
	ID         int            `db:"id"`
	Status     pgtype.Bit     `db:"status"`
	Phone      string         `db:"phone"`
	FirstName  sql.NullString `db:"first_name"`
	MiddleName sql.NullString `db:"middle_name"`
	LastName   sql.NullString `db:"last_name"`
	Birthday   sql.NullTime   `db:"birthday"`
	City       sql.NullString `db:"city"`
	Email      sql.NullString `db:"email"`
	CreatedAt  time.Time      `db:"created_at"`
	UpdatedAt  sql.NullTime   `db:"updated_at"`
}

func (u *User) Domain() *domain.User {
	d := &domain.User{
		ID:        u.ID,
		Phone:     u.Phone,
		CreatedAt: u.CreatedAt,
	}
	if len(u.Status.Bytes) > 0 {
		d.Status = domain.UserStatus(u.Status.Bytes[0])
	}
	if u.FirstName.Valid {
		d.FirstName = &u.FirstName.String
	}
	if u.MiddleName.Valid {
		d.MiddleName = &u.MiddleName.String
	}
	if u.LastName.Valid {
		d.LastName = &u.LastName.String
	}
	if u.Birthday.Valid {
		d.Birthday = &u.Birthday.Time
	}
	if u.City.Valid {
		d.City = &u.City.String
	}
	if u.Email.Valid {
		d.Email = &u.Email.String
	}
	if u.UpdatedAt.Valid {
		d.UpdatedAt = &u.UpdatedAt.Time
	}

	return d
}
