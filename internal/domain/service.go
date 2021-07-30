package domain

import (
	"context"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"time"
)

type service struct {
	logger   logrus.FieldLogger
	db       Database
	security Security
	otpStore OTPStore
	email    Email
	sms      SMSSender
}

func NewService(logger logrus.FieldLogger, db Database, security Security, otpStore OTPStore, email Email, sms SMSSender) Service {
	s := &service{
		logger:   logger,
		db:       db,
		security: security,
		otpStore: otpStore,
		email:    email,
		sms:      sms,
	}

	return s
}

func (s *service) GetUser(ctx context.Context) (*User, error) {
	userID, ok := ctx.Value(ContextUserID).(int)
	if !ok {
		return nil, ErrInvalidInputData
	}

	return s.db.GetUser(userID)
}

func (s *service) UpdateUser(ctx context.Context, r *ProfileUpdateRequest) (*User, error) {
	userID, ok := ctx.Value(ContextUserID).(int)
	if !ok {
		return nil, ErrInvalidInputData
	}

	return s.db.UpdateUser(userID, r)
}

func (s *service) Register(rr *RegistrationRequest) (*AuthResponse, error) {
	switch rr.Type {
	case RegistrationRequestTypeStart:
		phone := rr.Payload.(*RegistrationRequestStartPayload).Phone

		// Start registration process
		userID, err := s.db.RegisterStart(phone)
		if err != nil {
			return nil, err
		}

		requestID := uuid.New()
		if err := s.otpStore.StoreID(requestID, userID); err != nil {
			return nil, err
		}

		code, err := s.security.GetRandomCode(6)
		if err != nil {
			return nil, err
		}

		if err := s.otpStore.Store(OTPTypeRegistration, requestID, phone, code); err != nil {
			return nil, err
		}

		if err = s.sms.SendSMS(phone, "Ваш код для регистрации в Woman Club: "+code); err != nil {
			return nil, err
		}

		return &AuthResponse{
			Status:    "ok",
			RequestID: requestID,
		}, nil
	case RegistrationRequestTypeResend:
		userID, err := s.otpStore.LoadID(rr.RequestID)
		if err != nil {
			return nil, err
		}

		user, err := s.db.GetUser(userID)
		if err != nil {
			return nil, err
		}

		if user.Status.IsFinished() {
			return nil, ErrUserAlreadyExists
		}

		code, err := s.security.GetRandomCode(6)
		if err != nil {
			return nil, err
		}

		if err := s.otpStore.Store(OTPTypeRegistration, rr.RequestID, user.Phone, code); err != nil {
			return nil, err
		}

		if err = s.sms.SendSMS(user.Phone, "Ваш код для регистрации в Woman Club: "+code); err != nil {
			return nil, err
		}

		return &AuthResponse{
			Status: "ok",
		}, nil
	case RegistrationRequestTypeConfirm:
		userID, err := s.otpStore.LoadID(rr.RequestID)
		if err != nil {
			return nil, err
		}

		user, err := s.db.GetUser(userID)
		if err != nil {
			return nil, err
		}

		if user.Status.IsFinished() {
			return nil, ErrUserAlreadyExists
		}

		if err := s.otpStore.Verify(
			OTPTypeRegistration,
			rr.RequestID,
			user.Phone,
			rr.Payload.(*RegistrationRequestConfirmPayload).SMSCode,
		); err != nil {
			return nil, err
		}

		if err := s.db.RegisterConfirm(userID); err != nil {
			return nil, err
		}

		return &AuthResponse{
			Status: "ok",
		}, nil
	case RegistrationRequestTypeFinish:
		userID, err := s.otpStore.LoadID(rr.RequestID)
		if err != nil {
			return nil, err
		}

		user, err := s.db.GetUser(userID)
		if err != nil {
			return nil, err
		}

		if !user.Status.IsStarted() && !user.Status.IsConfirmed() {
			return nil, ErrInvalidRegistrationOrder
		}

		p := rr.Payload.(*RegistrationRequestFinishPayload)
		if err := s.db.RegisterFinish(
			userID,
			p.FirstName,
			p.MiddleName,
			p.LastName,
			p.Birthday,
			p.City,
		); err != nil {
			return nil, err
		}

		refreshToken, err := s.db.CreateRefreshSession(
			userID,
			p.Fingerprint,
			p.UserAgent,
			p.IP,
			time.Now().In(time.UTC).Add(60*24*time.Hour),
		)
		if err != nil {
			return nil, err
		}

		accessToken, err := s.security.GetAccessToken(userID, 30*time.Minute)
		if err != nil {
			return nil, err
		}

		return &AuthResponse{
			Status:       "ok",
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}, nil
	default:
		return nil, ErrInvalidInputData
	}
}

func (s *service) Login(lr *LoginRequest) (*AuthResponse, error) {
	switch lr.Type {
	case LoginRequestTypeStart:
		phone := lr.Payload.(*LoginRequestStartPayload).Phone

		user, err := s.db.GetUserByPhone(phone)
		if err != nil {
			return nil, err
		}

		requestID := uuid.New()
		if err := s.otpStore.StoreID(requestID, user.ID); err != nil {
			return nil, err
		}

		code, err := s.security.GetRandomCode(6)
		if err != nil {
			return nil, err
		}

		if err := s.otpStore.Store(OTPTypeLogin, requestID, phone, code); err != nil {
			return nil, err
		}

		if err = s.sms.SendSMS(phone, "Ваш код для входа в Woman Club: "+code); err != nil {
			return nil, err
		}

		return &AuthResponse{
			Status:    "ok",
			RequestID: requestID,
		}, nil
	case LoginRequestTypeResend:
		userID, err := s.otpStore.LoadID(lr.RequestID)
		if err != nil {
			return nil, err
		}

		user, err := s.db.GetUser(userID)
		if err != nil {
			return nil, err
		}

		code, err := s.security.GetRandomCode(6)
		if err != nil {
			return nil, err
		}

		if err := s.otpStore.Store(OTPTypeLogin, lr.RequestID, user.Phone, code); err != nil {
			return nil, err
		}

		if err = s.sms.SendSMS(user.Phone, "Ваш код для входа в Woman Club: "+code); err != nil {
			return nil, err
		}

		return &AuthResponse{
			Status: "ok",
		}, nil
	case LoginRequestTypeConfirm:
		userID, err := s.otpStore.LoadID(lr.RequestID)
		if err != nil {
			return nil, err
		}

		user, err := s.db.GetUser(userID)
		if err != nil {
			return nil, err
		}

		p := lr.Payload.(*LoginRequestConfirmPayload)
		if err := s.otpStore.Verify(
			OTPTypeLogin,
			lr.RequestID,
			user.Phone,
			p.SMSCode,
		); err != nil {
			return nil, err
		}

		refreshToken, err := s.db.CreateRefreshSession(
			userID,
			p.Fingerprint,
			p.UserAgent,
			p.IP,
			time.Now().In(time.UTC).Add(60*24*time.Hour),
		)
		if err != nil {
			return nil, err
		}

		accessToken, err := s.security.GetAccessToken(userID, 30*time.Minute)
		if err != nil {
			return nil, err
		}

		return &AuthResponse{
			Status:       "ok",
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}, nil
	default:
		return nil, ErrInvalidInputData
	}
}

func (s *service) GetJWT(jwtRequest *JWTRequest) (string, uuid.UUID, error) {
	accessToken, err := s.security.GetAccessToken(jwtRequest.UserID, 30*time.Minute)
	if err != nil {
		return "", uuid.UUID{}, err
	}

	refreshToken, err := s.db.CreateRefreshSession(
		jwtRequest.UserID,
		jwtRequest.Fingerprint,
		jwtRequest.UserAgent,
		jwtRequest.IP,
		time.Now().In(time.UTC).Add(60*24*time.Hour),
	)
	if err != nil {
		return "", uuid.UUID{}, err
	}

	return accessToken, refreshToken, nil
}

func (s *service) ValidateRefreshToken(token string) (int, error) {
	session, err := s.db.GetRefreshSessionByToken(token)
	if err != nil {
		return 0, err
	}

	if time.Now().After(session.ExpiresAt) {
		return 0, ErrUnauthorized
	}

	return session.UserID, nil
}

func (s *service) RefreshToken(ctx context.Context, fingerprint, userAgent, ip string) (*AuthResponse, error) {
	token, ok := ctx.Value(ContextRefreshToken).(string)
	if !ok {
		return nil, ErrInvalidInputData
	}

	session, err := s.db.GetRefreshSessionByToken(token)
	if err != nil {
		return nil, err
	}

	if time.Now().After(session.ExpiresAt) || fingerprint != session.Fingerprint {
		return nil, ErrUnauthorized
	}

	refreshToken, err := s.db.CreateRefreshSession(
		session.UserID,
		fingerprint,
		userAgent,
		ip,
		time.Now().In(time.UTC).Add(60*24*time.Hour),
	)
	if err != nil {
		return nil, err
	}

	// Revoke current token and obsolete ones
	if err := s.db.RevokeSession(token); err != nil {
		return nil, err
	}
	if err := s.db.RevokeObsoleteSessions(session.UserID); err != nil {
		return nil, err
	}

	accessToken, err := s.security.GetAccessToken(session.UserID, 30*time.Minute)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Status:       "ok",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *service) Logout(ctx context.Context, everywhere bool) error {
	token, ok := ctx.Value(ContextRefreshToken).(string)
	if !ok {
		return ErrInvalidInputData
	}

	if err := s.db.RevokeSession(token); err != nil {
		return err
	}

	if everywhere {
		session, err := s.db.GetRefreshSessionByToken(token)
		if err != nil {
			return err
		}

		return s.db.RevokeAllSessions(session.UserID)
	}

	return nil
}

func (s *service) UpdateEmail(ctx context.Context, emailAddress string) error {
	userID, ok := ctx.Value(ContextUserID).(int)
	if !ok {
		return ErrInvalidInputData
	}

	user, err := s.db.GetUser(userID)
	if err != nil {
		return err
	}

	if user.FirstName == nil {
		return ErrInvalidInputData
	}

	if err := s.db.UpdateEmail(userID, emailAddress); err != nil {
		return err
	}

	token, err := s.security.GetRandomToken()
	if err != nil {
		return err
	}

	if err := s.otpStore.StoreEmail(token, emailAddress); err != nil {
		return err
	}

	if err := s.email.SendEmailConfirmation(emailAddress, *user.FirstName, token); err != nil {
		return err
	}

	return nil
}

func (s *service) ResendConfirmationEmail(ctx context.Context) error {
	userID, ok := ctx.Value(ContextUserID).(int)
	if !ok {
		return ErrInvalidInputData
	}

	user, err := s.db.GetUser(userID)
	if err != nil {
		return err
	}

	if user.FirstName == nil || user.Email == nil {
		return ErrInvalidInputData
	}

	if user.Status.IsEmailConfirmed() {
		return ErrEmailAlreadyConfirmed
	}

	token, err := s.security.GetRandomToken()
	if err != nil {
		return err
	}

	if err := s.otpStore.StoreEmail(token, *user.Email); err != nil {
		return err
	}

	if err := s.email.SendEmailConfirmation(*user.Email, *user.FirstName, token); err != nil {
		return err
	}

	return nil
}

func (s *service) ConfirmEmail(token string) error {
	emailAddress, err := s.otpStore.GetEmail(token)
	if err != nil {
		return err
	}

	if err := s.db.ConfirmEmail(emailAddress); err != nil {
		return err
	}

	return nil
}
