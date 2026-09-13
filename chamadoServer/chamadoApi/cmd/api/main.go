package main

import (
	"net/http"
	"os"

	"chamadoApi/internal/handlers"
	"chamadoApi/internal/tools"

	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetReportCaller(true)

	if err := godotenv.Load(); err != nil {
		log.Warn("no .env file found, relying on environment variables")
	}

	r := chi.NewRouter()
	handlers.Handler(r)

	if err := tools.Setup(); err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = ":8083"
	}

	log.Info("Starting Chamado API")
	log.Infof("listening on %s", port)

	if err := http.ListenAndServe(port, r); err != nil {
		log.Error(err)
	}
}
