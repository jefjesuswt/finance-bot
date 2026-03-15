package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jefjesuswt/finance-bot/internal/rates"
)

func main() {
	httpClient := &http.Client{Timeout: 10 * time.Second}

	ratesService := rates.NewService(httpClient)
	ratesHandler := rates.NewHandler(ratesService)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	r.Mount("/api/rates", rates.Routes(ratesHandler))

	log.Println("🚀 Servidor iniciado en el puerto 8081...")
	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}
