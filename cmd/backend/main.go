package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/melihgurlek/backend-path/internal/config"
	"github.com/melihgurlek/backend-path/internal/handler"
	"github.com/melihgurlek/backend-path/internal/repository"
	"github.com/melihgurlek/backend-path/internal/service"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize zerolog (logs to stdout by default)
	log.Info().Msg("Backend Path API starting...")
	log.Info().Str("port", cfg.Port).Str("db_url", cfg.DBUrl).Msg("Loaded configuration")

	// Connect to PostgreSQL
	ctx := context.Background()
	conn, err := repository.ConnectDB(ctx, cfg.DBUrl)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	log.Info().Msg("Connected to PostgreSQL database!")
	defer func() {
		_ = conn.Close(ctx)
		log.Info().Msg("Database connection closed.")
	}()

	// Set up repository, service, handler
	userRepo := repository.NewUserPostgresRepository(conn)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// Set up chi router
	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		userHandler.RegisterRoutes(r)
	})

	// Start HTTP server in a goroutine
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}
	go func() {
		log.Info().Str("port", cfg.Port).Msg("HTTP server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server error")
		}
	}()

	// Graceful shutdown setup
	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Info().Msg("Press Ctrl+C to exit")
	<-shutdownCtx.Done() // Wait for shutdown signal
	log.Info().Msg("Shutting down gracefully...")

	// Shutdown HTTP server
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxTimeout); err != nil {
		log.Error().Err(err).Msg("HTTP server shutdown error")
	}
	log.Info().Msg("Shutdown complete.")
}
