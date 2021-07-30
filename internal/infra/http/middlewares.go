package http

import (
	"context"
	"net/http"
	"strconv"
	"trainee-assignment-backend/internal/domain"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"
)

func (a *adapter) accessTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := jwtauth.VerifyRequest(a.jwtAuth, r, jwtauth.TokenFromHeader)
		if err != nil {
			a.logger.WithError(err).Error("Error while verifying an access token!")

			w.Header().Add("WWW-Authenticate", "Bearer")
			_ = jError(w, domain.ErrUnauthorized)
			return
		}

		sub, ok := token.Claims.(jwt.MapClaims)["sub"]
		if !ok {
			a.logger.Error("Token is without 'sub' field!")
			_ = jError(w, domain.ErrUnauthorized)
			return
		}

		subStr, ok := sub.(string)
		if !ok {
			a.logger.Error("'sub' field is not a string!")
			_ = jError(w, domain.ErrUnauthorized)
			return
		}

		id, err := strconv.Atoi(subStr)
		if err != nil {
			a.logger.WithError(err).Error("Error while converting string into int!")
			_ = jError(w, domain.ErrUnauthorized)
			return
		}

		w.Header().Set("User-ID", strconv.Itoa(id))
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), domain.ContextUserID, id)))
	})
}

func (a *adapter) refreshTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirect := false
		if r.URL.Query().Get("sso") != "" && r.URL.Query().Get("sig") != "" {
			redirect = true
		}

		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			a.logger.WithError(err).Error("Error while reading a cookie!")
			if !redirect {
				_ = jError(w, domain.ErrUnauthorized)
			} else {
				http.Redirect(w, r, a.config.BaseFrontendURL+"/auth?"+r.URL.RawQuery, http.StatusTemporaryRedirect)
			}

			return
		}

		userID, err := a.service.ValidateRefreshToken(cookie.Value)
		if err != nil {
			if !redirect {
				_ = jError(w, err)
			} else {
				http.Redirect(w, r, a.config.BaseFrontendURL+"/auth?"+r.URL.RawQuery, http.StatusTemporaryRedirect)
			}

			return
		}

		ctx := context.WithValue(r.Context(), domain.ContextUserID, userID)
		ctx = context.WithValue(ctx, domain.ContextRefreshToken, cookie.Value)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
