package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"github.com/yourusername/bahasa-daerah-platform/internal/auth"
	"github.com/yourusername/bahasa-daerah-platform/internal/language"
	"github.com/yourusername/bahasa-daerah-platform/internal/phrase"
	"github.com/yourusername/bahasa-daerah-platform/internal/validation"
	"github.com/yourusername/bahasa-daerah-platform/pkg/db"
)

func main() {
	// Load .env if present (ignored in production where env vars are set directly)
	_ = godotenv.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ── Database ──────────────────────────────────────────────────────────────
	pool, err := db.New(ctx)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()
	log.Println("✓ database connected")

	// ── Migrations ────────────────────────────────────────────────────────────
	if err := db.RunMigrations(ctx, pool, db.Migrations); err != nil {
		log.Fatalf("run migrations: %v", err)
	}
	log.Println("✓ migrations applied")

	// ── Router ────────────────────────────────────────────────────────────────
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Timeout(30 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	// ── Auth routes ───────────────────────────────────────────────────────────
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "changeme"
	}

	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authSvc)

	r.Route("/api/v1/auth", authHandler.Routes(jwtSecret))

	// ── Language routes ───────────────────────────────────────────────────────
	langRepo := language.NewRepository(pool)
	langSvc := language.NewService(langRepo)
	langHandler := language.NewHandler(langSvc)

	r.Route("/api/v1/languages", langHandler.PublicRoutes())
	r.Route("/api/v1/admin/languages", langHandler.AdminRoutes(jwtSecret))

	// ── Phrase routes ─────────────────────────────────────────────────────────
	phraseRepo := phrase.NewRepository(pool)
	phraseSvc := phrase.NewService(phraseRepo)
	phraseHandler := phrase.NewHandler(phraseSvc)

	// ── Validation routes (votes, flags) ──────────────────────────────────────
	validationRepo := validation.NewRepository(pool)
	validationSvc := validation.NewService(validationRepo)
	validationHandler := validation.NewHandler(validationSvc)

	// Mount both phrase and validation routes under /api/v1/phrases
	r.Route("/api/v1/phrases", func(router chi.Router) {
		phraseHandler.Routes(jwtSecret)(router)
		validationHandler.Routes(jwtSecret)(router)
	})

	// ── Server ────────────────────────────────────────────────────────────────
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("✓ server listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}
