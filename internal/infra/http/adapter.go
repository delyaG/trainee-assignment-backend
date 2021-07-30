package http

import (
	"context"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"trainee-assignment-backend/internal/domain"
)

type adapter struct {
	logger  *logrus.Logger
	config  *Config
	service domain.Service

	server *http.Server

	// jwt
	jwtAuth *jwtauth.JWTAuth
}

// Creating a new HTTP adapter.
func NewAdapter(logger *logrus.Logger, config *Config, service domain.Service) (domain.Delivery, error) {
	a := &adapter{
		logger:  logger,
		config:  config,
		service: service,
	}

	// Read JWT signing key
	fileWithAccessToken, err := os.Open(a.config.JWTPrivateKey)
	if err != nil {
		a.logger.WithError(err).Error("Error while opening file!")
		return nil, err
	}

	//noinspection ALL
	defer fileWithAccessToken.Close()

	bytes, err := ioutil.ReadAll(fileWithAccessToken)
	if err != nil {
		a.logger.WithError(err).Error("Error while reading file!")
		return nil, err
	}

	jwtAuth := jwtauth.New(jwt.SigningMethodHS256.Name, bytes, nil)
	a.jwtAuth = jwtAuth

	r, err := a.newRouter()
	if err != nil {
		logger.WithError(err).Error("Error while creating new router!")
		return nil, err
	}

	a.server = &http.Server{
		Addr:    config.Address,
		Handler: r,
	}

	return a, nil
}

// ListenAndServe HTTP requests.
func (a *adapter) ListenAndServe() error {
	a.logger.WithField("address", a.config.Address).Info("Listening and serving HTTP requests.")

	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		a.logger.WithError(err).Error("Error listening and serving HTTP requests!")
		return err
	}

	return nil
}

// Shutdown the HTTP adapter
func (a *adapter) Shutdown(ctx context.Context) error {
	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.WithError(err).Error("Error shutting down HTTP adapter!")
		return err
	}

	return nil
}
