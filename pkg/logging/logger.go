package logging

import (
	"github.com/sirupsen/logrus"
)

//NewLogger creates a new logger.
func NewLogger(config *Config) (*logrus.Logger, error) {
	logger := logrus.New()

	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		return nil, err
	}
	logger.SetLevel(level)

	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	return logger, nil
}
