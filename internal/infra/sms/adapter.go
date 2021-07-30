package sms

import (
	"trainee-assignment-backend/internal/domain"

	"github.com/sirupsen/logrus"
)

type adapter struct {
	logger *logrus.Logger
	config *Config
}

func NewAdapter(logger *logrus.Logger, config *Config) domain.SMSSender {
	return &adapter{
		logger: logger,
		config: config,
	}
}

func (a *adapter) SendSMS(phone, text string) error {
	// temp info log with sms text
	a.logger.Info("phone:" + phone + ", text:" + text)
	// add implementation
	return nil
}
