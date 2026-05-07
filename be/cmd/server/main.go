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
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/yourusername/bahasa-daerah-platform/internal/admin"
	aiPkg "github.com/yourusername/bahasa-daerah-platform/internal/ai"
	"github.com/yourusername/bahasa-daerah-platform/internal/auth"
	"github.com/yourusername/bahasa-daerah-platform/internal/flashcard"
	"github.com/yourusername/bahasa-daerah-platform/internal/language"
	"github.com/yourusername/bahasa-daerah-platform/internal/phrase"
	"github.com/yourusername/bahasa-daerah-platform/internal/search"
	storagePkg "github.com/yourusername/bahasa-daerah-platform/internal/storage"
	"github.com/yourusername/bahasa-daerah-platform/internal/validation"
	"github.com/yourusername/bahasa-daerah-platform/pkg/db"
)

// pipelineEnqueuer wraps the AI pipeline to satisfy phrase.Enqueuer.
type pipelineEnqueuer struct {
	pipeline *aiPkg.Pipeline
}

func (e *pipelineEnqueuer) EnqueuePhrase(phraseID uuid.UUID, textLatin, translation string) {
	e.pipeline.Enqueue(aiPkg.Job{
		PhraseID:    phraseID,
		TextLatin:   textLatin,
		Translation: translation,
	})
}

func main() {
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

	if err := db.RunMigrations(ctx, pool, db.Migrations); err != nil {
		log.Fatalf("run migrations: %v", err)
	}
	log.Println("✓ migrations applied")

	// ── Storage ───────────────────────────────────────────────────────────────
	var storageSvc *storagePkg.Service
	s3, err := storagePkg.NewS3Storage()
	if err != nil {
		log.Printf("⚠ S3 not configured, using no-op storage: %v", err)
		s3 = &storagePkg.NoOpStorage{}
	}
	storageRepo := storagePkg.NewRepository(pool)
	storageSvc = storagePkg.NewService(storageRepo, s3)
	log.Println("✓ storage ready")

	// ── AI Pipeline ───────────────────────────────────────────────────────────
	aiRepo := aiPkg.NewRepository(pool)
	aiSvc := aiPkg.NewService(aiRepo)
	pipeline := aiPkg.NewPipeline(aiSvc, 3)
	pipeline.Start(ctx)
	log.Println("✓ AI pipeline started")

	// ── JWT ───────────────────────────────────────────────────────────────────
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "changeme"
	}

	// ── Auth ──────────────────────────────────────────────────────────────────
	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authSvc)

	// ── Language ──────────────────────────────────────────────────────────────
	langRepo := language.NewRepository(pool)
	langSvc := language.NewService(langRepo)
	langHandler := language.NewHandler(langSvc)

	// ── Phrase ────────────────────────────────────────────────────────────────
	phraseRepo := phrase.NewRepository(pool)
	enqueuer := &pipelineEnqueuer{pipeline: pipeline}
	phraseSvc := phrase.NewService(phraseRepo, enqueuer)
	phraseHandler := phrase.NewHandler(phraseSvc)

	// ── Validation ────────────────────────────────────────────────────────────
	validationRepo := validation.NewRepository(pool)
	validationSvc := validation.NewService(validationRepo)
	validationHandler := validation.NewHandler(validationSvc)

	// ── Flashcard + Practice ──────────────────────────────────────────────────
	flashcardRepo := flashcard.NewRepository(pool)
	flashcardSvc := flashcard.NewService(flashcardRepo)
	flashcardHandler := flashcard.NewHandler(flashcardSvc)

	// ── Search ────────────────────────────────────────────────────────────────
	searchRepo := search.NewRepository(pool)
	searchSvc := search.NewService(searchRepo)
	searchHandler := search.NewHandler(searchSvc)

	// ── Admin ─────────────────────────────────────────────────────────────────
	adminRepo := admin.NewRepository(pool)
	adminSvc := admin.NewService(adminRepo, storageSvc)
	adminHandler := admin.NewHandler(adminSvc)

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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	// Auth
	r.Route("/api/v1/auth", authHandler.Routes(jwtSecret))

	// Languages
	r.Route("/api/v1/languages", langHandler.PublicRoutes())
	r.Route("/api/v1/admin/languages", langHandler.AdminRoutes(jwtSecret))

	// Phrases + Votes/Flags
	r.Route("/api/v1/phrases", func(router chi.Router) {
		phraseHandler.Routes(jwtSecret)(router)
		validationHandler.Routes(jwtSecret)(router)
	})

	// Flashcards, Conversation Scenarios, Practice
	r.Route("/api/v1", flashcardHandler.Routes(jwtSecret))

	// Search
	r.Route("/api/v1", searchHandler.Routes(jwtSecret))

	// Admin
	r.Route("/api/v1/admin", adminHandler.Routes(jwtSecret))

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
