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

	"github.com/ayush10/email-waitlist/internal/config"
	"github.com/ayush10/email-waitlist/internal/database"
	"github.com/ayush10/email-waitlist/internal/handler"
	"github.com/ayush10/email-waitlist/internal/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := database.RunMigrations(ctx, pool); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	log.Println("database connected and migrations applied")

	// Handlers
	subscribeH := handler.NewSubscribeHandler(pool)
	subscribersH := handler.NewSubscribersHandler(pool)
	statsH := handler.NewStatsHandler(pool)
	projectsH := handler.NewProjectsHandler(pool)

	// Middleware
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit)
	apiAuth := middleware.APIKeyAuth(pool)
	adminAuth := middleware.AdminAuth(cfg.AdminKey)
	cors := middleware.CORS(pool)

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// API docs
	mux.HandleFunc("GET /docs", handler.DocsHandler)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/docs", http.StatusMovedPermanently)
	})

	// Public endpoints (API key auth)
	mux.Handle("POST /api/v1/subscribe", chain(subscribeH, cors, rateLimiter.Middleware(), apiAuth))

	// Project-scoped management endpoints (API key auth)
	mux.Handle("GET /api/v1/subscribers", chain(http.HandlerFunc(subscribersH.List), cors, apiAuth))
	mux.Handle("DELETE /api/v1/subscribers/{email}", chain(http.HandlerFunc(subscribersH.Delete), cors, apiAuth))
	mux.Handle("GET /api/v1/subscribers/export", chain(http.HandlerFunc(subscribersH.Export), cors, apiAuth))
	mux.Handle("GET /api/v1/stats", chain(statsH, cors, apiAuth))

	// Admin endpoints (admin key auth)
	mux.Handle("POST /api/v1/projects", chain(http.HandlerFunc(projectsH.Create), cors, adminAuth))
	mux.Handle("GET /api/v1/projects", chain(http.HandlerFunc(projectsH.List), cors, adminAuth))

	// Handle OPTIONS for all api routes
	mux.Handle("OPTIONS /api/", chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), cors))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	log.Printf("email waitlist service running on :%d", cfg.Port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

// chain applies middleware in order (outermost first)
func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
