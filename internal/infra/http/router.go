package http

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
	"github.com/rs/cors"
)

func (a *adapter) newRouter() (http.Handler, error) {
	r := chi.NewRouter()

	// Set default middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	c := cors.New(cors.Options{
		AllowedOrigins:   a.config.AllowedOrigins,
		AllowCredentials: true,
	})
	r.Use(c.Handler)

	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Method(http.MethodPost, "/register", a.wrap(a.register))
			r.Method(http.MethodPost, "/login", a.wrap(a.login))

			r.Method(http.MethodPost, "/jwt", a.wrap(a.getJWT))

			r.Group(func(r chi.Router) {
				r.Use(a.refreshTokenMiddleware)
				r.Method(http.MethodPost, "/refresh", a.wrap(a.refresh))
				r.Method(http.MethodPost, "/logout", a.wrap(a.logout))
			})

			r.Method(http.MethodGet, "/profile/email/confirm", a.wrap(a.confirmEmail))

			r.Group(func(r chi.Router) {
				r.Use(jwtauth.Verifier(a.jwtAuth))
				r.Use(a.accessTokenMiddleware)

				r.Method(http.MethodGet, "/profile", a.wrap(a.getProfile))
				r.Method(http.MethodPatch, "/profile", a.wrap(a.updateProfile))
				r.Method(http.MethodPost, "/profile/email", a.wrap(a.changeEmail))
				r.Method(http.MethodPost, "/profile/email/resend", a.wrap(a.resendConfirmationEmail))
			})
		})
	})

	return r, nil
}
