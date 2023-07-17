package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *AppConfig) initRoutes() http.Handler {

	mux := chi.NewRouter()
	// Attach middlewares
	mux.Use(middleware.Recoverer)
	mux.Use(func(next http.Handler) http.Handler {
		return app.Session.LoadAndSave(next)
	})

	// Define routes
	mux.Get("/", app.getHomePage)
	mux.Get("/login", app.viewLogin)
	mux.Post("/login", app.login)
	mux.Get("/register", app.viewRegister)
	mux.Post("/register", app.postRegister)
	mux.Get("/logout", app.logout)
	mux.Get("/testing", app.testing)
	mux.Get("/activate", app.getActivate)

	mux.Mount("/member", app.authRoutes())

	return mux
}

func (app *AppConfig) authRoutes() http.Handler {
	mux := chi.NewRouter()

	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authenticated := app.Session.Exists(r.Context(), "userID")

			if !authenticated {
				app.Session.Put(r.Context(), "error", "Not authenticated")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	mux.Use(authMiddleware)
	mux.Get("/plans", app.chooseSubscription)
	mux.Get("/subscribe", app.subscribe)

	return mux

}
