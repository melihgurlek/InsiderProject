package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"

	"github.com/melihgurlek/backend-path/internal/config"
	"github.com/melihgurlek/backend-path/internal/handler"
	"github.com/melihgurlek/backend-path/internal/middleware"
	"github.com/melihgurlek/backend-path/internal/repository"
	"github.com/melihgurlek/backend-path/internal/service"
	"github.com/melihgurlek/backend-path/internal/worker"
	"github.com/melihgurlek/backend-path/pkg"
	"github.com/melihgurlek/backend-path/pkg/cache"
	"github.com/melihgurlek/backend-path/pkg/tracing"
)

func main() {
	// Load environment variables (optional - will use system env vars if .env doesn't exist)
	if err := godotenv.Load(); err != nil {
		log.Info().Msg("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize zerolog (logs to stdout by default)
	log.Info().Msg("Backend Path API starting...")
	log.Info().Str("port", cfg.Port).Str("db_url", cfg.DBUrl).Msg("Loaded configuration")

	// Initialize OpenTelemetry tracing
	jaegerURL := os.Getenv("JAEGER_URL")
	if jaegerURL == "" {
		jaegerURL = "jaeger:4318"
	}

	traceCleanup, err := tracing.InitTracer("backend-path-api", "1.0.0", jaegerURL)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize tracing")
	} else {
		defer traceCleanup()
		log.Info().Msg("OpenTelemetry tracing initialized")
	}

	// Initialize Redis cache
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://redis:6379"
	}

	redisCache, err := cache.NewRedisCache(redisURL)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize Redis cache")
	} else {
		defer redisCache.Close()
		log.Info().Msg("Redis cache initialized")
	}

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
	userHandler := handler.NewUserHandler(userService, cfg.JWTSecret)

	balanceRepo := repository.NewBalancePostgresRepository(conn)
	transactionRepo := repository.NewTransactionPostgresRepository(conn)
	transactionService := service.NewTransactionService(transactionRepo, balanceRepo)
	transactionHandler := handler.NewTransactionHandler(transactionService)

	balanceService := service.NewBalanceService(balanceRepo)
	balanceHandler := handler.NewBalanceHandler(balanceService)

	// Initialize business metrics service
	businessMetricsService := service.NewBusinessMetricsService(
		userRepo,
		transactionRepo,
		balanceRepo,
	)

	testHandler := handler.NewTestHandler()

	// Initialize business metrics handler
	businessMetricsHandler := handler.NewBusinessMetricsHandler(businessMetricsService)

	// Initialize transaction processor (worker pool)
	transactionProcessor := worker.NewTransactionProcessor(
		transactionService,
		balanceService,
		5,   // 5 workers
		100, // queue size of 100
	)

	// Start the transaction processor
	if err := transactionProcessor.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to start transaction processor")
	}
	defer func() {
		if err := transactionProcessor.Stop(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to stop transaction processor")
		}
	}()

	// Start the business metrics service
	businessMetricsService.Start(ctx)
	defer businessMetricsService.Stop()

	batchProcessor := worker.NewBatchProcessor(transactionProcessor, 5, 30*time.Second)

	// Initialize worker handler
	workerHandler := handler.NewWorkerHandler(transactionProcessor, batchProcessor)

	jwtValidator := pkg.NewJWTValidator(cfg.JWTSecret)
	authMiddleware := middleware.NewAuthMiddleware(jwtValidator)

	// Set up chi router
	r := chi.NewRouter()
	r.Use(middleware.DefaultPerformanceMiddleware())
	r.Use(middleware.ErrorMiddleware())

	// Add tracing middleware (should be early in the chain)
	tracingMiddleware := middleware.NewTracingMiddleware()
	r.Use(tracingMiddleware.Middleware)

	// Add metrics middleware
	metricsMiddleware := middleware.NewMetricsMiddleware()
	r.Use(metricsMiddleware.Middleware)

	// Add cache middleware (if Redis is available)
	if redisCache != nil {
		cacheMiddleware := middleware.NewCacheMiddleware(redisCache, 5*time.Minute)
		r.Use(cacheMiddleware.Middleware)
		log.Info().Msg("Cache middleware enabled")
	}

	jsonValidator := &middleware.JSONValidator{}
	validateRegister := middleware.ValidationMiddleware(jsonValidator, func() interface{} { return &handler.RegisterRequest{} })
	validateLogin := middleware.ValidationMiddleware(jsonValidator, func() interface{} { return &handler.LoginRequest{} })
	validateUpdate := middleware.ValidationMiddleware(jsonValidator, func() interface{} { return &handler.UpdateRequest{} })

	r.Route("/api/v1", func(r chi.Router) {
		r.With(validateRegister).Post("/auth/register", userHandler.Register)
		r.With(validateLogin).Post("/auth/login", userHandler.Login)

		// Test routes (no auth required)
		r.Route("/test", func(r chi.Router) {
			testHandler.RegisterRoutes(r)
		})

		// Business metrics routes (no auth required for monitoring)
		r.Route("/metrics", func(r chi.Router) {
			businessMetricsHandler.RegisterRoutes(r)
		})

		r.With(authMiddleware.Middleware).Group(func(r chi.Router) {
			// --- Worker Routes ---
			r.Route("/worker", func(r chi.Router) {
				workerHandler.RegisterRoutes(r)
			})

			// --- User Routes ---
			r.Route("/users", func(r chi.Router) {
				r.With(middleware.RequireRoles("admin")).Get("/", userHandler.ListUsers)
				r.Get("/{id}", userHandler.GetUserByID)
				r.With(validateUpdate).Put("/{id}", userHandler.UpdateUser)
				r.Delete("/{id}", userHandler.DeleteUser)
			})

			// --- Transaction Routes ---
			r.Route("/transactions", func(r chi.Router) {
				transactionHandler.RegisterRoutes(r)
			})

			// --- Balance Routes ---
			r.Route("/balances", func(r chi.Router) {
				balanceHandler.RegisterRoutes(r)
			})
		})
	})

	// Metrics endpoint for Prometheus
	r.Handle("/metrics", promhttp.Handler())

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
