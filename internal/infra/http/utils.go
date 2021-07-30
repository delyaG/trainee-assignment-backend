package http

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
	"trainee-assignment-backend/internal/domain"
)

func generateFields(r *http.Request) logrus.Fields {
	fields := logrus.Fields{
		"ts":          time.Now().UTC().Format(time.RFC3339),
		"http_proto":  r.Proto,
		"http_method": r.Method,
		"remote_addr": r.RemoteAddr,
		"user_agent":  r.UserAgent(),
		"uri":         r.RequestURI,
	}

	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		fields["req_id"] = reqID
	}

	return fields
}

func j(w http.ResponseWriter, code int, payload interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		return fmt.Errorf("cannot write response: %w", err)
	}

	return nil
}

func jError(w http.ResponseWriter, err error) error {
	code := http.StatusInternalServerError
	localizedError := "Внутренняя ошибка!"

	switch err {
	case domain.ErrInternalDatabase:
		localizedError = "Внутренняя ошибка базы данных!"
	case domain.ErrUnauthorized:
		code = http.StatusUnauthorized
		localizedError = "Вы не авторизованы!"
	case domain.ErrInvalidInputData:
		code = http.StatusBadRequest
		localizedError = "Неверный запрос!"
	case domain.ErrValidationFailed:
		code = http.StatusBadRequest
		localizedError = "Запрос не прошёл валидацию!"
	case domain.ErrUserAlreadyExists:
		code = http.StatusBadRequest
		localizedError = "Пользователь с данным номером телефона уже зарегистрирован!"
	case domain.ErrNonexistentOrExpiredCode:
		code = http.StatusBadRequest
		localizedError = "Данный одноразовый код не существует или его срок действия истёк!"
	case domain.ErrInvalidOTPCode:
		code = http.StatusBadRequest
		localizedError = "Неверный одноразовый код!"
	case domain.ErrOTPSendingExceeded:
		code = http.StatusTooManyRequests
		localizedError = "Лимит на отправку СМС исчерпан! Попробуйте позже в течении дня."
	case domain.ErrOTPRateLimitReached:
		code = http.StatusTooManyRequests
		localizedError = "Отправка СМС временно недоступна! Пожалуйста, подождите."
	case domain.ErrOTPAttemptsExceeded:
		code = http.StatusTooManyRequests
		localizedError = "Лимит на проверку СМС кода исчепан! Попробуйте позже."
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"error":           err.Error(),
		"localized_error": localizedError,
	}); err != nil {
		return fmt.Errorf("cannot write response: %w", err)
	}

	return nil
}
