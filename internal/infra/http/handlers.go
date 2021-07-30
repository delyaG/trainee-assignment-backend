package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"trainee-assignment-backend/internal/domain"
	"trainee-assignment-backend/internal/infra/http/viewmodels"
)

func (a *adapter) wrap(handler func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			a.logger.WithFields(generateFields(r)).WithError(err).Error("Error handling request")
		}
	})
}

func (a *adapter) register(w http.ResponseWriter, r *http.Request) error {
	var registerRequest viewmodels.RegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&registerRequest); err != nil {
		a.logger.WithError(err).Error("Error while decoding request body!")
		return jError(w, domain.ErrInvalidInputData)
	}

	if err := registerRequest.Validate(); err != nil {
		a.logger.WithError(err).Error("Error while validating a register request!")
		return jError(w, domain.ErrValidationFailed)
	}

	d := registerRequest.Domain()
	if d.Type == domain.RegistrationRequestTypeFinish {
		d.Payload.(*domain.RegistrationRequestFinishPayload).UserAgent = r.UserAgent()
		d.Payload.(*domain.RegistrationRequestFinishPayload).IP = r.Header.Get("X-Real-IP")
	}

	resp, err := a.service.Register(d)
	if err != nil {
		return jError(w, err)
	}

	var vm viewmodels.AuthResponse
	vm.Model(resp)

	if d.Type == domain.RegistrationRequestTypeFinish {
		now := time.Now().In(time.UTC)
		maxAge := int(now.Add(60 * 24 * time.Hour).Sub(now).Seconds())

		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    resp.RefreshToken.String(),
			Path:     a.config.CookiePath,
			Domain:   a.config.CookieDomain,
			MaxAge:   maxAge,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
	}

	return j(w, http.StatusOK, vm)
}

func (a *adapter) login(w http.ResponseWriter, r *http.Request) error {
	var loginRequest viewmodels.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		a.logger.WithError(err).Error("Error while decoding request body!")
		return jError(w, domain.ErrInvalidInputData)
	}

	if err := loginRequest.Validate(); err != nil {
		a.logger.WithError(err).Error("Error while validating a login request!")
		return jError(w, domain.ErrValidationFailed)
	}

	d := loginRequest.Domain()
	if d.Type == domain.LoginRequestTypeConfirm {
		d.Payload.(*domain.LoginRequestConfirmPayload).UserAgent = r.UserAgent()
		d.Payload.(*domain.LoginRequestConfirmPayload).IP = r.Header.Get("X-Real-IP")
	}

	resp, err := a.service.Login(d)
	if err != nil {
		return jError(w, err)
	}

	var vm viewmodels.AuthResponse
	vm.Model(resp)

	if d.Type == domain.LoginRequestTypeConfirm {
		now := time.Now().In(time.UTC)
		maxAge := int(now.Add(60 * 24 * time.Hour).Sub(now).Seconds())

		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    resp.RefreshToken.String(),
			Path:     a.config.CookiePath,
			Domain:   a.config.CookieDomain,
			MaxAge:   maxAge,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
	}

	return j(w, http.StatusOK, vm)
}

func (a *adapter) getJWT(w http.ResponseWriter, r *http.Request) error {
	var req viewmodels.JWTRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.logger.WithError(err).Error("Error while decoding request body!")
		return jError(w, domain.ErrInvalidInputData)
	}

	if err := req.Validate(); err != nil {
		a.logger.WithError(err).Error("Error while validating a login request!")
		return jError(w, domain.ErrValidationFailed)
	}

	d := req.Domain(r.UserAgent(), r.Header.Get("X-Real-IP"))

	accessToken, refreshToken, err := a.service.GetJWT(d)
	if err != nil {
		return jError(w, err)
	}

	resp := struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.String(),
	}

	return j(w, http.StatusOK, resp)
}

func (a *adapter) refresh(w http.ResponseWriter, r *http.Request) error {
	var refreshRequest viewmodels.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&refreshRequest); err != nil {
		a.logger.WithError(err).Error("Error while decoding request body!")
		return jError(w, domain.ErrInvalidInputData)
	}

	if err := refreshRequest.Validate(); err != nil {
		a.logger.WithError(err).Error("Error while validating a refresh request!")
		return jError(w, domain.ErrValidationFailed)
	}

	resp, err := a.service.RefreshToken(
		r.Context(),
		refreshRequest.Fingerprint,
		r.UserAgent(),
		r.Header.Get("X-Real-IP"),
	)
	if err != nil {
		a.logger.WithError(err).Error("Error while verifying a refresh token!")
		return jError(w, err)
	}

	var vm viewmodels.AuthResponse
	vm.Model(resp)

	now := time.Now().In(time.UTC)
	maxAge := int(now.Add(60 * 24 * time.Hour).Sub(now).Seconds())

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    resp.RefreshToken.String(),
		Path:     a.config.CookiePath,
		Domain:   a.config.CookieDomain,
		MaxAge:   maxAge,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	return j(w, http.StatusOK, vm)
}

func (a *adapter) logout(w http.ResponseWriter, r *http.Request) error {
	var everywhere bool
	everywhereStr := r.URL.Query().Get("everywhere")
	if everywhereStr != "" {
		f, err := strconv.ParseBool(everywhereStr)
		if err != nil {
			a.logger.WithError(err).Error("Error while validating a logout request!")
			return jError(w, domain.ErrValidationFailed)
		}

		everywhere = f
	}

	if err := a.service.Logout(r.Context(), everywhere); err != nil {
		return jError(w, err)
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (a *adapter) getProfile(w http.ResponseWriter, r *http.Request) error {
	user, err := a.service.GetUser(r.Context())
	if err != nil {
		return jError(w, err)
	}

	var vm viewmodels.User
	vm.Model(user)

	return j(w, http.StatusOK, vm)
}

func (a *adapter) updateProfile(w http.ResponseWriter, r *http.Request) error {
	var profileUpdateRequest viewmodels.ProfileUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&profileUpdateRequest); err != nil {
		a.logger.WithError(err).Error("Error while decoding request body!")
		return jError(w, domain.ErrInvalidInputData)
	}

	if err := profileUpdateRequest.Validate(); err != nil {
		a.logger.WithError(err).Error("Error while validating a profile update request!")
		return jError(w, domain.ErrValidationFailed)
	}

	user, err := a.service.UpdateUser(r.Context(), profileUpdateRequest.Domain())
	if err != nil {
		return jError(w, err)
	}

	var vm viewmodels.User
	vm.Model(user)

	return j(w, http.StatusOK, vm)
}

func (a *adapter) changeEmail(w http.ResponseWriter, r *http.Request) error {
	var emailChangeRequest viewmodels.EmailChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&emailChangeRequest); err != nil {
		a.logger.WithError(err).Error("Error while decoding request body!")
		return jError(w, domain.ErrInvalidInputData)
	}

	if err := emailChangeRequest.Validate(); err != nil {
		a.logger.WithError(err).Error("Error while validating an email change request!")
		return jError(w, domain.ErrValidationFailed)
	}

	if err := a.service.UpdateEmail(r.Context(), emailChangeRequest.Email); err != nil {
		return jError(w, err)
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (a *adapter) resendConfirmationEmail(w http.ResponseWriter, r *http.Request) error {
	if err := a.service.ResendConfirmationEmail(r.Context()); err != nil {
		return jError(w, err)
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (a *adapter) confirmEmail(w http.ResponseWriter, r *http.Request) error {
	token := r.URL.Query().Get("token")
	if token == "" {
		return jError(w, domain.ErrInvalidInputData)
	}

	if err := a.service.ConfirmEmail(token); err != nil {
		return jError(w, err)
	}

	http.Redirect(w, r, a.config.BaseFrontendURL+"/personal/email-confirmation", http.StatusTemporaryRedirect)
	return nil
}
