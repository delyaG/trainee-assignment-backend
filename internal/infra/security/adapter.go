package security

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"strconv"
	"time"
	"trainee-assignment-backend/internal/domain"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"

	"github.com/sirupsen/logrus"
)

type adapter struct {
	logger  *logrus.Logger
	config  *Config
	jwtAuth *jwtauth.JWTAuth
}

func NewAdapter(logger *logrus.Logger, config *Config) (domain.Security, error) {
	a := &adapter{
		logger: logger,
		config: config,
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

	return a, nil
}

func (a *adapter) GetRandomCode(length int) (string, error) {
	randomness := make([]byte, 32)
	if _, err := rand.Read(randomness); err != nil {
		a.logger.WithError(err).Error("Error while generating randomness!")
		return "", domain.ErrInternalSecurity
	}

	dividend := new(big.Int).SetBytes(randomness)
	divider := new(big.Int).SetInt64(int64(math.Pow(10, float64(length))))

	return fmt.Sprintf("%0"+strconv.Itoa(length)+"d", int(new(big.Int).Mod(dividend, divider).Int64())), nil
}

func (a *adapter) GetRandomToken() (string, error) {
	randomness := make([]byte, 32)
	if _, err := rand.Read(randomness); err != nil {
		a.logger.WithError(err).Error("Error while generating randomness!")
		return "", domain.ErrInternalSecurity
	}

	return hex.EncodeToString(randomness), nil
}

func (a *adapter) GetAccessToken(userID int, duration time.Duration) (string, error) {
	now := time.Now().In(time.UTC)
	iss := now.Unix()
	exp := now.Add(duration).Unix()

	_, accessToken, err := a.jwtAuth.Encode(jwt.StandardClaims{
		ExpiresAt: exp,
		IssuedAt:  iss,
		Issuer:    "woman-bank",
		Subject:   strconv.Itoa(userID),
	})
	if err != nil {
		a.logger.WithError(err).Error("Error while encoding access token!")
		return "", domain.ErrInternalSecurity
	}

	return accessToken, nil
}
