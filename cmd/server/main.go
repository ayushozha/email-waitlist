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
	"github.com/ayush10/email-waitlist/internal/email"
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

	// Email service (optional — disabled if RESEND_API_KEY is not set)
	var emailService *email.Service
	if cfg.ResendAPIKey != "" {
		emailService = email.NewService(cfg.ResendAPIKey, pool, cfg.DefaultFromEmail)
		log.Println("email confirmations enabled (Resend)")
	} else {
		log.Println("email confirmations disabled (set RESEND_API_KEY to enable)")
	}

	// Handlers
	subscribeH := handler.NewSubscribeHandler(pool, emailService)
	subscribersH := handler.NewSubscribersHandler(pool)
	statsH := handler.NewStatsHandler(pool)
	projectsH := handler.NewProjectsHandler(pool)
	emailTmplH := handler.NewEmailTemplateHandler(pool)

	// Middleware. CORS must run inside auth so per-project allowed_origins
	// (read from request context) can actually be enforced.
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit, cfg.TrustProxy)
	subscribeAuth := middleware.APIKeyAuth(pool, middleware.ScopeSubscribe)
	manageAuth := middleware.APIKeyAuth(pool, middleware.ScopeManage)
	adminAuth := middleware.AdminAuth(cfg.AdminKey)
	cors := middleware.CORS

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Homepage and API docs
	mux.HandleFunc("GET /docs", handler.DocsHandler)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		handler.HomepageHandler(w, r)
	})

	// Public endpoint (publishable or secret key)
	mux.Handle("POST /api/v1/subscribe", chain(subscribeH, rateLimiter.Middleware(), subscribeAuth, cors))

	// Project-scoped management endpoints (secret key only)
	mux.Handle("GET /api/v1/subscribers", chain(http.HandlerFunc(subscribersH.List), manageAuth, cors))
	mux.Handle("DELETE /api/v1/subscribers/{email}", chain(http.HandlerFunc(subscribersH.Delete), manageAuth, cors))
	mux.Handle("GET /api/v1/subscribers/export", chain(http.HandlerFunc(subscribersH.Export), manageAuth, cors))
	mux.Handle("GET /api/v1/stats", chain(statsH, manageAuth, cors))

	// Email template management (secret key only)
	mux.Handle("GET /api/v1/email-template", chain(http.HandlerFunc(emailTmplH.Get), manageAuth, cors))
	mux.Handle("PUT /api/v1/email-template", chain(http.HandlerFunc(emailTmplH.Upsert), manageAuth, cors))
	mux.Handle("DELETE /api/v1/email-template", chain(http.HandlerFunc(emailTmplH.Delete), manageAuth, cors))

	// Admin endpoints (admin key auth)
	mux.Handle("POST /api/v1/projects", chain(http.HandlerFunc(projectsH.Create), adminAuth, cors))
	mux.Handle("GET /api/v1/projects", chain(http.HandlerFunc(projectsH.List), adminAuth, cors))

	// Handle OPTIONS preflight for all api routes (no auth — browsers never
	// send custom headers on preflight; enforcement happens on the request)
	mux.Handle("OPTIONS /api/", cors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown: ListenAndServe returns as soon as Shutdown starts,
	// so main must wait for shutdownDone or in-flight requests get killed.
	shutdownDone := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
		close(shutdownDone)
	}()

	log.Printf("email waitlist service running on :%d", cfg.Port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
	<-shutdownDone
}

// chain applies middleware in order (outermost first)
func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
