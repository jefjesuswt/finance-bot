package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jefjesuswt/finance-bot/internal/github"
	"github.com/jefjesuswt/finance-bot/internal/processor"
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

	// allowed chat id
	acid, err := strconv.ParseInt(os.Getenv("ALLOWED_CHAT_ID"), 10, 64)
	if err != nil {
		log.Fatalf("Error parsing ALLOWED_CHAT_ID: %v", err)
	}

	// git envs
	gitToken := os.Getenv("GIT_TOKEN")
	gitOwner := os.Getenv("GIT_OWNER")
	gitRepo := os.Getenv("GIT_REPO")

	// vault
	obsidianBasePath := os.Getenv("OBSIDIAN_BASE_PATH")

	if obsidianBasePath == "" {
		obsidianBasePath = "08 - Finances/Transacciones"
		log.Printf("⚠️ [WARN] OBSIDIAN_BASE_PATH vacío. Usando por defecto: '%s'", obsidianBasePath)
	}

	requiredVars := map[string]string{
		"TELEGRAM_TOKEN": token,
		"GIT_TOKEN":      gitToken,
		"GIT_OWNER":      gitOwner,
		"GIT_REPO":       gitRepo,
	}

	for key, value := range requiredVars {
		if value == "" {
			log.Fatalf("❌ [FATAL] Falta la variable de entorno: %s", key)
		}
	}

	// github
	ghClient := github.NewClient(httpClient, gitToken, gitOwner, gitRepo)

	// rates
	ratesService := rates.NewService(httpClient)
	ratesHandler := rates.NewHandler(ratesService)

	//processor
	processorService := processor.NewService(ratesService, ghClient, obsidianBasePath)

	// telegram
	tgService := telegram.NewService(token, httpClient)
	tgHandler := telegram.NewHandler(tgService, processorService, acid)


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
