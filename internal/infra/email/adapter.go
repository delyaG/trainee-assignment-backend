package email

import (
	"crypto/tls"
	"trainee-assignment-backend/internal/domain"

	"github.com/matcornic/hermes/v2"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

type adapter struct {
	logger *logrus.Logger
	config *Config
	hermes hermes.Hermes
}

func NewAdapter(logger *logrus.Logger, config *Config) domain.Email {
	a := &adapter{
		logger: logger,
		config: config,
	}

	a.hermes = hermes.Hermes{
		Product: hermes.Product{
			Name:        "Trainee Assignment",
		},
	}

	return a
}

func (a *adapter) SendEmailConfirmation(address, name, token string) error {
	email := hermes.Email{
		Body: hermes.Body{
			Name: name,
			Intros: []string{
				"Добро пожаловать!",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Пожалуйста, подтвердите Ваш email:",
					Button: hermes.Button{
						Color: "#000000",
						Text:  "Подтвердить",
						Link:  a.config.BaseBackendURL + "/v1/profile/email/confirm?token=" + token,
					},
				},
			},
			Outros: []string{
				"Возникли вопросы, нужна помощь? Просто ответьте на это письмо, мы постараемся помочь.",
			},
		},
	}

	emailBody, err := a.hermes.GenerateHTML(email)
	if err != nil {
		a.logger.WithError(err).Error("Error while generating an HTML!")
		return domain.ErrInternalEmail
	}

	m := gomail.NewMessage()
	m.SetAddressHeader("From", a.config.Username, "Trainee Assignment")
	m.SetHeader("To", address)
	m.SetHeader("Subject", "Подтвердите ваш email!")
	m.SetBody("text/html", emailBody)

	dialer := gomail.NewDialer(a.config.Host, a.config.Port, a.config.Username, a.config.Password)
	dialer.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	if err := dialer.DialAndSend(m); err != nil {
		a.logger.WithError(err).Error("Error while sending an email!")
		return domain.ErrInternalEmail
	}

	return nil
}
