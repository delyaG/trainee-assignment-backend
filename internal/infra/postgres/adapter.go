package postgres

import (
	"database/sql"
	"errors"
	"time"
	"trainee-assignment-backend/internal/domain"
	"trainee-assignment-backend/internal/infra/postgres/models"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type adapter struct {
	logger logrus.FieldLogger
	config *Config
	db     *sqlx.DB
}

func NewAdapter(logger logrus.FieldLogger, config *Config) (domain.Database, error) {
	a := &adapter{
		logger: logger,
		config: config,
	}

	db, err := sqlx.Open("pgx", config.ConnectionString())
	if err != nil {
		logger.Errorf("cannot open an sql connection: %w", err)
		return nil, err
	}
	a.db = db

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifeTime)

	// Migrations block
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(config.MigrationsSourceURL, config.Name, driver)
	if err != nil {
		return nil, err
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	}

	return a, nil
}

func (a *adapter) RegisterStart(phone string) (int, error) {
	var id int
	if err := a.db.QueryRowx(
		`INSERT INTO users (phone, status) VALUES ($1, B'00000001') RETURNING id`,
		phone,
	).Scan(&id); err != nil {
		if err, ok := err.(*pgconn.PgError); !ok || err.Code != "23505" {
			a.logger.WithError(err).Error("Error while trying to start a registration!")
			return 0, domain.ErrInternalDatabase
		}
	}

	if err := a.db.QueryRowx(
		`SELECT id FROM users WHERE phone = $1 AND status & B'00000100' != B'00000100'`,
		phone,
	).Scan(&id); err != nil {
		a.logger.WithError(err).Error("Error while trying to get a user_id!")
		if errors.Is(err, sql.ErrNoRows) {
			return 0, domain.ErrUserAlreadyExists
		}

		return 0, domain.ErrInternalDatabase
	}

	return id, nil
}

func (a *adapter) RegisterConfirm(id int) error {
	res, err := a.db.Exec(
		`UPDATE users SET status = status | B'00000010' WHERE id = $1`,
		id,
	)
	if err != nil {
		a.logger.WithError(err).Error("Error while trying to confirm a registration!")
		return domain.ErrInternalDatabase
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		a.logger.WithError(err).Error("Error while trying to confirm a registration!")
		return domain.ErrInternalDatabase
	}

	if rowsAffected != 1 {
		a.logger.WithError(err).Error("Count of affected rows is not equal 1!")
		return domain.ErrInternalDatabase
	}

	return nil
}

func (a *adapter) RegisterFinish(id int, firstName, middleName, lastName, birthday, city string) error {
	res, err := a.db.Exec(
		`UPDATE users
				SET status      = status | B'00000100',
				    first_name  = $2,
				    middle_name = $3,
				    last_name   = $4,
				    birthday    = $5,
				    city        = $6
				WHERE id = $1`,
		id,
		firstName,
		middleName,
		lastName,
		birthday,
		city,
	)
	if err != nil {
		a.logger.WithError(err).Error("Error while trying to finish a registration!")
		return domain.ErrInternalDatabase
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		a.logger.WithError(err).Error("Error while trying to finish a registration!")
		return domain.ErrInternalDatabase
	}

	if rowsAffected != 1 {
		a.logger.WithError(err).Error("Count of affected rows is not equal 1!")
		return domain.ErrInternalDatabase
	}

	return nil
}

func (a *adapter) GetUser(id int) (*domain.User, error) {
	var m models.User
	if err := a.db.Get(
		&m,
		`SELECT id,
				    status,
				    phone,
				    first_name,
				    middle_name,
				    last_name,
				    birthday,
				    city,
				    email,
				    created_at,
				    updated_at
				FROM users
				WHERE id = $1`,
		id,
	); err != nil {
		a.logger.WithError(err).Error("Error while trying to get a user!")
		return nil, domain.ErrInternalDatabase
	}

	return m.Domain(), nil
}

func (a *adapter) UpdateUser(id int, r *domain.ProfileUpdateRequest) (*domain.User, error) {
	var m models.User
	if err := a.db.Get(
		&m,
		`UPDATE users
				SET first_name  = $2,
				    middle_name = $3,
				    last_name   = $4,
				    birthday    = $5,
				    city        = $6
				WHERE id = $1 RETURNING id,
				       status,
				       phone,
				       first_name,
				       middle_name,
				       last_name,
				       birthday,
				       city,
				       email,
				       created_at,
				       updated_at`,
		id,
		r.FirstName,
		r.MiddleName,
		r.LastName,
		r.Birthday,
		r.City,
	); err != nil {
		a.logger.WithError(err).Error("Error while trying to get a user!")
		return nil, domain.ErrInternalDatabase
	}

	return m.Domain(), nil
}

func (a *adapter) GetUserByPhone(phone string) (*domain.User, error) {
	var m models.User
	if err := a.db.Get(
		&m,
		`SELECT id, status, phone, first_name, middle_name, last_name, city, birthday, email,
       					created_at, updated_at FROM users WHERE phone = $1`,
		phone,
	); err != nil {
		a.logger.WithError(err).Error("Error while trying to get a user by phone!")
		return nil, domain.ErrInternalDatabase
	}

	return m.Domain(), nil
}

func (a *adapter) CreateRefreshSession(userID int, fingerprint, userAgent, ip string, expiresAt time.Time) (uuid.UUID, error) {
	var refreshToken uuid.UUID
	if err := a.db.QueryRowx(
		`INSERT INTO refresh_sessions (user_id, fingerprint, user_agent, ip, expires_at)
				VALUES ($1, $2, $3, $4, $5)
				RETURNING refresh_token`,
		userID,
		fingerprint,
		userAgent,
		ip,
		expiresAt,
	).Scan(&refreshToken); err != nil {
		a.logger.WithError(err).Error("Error while trying to create a new refresh session!")
		return uuid.Nil, domain.ErrInternalDatabase
	}

	return refreshToken, nil
}

func (a *adapter) GetRefreshSessionByToken(token string) (*domain.RefreshSession, error) {
	var m models.RefreshSession
	if err := a.db.Get(
		&m,
		`SELECT id,
				       user_id,
				       refresh_token,
				       fingerprint,
				       user_agent,
				       ip,
				       expires_at,
				       created_at
				FROM refresh_sessions
				WHERE refresh_token = $1`,
		token,
	); err != nil {
		a.logger.WithError(err).Error("Error while trying to get a refresh session by token!")
		return nil, domain.ErrInternalDatabase
	}

	return m.Domain(), nil
}

func (a *adapter) RevokeSession(token string) error {
	if _, err := a.db.Exec(
		`UPDATE refresh_sessions SET expires_at = now()
				WHERE refresh_token = $1`,
		token,
	); err != nil {
		a.logger.WithError(err).Error("Error while revoking a session!")
		return domain.ErrInternalDatabase
	}

	return nil
}

func (a *adapter) RevokeObsoleteSessions(userID int) error {
	if _, err := a.db.Exec(
		`UPDATE refresh_sessions SET expires_at = now()
				WHERE user_id = $1 AND id NOT IN (SELECT id FROM refresh_sessions 
													WHERE user_id = $1 AND expires_at > now()
													ORDER BY created_at DESC LIMIT 5)`,
		userID,
	); err != nil {
		a.logger.WithError(err).Error("Error while revoking old sessions!")
		return domain.ErrInternalDatabase
	}

	return nil
}

func (a *adapter) RevokeAllSessions(userID int) error {
	if _, err := a.db.Exec(
		`UPDATE
				refresh_sessions SET expires_at = now()
				WHERE user_id = $1`,
		userID,
	); err != nil {
		a.logger.WithError(err).Error("Error while revoking all sessions!")
		return domain.ErrInternalDatabase
	}

	return nil
}

func (a *adapter) UpdateEmail(userID int, email string) error {
	var currentEmail sql.NullString
	if err := a.db.QueryRow(
		`SELECT email FROM users WHERE id = $1`,
		userID,
	).Scan(&currentEmail); err != nil {
		a.logger.WithError(err).Error("Error while getting an email!")
		return domain.ErrInternalDatabase
	}

	if currentEmail.Valid && currentEmail.String == email {
		return domain.ErrSameEmail
	}

	if _, err := a.db.Exec(
		`UPDATE users SET email = $2, status = status & ~B'00010000' | B'00001000' -- also sets to 0 email confirmation bit
				WHERE id = $1`,
		userID,
		email,
	); err != nil {
		a.logger.WithError(err).Error("Error while updating an email!")
		return domain.ErrInternalDatabase
	}

	return nil
}

func (a *adapter) ConfirmEmail(emailAddress string) error {
	if _, err := a.db.Exec(
		`UPDATE users SET status = status | B'00010000'
				WHERE email = $1`,
		emailAddress,
	); err != nil {
		a.logger.WithError(err).Error("Error while confirming an email!")
		return domain.ErrInternalDatabase
	}

	return nil
}
