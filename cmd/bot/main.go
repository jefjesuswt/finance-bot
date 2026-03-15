package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jefjesuswt/finance-bot/internal/rates"
	"github.com/jefjesuswt/finance-bot/internal/telegram"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No se encontró el archivo .env, usando variables del sistema")
	}
	httpClient := &http.Client{Timeout: 10 * time.Second}

	// telegram token
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_TOKEN no está configurado")
	}

	// allowed chat id
	acid, err := strconv.ParseInt(os.Getenv("ALLOWED_CHAT_ID"), 10, 64)
	if err != nil {
		log.Fatalf("Error parsing ALLOWED_CHAT_ID: %v", err)
	}

	// rates
	ratesService := rates.NewService(httpClient)
	ratesHandler := rates.NewHandler(ratesService)

	// telegram
	tgService := telegram.NewService(token, httpClient)
	tgHandler := telegram.NewHandler(tgService, ratesService, acid)


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
	r.Mount("/api/telegram", telegram.Routes(tgHandler))

	log.Println("🚀 Servidor iniciado en el puerto 8081...")
	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}
