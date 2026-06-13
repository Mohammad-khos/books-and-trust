package main

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"

	_ "books-and-trust/services/api-gateway/docs"
	"books-and-trust/shared/tracing"
)

func (app *Application) mount() {
	r := app.Router

	r.Use(tracing.ChiMiddleware(app.Config.App.ServiceName))
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(app.Middleware.MetricsMiddleware)
	r.Use(app.Middleware.RequestLoggerMiddleware)
	r.Use(middleware.StripSlashes)
	//cors (this config uses only for development)
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	})
	r.Use(c.Handler)
	// r.Use(app.Middleware.GlobalRateLimitMiddleware)
	r.Use(app.Middleware.SecurityHeadersMiddleware)

	//swagger
	// docsURL := fmt.Sprintf("%s/swagger/doc.json", app.Config.App.Addr)
	// r.Get("/swagger/*", httpSwagger.Handler(
	// 	httpSwagger.URL(docsURL),
	// ))
	r.Handle("/metrics",promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	//routes
	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Get("/health", app.Handler.HeathCheck)
			r.Route("/users", func(r chi.Router) {
				r.With(app.Middleware.IPRateLimitMiddleware).Post("/register", app.Handler.UserHandler.RegisterUserHandler)
				r.With(app.Middleware.IPRateLimitMiddleware).Post("/login", app.Handler.UserHandler.LoginUserHandler)
				r.Group(func(r chi.Router) {
					r.Use(app.Middleware.AuthMiddleware, app.Middleware.UserRateLimitMiddleware)
					r.Get("/{id}", app.Handler.UserHandler.GetUserByIDHandler)
					r.Patch("/update", app.Handler.UserHandler.UpdateUser)
					r.Delete("/{id}", app.Handler.UserHandler.DeleteUserHandler)
				})
			})
			r.Route("/loans", func(r chi.Router) {
				r.Use(app.Middleware.AuthMiddleware)
				// r.Use(app.Middleware.UserRateLimitMiddleware)
				r.Post("/", app.Handler.LoanHandler.CreateLoanHanlder)
				r.Patch("/", app.Handler.LoanHandler.UpdateLoanHandler)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", app.Handler.LoanHandler.GetLoanByIDHandler)
					r.Post("/delivery", app.Handler.LoanHandler.DeliveryLoanHandler)
					r.Post("/claim", app.Handler.LoanHandler.ClaimLoanHandler)
					r.Post("/confirm-delivery", app.Handler.LoanHandler.ConfirmDeliveryHandler)
				})
				r.Get("/owner/{ownerID}", app.Handler.LoanHandler.ListLoanByOwner)
			})
			r.Route("/admin", func(r chi.Router) {
				r.Route("/users", func(r chi.Router) {
					r.Use(app.Middleware.AuthMiddleware, app.Middleware.AdminsMiddleware)
					r.Post("/ban", app.Handler.LoanHandler.BanUserHandler)

				})
			})
		})
	})
}
