package main

import (
	"context"
	"fmt"
	"github.com/jessevdk/go-flags"
	"os"
	"os/signal"
	"syscall"
	"time"
	"trainee-assignment-backend/internal/configs"
	"trainee-assignment-backend/internal/domain"
	"trainee-assignment-backend/internal/infra/email"
	"trainee-assignment-backend/internal/infra/http"
	"trainee-assignment-backend/internal/infra/postgres"
	"trainee-assignment-backend/internal/infra/redis"
	"trainee-assignment-backend/internal/infra/security"
	"trainee-assignment-backend/internal/infra/sms"
	"trainee-assignment-backend/pkg/logging"
)

func main() {
	config, err := configs.Parse()
	if err != nil {
		if err, ok := err.(*flags.Error); ok {
			fmt.Println(err)
			os.Exit(0)
		}

		fmt.Printf("Invalid args: %v\n", err)
		os.Exit(1)
	}

	// Init logger
	logger, err := logging.NewLogger(config.Logger)
	if err != nil {
		panic(err)
	}

	// Init PostgreSQL
	db, err := postgres.NewAdapter(logger, config.Postgres)
	if err != nil {
		logger.WithError(err).Fatal("Error while creating a new database adapter!")
	}

	// Init Security
	sec, err := security.NewAdapter(logger, config.Security)
	if err != nil {
		logger.WithError(err).Fatal("Error while creating a new security adapter!")
	}

	// Init SMS
	s := sms.NewAdapter(logger, config.SMS)

	// Init OTPStore
	otpStore, err := redis.NewAdapter(logger, config.Redis)
	if err != nil {
		logger.WithError(err).Fatal("Error while creating a new OTPStore adapter")
	}

	// Init Email adapter
	e := email.NewAdapter(logger, config.Email)

	// Init service
	service := domain.NewService(logger, db, sec, otpStore, e, s)

	// Init HTTP adapter
	httpAdapter, err := http.NewAdapter(logger, config.HTTP, service)
	if err != nil {
		logger.WithError(err).Fatal("Error creating new HTTP adapter!")
	}

	shutdown := make(chan error, 1)

	go func(shutdown chan<- error) {
		shutdown <- httpAdapter.ListenAndServe()
	}(shutdown)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case s := <-sig:
		logger.WithField("signal", s).Info("Got the signal!")
	case err := <-shutdown:
		logger.WithError(err).Error("Error running the application!")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	logger.Info("Stopping application...")

	if err := httpAdapter.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Error shutting down the HTTP server!")
	}

	time.Sleep(time.Second)

	logger.Info("The application stopped.")
}
