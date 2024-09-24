package router

import (
	"github.com/go-chi/chi/v5"

	"github.com/ry461ch/loyalty_system/internal/handlers"
	"github.com/ry461ch/loyalty_system/pkg/authentication"
	"github.com/ry461ch/loyalty_system/pkg/authentication/middleware"
	"github.com/ry461ch/loyalty_system/pkg/logging/middleware"
	"github.com/ry461ch/loyalty_system/pkg/middlewares/compressor"
	"github.com/ry461ch/loyalty_system/pkg/middlewares/contenttypes"
)

func NewRouter(
	authHandlers handlers.AuthHandlers,
	moneyHandlers handlers.MoneyHandlers,
	orderHandlers handlers.OrderHandlers,
	authenticator *authentication.Authenticator,
) chi.Router {
	r := chi.NewRouter()

	r.Use(requestlogger.WithLogging)

	r.Route("/api/user", func(r chi.Router) {
		r.Route("/register", func(r chi.Router) {
			r.Use(contenttypes.ValidateJSONContentType)
			r.Post("/", authHandlers.Register)
		})
		r.Route("/login", func(r chi.Router) {
			r.Use(contenttypes.ValidateJSONContentType)
			r.Post("/", authHandlers.Login)
		})
		r.Group(func(r chi.Router) {
			r.Use(authmiddleware.Authenticate(authenticator))
			r.Route("/orders", func(r chi.Router) {
				r.Use(contenttypes.ValidatePlainContentType)
				r.Post("/", orderHandlers.PostOrder)

				r.Group(func(r chi.Router) {
					r.Use(compressor.GzipHandle)
					r.Get("/", orderHandlers.GetOrders)
				})
			})

			r.Route("/withdrawals", func(r chi.Router) {
				r.Use(compressor.GzipHandle, contenttypes.ValidateJSONContentType)
				r.Get("/", moneyHandlers.GetWithdrawals)
			})

			r.Route("/balance", func(r chi.Router) {
				r.Route("/withdraw", func(r chi.Router) {
					r.Use(contenttypes.ValidateJSONContentType)
					r.Post("/", moneyHandlers.PostWithdrawal)
				})

				r.Group(func(r chi.Router) {
					r.Use(contenttypes.ValidatePlainContentType)
					r.Get("/", moneyHandlers.GetBalance)
				})
			})
		})
	})

	return r
}
