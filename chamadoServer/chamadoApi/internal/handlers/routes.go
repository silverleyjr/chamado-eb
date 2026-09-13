package handlers

import (
	"chamadoApi/internal/middleware"
	"net/http"

	"github.com/go-chi/chi"
	chimiddle "github.com/go-chi/chi/middleware"
)

func Handler(r *chi.Mux) {
	r.Use(chimiddle.StripSlashes)

	r.Route("/chamado", func(router chi.Router) {
		// Opening a ticket, and browsing a section's tickets, are public:
		// clients can report a problem and check on it without logging in.
		router.Post("/", PostChamado)
		router.Get("/public", GetChamadoPublic)
		router.Options("/", AuthorizationOptions)
		router.Options("/public", AuthorizationOptions)

		router.Group(func(protected chi.Router) {
			protected.Use(middleware.Authorization)
			protected.Get("/", GetChamado)
			protected.Put("/", PutChamado)
			protected.Delete("/", DeleteChamado)
		})
	})

	r.Route("/user", func(router chi.Router) {
		router.Use(middleware.Authorization)
		router.Get("/", GetUser)
		router.Post("/", PostUser)
		router.Delete("/", DeleteUser)
		router.Options("/", AuthorizationOptions)
	})

	r.Route("/login", func(router chi.Router) {
		router.Get("/", Login)
		router.Options("/", AuthorizationOptions)
	})
}

func noAcessControl(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

func AuthorizationOptions(w http.ResponseWriter, r *http.Request) {
	noAcessControl(w)
}
