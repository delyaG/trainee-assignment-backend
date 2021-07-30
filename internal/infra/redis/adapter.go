package redis

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"
	"trainee-assignment-backend/internal/domain"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type adapter struct {
	logger *logrus.Logger
	config *Config
	rds    *redis.Client
}

func NewAdapter(logger *logrus.Logger, config *Config) (domain.OTPStore, error) {
	rds := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	if err := rds.Ping().Err(); err != nil {
		logger.WithError(err).Error("Error while trying to ping redis!")
		return nil, err
	}

	return &adapter{
		logger: logger,
		config: config,
		rds:    rds,
	}, nil
}

func (a *adapter) StoreID(requestID uuid.UUID, id int) error {
	if err := a.rds.Del(requestID.String()).Err(); err != nil {
		a.logger.WithError(err).Error("Error while trying to delete an old code!")
		return domain.ErrInternalOTPStore
	}

	if err := a.rds.SetNX(requestID.String(), strconv.Itoa(id), time.Hour).Err(); err != nil {
		a.logger.WithError(err).Error("Error while trying to store OTP!")
		return domain.ErrInternalOTPStore
	}

	return nil
}

func (a *adapter) LoadID(requestID uuid.UUID) (int, error) {
	idStr, err := a.rds.Get(requestID.String()).Result()
	if err != nil {
		a.logger.WithError(err).Error("Error while trying to get an id!")
		return 0, domain.ErrInternalOTPStore
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		a.logger.WithError(err).Error("Error while converting id from str to int!")
		return 0, domain.ErrInternalOTPStore
	}

	return id, nil
}

type otpSendRateLimit struct {
	Phone       string    `json:"phone"`
	Sending     int       `json:"sending"`
	LastSending time.Time `json:"last_sending"`
}

type otpCheckRateLimit struct {
	Code    string `json:"code"`
	Attempt int    `json:"attempt"`
}

func (a *adapter) Store(otpType domain.OTPType, requestID uuid.UUID, phone, code string) error {
	otpSendRateLimitStr, err := a.rds.Get(string(otpType) + ":" + phone).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		a.logger.WithError(err).Error("Error while trying to get number of attempts!")
		return domain.ErrInternalOTPStore
	}

	var sendRateLimit otpSendRateLimit
	if otpSendRateLimitStr != "" {
		_ = json.Unmarshal([]byte(otpSendRateLimitStr), &sendRateLimit)
	}

	// Security checks
	if sendRateLimit.Sending > 5 {
		return domain.ErrOTPSendingExceeded
	}
	if !sendRateLimit.LastSending.IsZero() && time.Now().In(time.UTC).Before(sendRateLimit.LastSending.Add(30*time.Second)) {
		return domain.ErrOTPRateLimitReached
	}

	if err := a.rds.Del(string(otpType) + ":" + phone).Err(); err != nil {
		a.logger.WithError(err).Error("Error while trying to delete an OTP sending rate limit!")
		return domain.ErrInternalOTPStore
	}
	if err := a.rds.Del(string(otpType) + ":" + requestID.String()).Err(); err != nil {
		a.logger.WithError(err).Error("Error while trying to delete an old code!")
		return domain.ErrInternalOTPStore
	}

	sendRateLimit.Phone = phone
	sendRateLimit.Sending++
	sendRateLimit.LastSending = time.Now().In(time.UTC)
	b1, _ := json.Marshal(sendRateLimit)

	if err := a.rds.SetNX(string(otpType)+":"+phone, b1, 1*time.Hour).Err(); err != nil {
		a.logger.WithError(err).Error("Error while trying to store OTP!")
		return domain.ErrInternalOTPStore
	}

	var checkRateLimit otpCheckRateLimit
	checkRateLimit.Code = code
	b2, _ := json.Marshal(checkRateLimit)

	if err := a.rds.SetNX(string(otpType)+":"+requestID.String(), b2, 5*time.Minute).Err(); err != nil {
		a.logger.WithError(err).Error("Error while trying to store OTP!")
		return domain.ErrInternalOTPStore
	}

	return nil
}

func (a *adapter) Verify(otpType domain.OTPType, requestID uuid.UUID, phone, code string) error {
	otpCheckRateLimitStr, err := a.rds.Get(string(otpType) + ":" + requestID.String()).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			a.logger.WithError(err).Error("There was no code sent or it's already expired!")
			return domain.ErrNonexistentOrExpiredCode
		}

		a.logger.WithError(err).Error("Error while trying to get a code!")
		return domain.ErrInternalOTPStore
	}

	var rateLimit otpCheckRateLimit
	_ = json.Unmarshal([]byte(otpCheckRateLimitStr), &rateLimit)

	if rateLimit.Attempt >= 5 {
		if err := a.rds.Del(string(otpType) + ":" + phone).Err(); err != nil {
			a.logger.WithError(err).Error("Error while trying to delete an OTP sending rate limit!")
			return domain.ErrInternalOTPStore
		}
		if err := a.rds.Del(string(otpType) + ":" + requestID.String()).Err(); err != nil {
			a.logger.WithError(err).Error("Error while trying to delete a used code!")
			return domain.ErrInternalOTPStore
		}

		return domain.ErrOTPAttemptsExceeded
	}

	if code != rateLimit.Code && (!a.config.Dev || code != "123456") {
		rateLimit.Attempt++
		b, _ := json.Marshal(rateLimit)

		if err := a.rds.Del(string(otpType) + ":" + requestID.String()).Err(); err != nil {
			a.logger.WithError(err).Error("Error while trying to delete an old code!")
			return domain.ErrInternalOTPStore
		}
		if err := a.rds.SetNX(string(otpType)+":"+requestID.String(), string(b), 5*time.Minute).Err(); err != nil {
			a.logger.WithError(err).Error("Error while trying to store OTP!")
			return domain.ErrInternalOTPStore
		}

		return domain.ErrInvalidOTPCode
	}

	if err := a.rds.Del(string(otpType) + ":" + phone).Err(); err != nil {
		a.logger.WithError(err).Error("Error while trying to delete an OTP sending rate limit!")
		return domain.ErrInternalOTPStore
	}
	if err := a.rds.Del(string(otpType) + ":" + requestID.String()).Err(); err != nil {
		a.logger.WithError(err).Error("Error while trying to delete a used code!")
		return domain.ErrInternalOTPStore
	}

	return nil
}

func (a *adapter) StoreEmail(token, emailAddress string) error {
	if err := a.rds.Del("email:" + token).Err(); err != nil {
		a.logger.WithError(err).Error("Error while trying to delete an old email token!")
		return domain.ErrInternalOTPStore
	}

	if err := a.rds.SetNX("email:"+token, emailAddress, 24*time.Hour).Err(); err != nil {
		a.logger.WithError(err).Error("Error while trying to store email token!")
		return domain.ErrInternalOTPStore
	}

	return nil
}

func (a *adapter) GetEmail(token string) (string, error) {
	storedEmail, err := a.rds.Get("email:" + token).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			a.logger.WithError(err).Error("There was no email sent or it's already expired!")
			return "", domain.ErrNonexistentOrExpiredToken
		}

		a.logger.WithError(err).Error("Error while trying to get an email token!")
		return "", domain.ErrInternalOTPStore
	}

	if err := a.rds.Del("email:" + token).Err(); err != nil {
		a.logger.WithError(err).Error("Error while trying to delete a used token!")
		return "", domain.ErrInternalOTPStore
	}

	return storedEmail, nil
}
