package server

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"proxydav/internal/config"
	"proxydav/internal/filesystem"
	"proxydav/internal/handlers"
	"proxydav/internal/storage"
)

// ErrRestart is returned when the server should restart
var ErrRestart = errors.New("server restart requested")

type Server struct {
	config        *config.Config
	vfs           *filesystem.VirtualFS
	store         *storage.PersistentStore
	httpServer    *http.Server
	webdavHandler *handlers.WebDAVHandler
	apiHandler    *handlers.APIHandler
	adminHandler  *handlers.AdminHandler
	restartChan   chan bool // Channel to signal restart
	shutdownChan  chan bool // Channel to signal shutdown
}

func New(cfg *config.Config) (*Server, error) {
	store, err := storage.New(cfg.DataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create persistent store: %w", err)
	}

	log.Printf("💾 Initialized persistent storage in: %s", cfg.DataDir)

	// Try to load saved configuration from database
	if savedConfig, err := config.LoadFromStore(store); err == nil && savedConfig != nil {
		log.Printf("📋 Loaded configuration from database")
		cfg = savedConfig
	}

	vfs, err := filesystem.New(store)
	if err != nil {
		store.Close()
		return nil, fmt.Errorf("failed to create virtual filesystem: %w", err)
	}

	log.Println("🗂️  Virtual filesystem initialized")

	webdavHandler := handlers.NewWebDAVHandler(vfs, store, cfg.UseRedirect)
	apiHandler := handlers.NewAPIHandler(vfs)

	mux := http.NewServeMux()
	server := &Server{
		config:        cfg,
		vfs:           vfs,
		store:         store,
		webdavHandler: webdavHandler,
		apiHandler:    apiHandler,
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Port),
			Handler:      mux,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		restartChan:  make(chan bool),
		shutdownChan: make(chan bool),
	}

	// Create admin handler with server as config updater
	adminHandler := handlers.NewAdminHandler(vfs, store, cfg, server)
	server.adminHandler = adminHandler

	server.setupRoutes(mux)

	log.Println("🛠️  HTTP handlers and routes configured")

	return server, nil
}

func (s *Server) setupRoutes(mux *http.ServeMux) {
	// Use dynamic middleware that checks current config state
	adminHandler := s.loggingMiddleware(s.dynamicAuthMiddleware(s.adminHandler.ServeHTTP))
	mux.HandleFunc("/admin/", adminHandler)

	apiHandler := s.loggingMiddleware(s.dynamicAuthMiddleware(s.apiHandler.ServeHTTP))
	mux.HandleFunc("/api/", apiHandler)
	mux.HandleFunc("/api/health", s.handleHealth)

	// WebDAV routes (catch-all, should be last)
	webdavHandler := s.loggingMiddleware(s.dynamicAuthMiddleware(s.webdavHandler.ServeHTTP))
	mux.HandleFunc("/", webdavHandler)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"healthy","data_dir":"%s"}`, s.config.DataDir)
}

// basicAuthMiddleware provides HTTP Basic authentication
func (s *Server) basicAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="ProxyDAV"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Use constant-time comparison to prevent timing attacks
		usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(s.config.AuthUser)) == 1
		passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(s.config.AuthPass)) == 1

		if !usernameMatch || !passwordMatch {
			w.Header().Set("WWW-Authenticate", `Basic realm="ProxyDAV"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

// dynamicAuthMiddleware applies authentication only when enabled in current config
func (s *Server) dynamicAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.config.AuthEnabled {
			s.basicAuthMiddleware(next)(w, r)
		} else {
			next(w, r)
		}
	}
}

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next(wrapped, r)

		duration := time.Since(start)
		statusEmoji := "✅"
		if wrapped.statusCode >= 400 && wrapped.statusCode < 500 {
			statusEmoji = "⚠️ "
		} else if wrapped.statusCode >= 500 {
			statusEmoji = "❌"
		}

		log.Printf("%s %s %s %d %v %s", statusEmoji, r.Method, r.URL.Path, wrapped.statusCode, duration, r.UserAgent())
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (s *Server) Start() error {

	log.Println("📋 Server Configuration:")
	log.Printf("   🌐 Port: %d", s.config.Port)
	log.Printf("   📁 Data Directory: %s", s.config.DataDir)
	log.Printf("   🔄 Redirect Mode: %v", s.config.UseRedirect)
	log.Printf("   🔐 Authentication: %v", s.config.AuthEnabled)
	if s.config.AuthEnabled {
		log.Printf("   👤 Username: %s", s.config.AuthUser)
	}
	log.Printf("   🩺 Health Endpoint: /api/health")
	log.Println()

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed to start: %v", err)
		}
	}()

	log.Println("✅ ProxyDAV server started successfully!")
	log.Printf("🌍 Server URLs:")
	log.Printf("   🔗 WebDAV Endpoint: http://localhost:%d/", s.config.Port)
	log.Printf("   🛠️  API Endpoint: http://localhost:%d/api/", s.config.Port)
	log.Printf("   🎛️  Admin Panel: http://localhost:%d/admin/", s.config.Port)
	log.Println()
	log.Println("🛑 Press Ctrl+C to stop the server")
	log.Println()

	return s.waitForShutdown()
}

// waitForShutdown waits for shutdown signals and gracefully shuts down the server
func (s *Server) waitForShutdown() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	var isRestart bool

	select {
	case <-quit:
		log.Println()
		log.Println("🛑 Shutdown signal received. Gracefully shutting down...")
	case <-s.restartChan:
		log.Println()
		log.Println("🔄 Restart signal received. Gracefully restarting...")
		isRestart = true
	case <-s.shutdownChan:
		log.Println()
		log.Println("🛑 Admin shutdown signal received. Gracefully shutting down...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
		return err
	}

	if err := s.store.Close(); err != nil {
		log.Printf("⚠️  Error closing persistent store: %v", err)
	}

	if isRestart {
		log.Println("✅ Server shutdown complete. Preparing to restart...")
		return ErrRestart
	}

	log.Println("✅ Server shutdown complete. Goodbye!")
	return nil
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	return s.store.Close()
}

// UpdateConfig applies configuration changes that can take effect without restart
func (s *Server) UpdateConfig(newConfig *config.Config) error {
	if err := newConfig.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	s.config = newConfig

	s.webdavHandler.SetUseRedirect(newConfig.UseRedirect)

	if err := newConfig.SaveToStore(s.store); err != nil {
		log.Printf("⚠️  Warning: Failed to save configuration to database: %v", err)
	} else {
		log.Printf("💾 Configuration saved to database")
	}

	log.Printf("🔄 Configuration updated successfully")
	log.Printf("   🔄 Redirect Mode: %v", newConfig.UseRedirect)
	log.Printf("   🔐 Authentication: %v", newConfig.AuthEnabled)
	if newConfig.AuthEnabled {
		log.Printf("   👤 Username: %s", newConfig.AuthUser)
	}

	return nil
}

// GetConfig returns a copy of the current configuration
func (s *Server) GetConfig() *config.Config {
	configCopy := *s.config
	return &configCopy
}

// Restart signals the server to restart gracefully
func (s *Server) Restart() error {
	select {
	case s.restartChan <- true:
		return nil
	default:
		return errors.New("restart already in progress")
	}
}

// Shutdown signals the server to shutdown gracefully via admin panel
func (s *Server) Shutdown() error {
	select {
	case s.shutdownChan <- true:
		return nil
	default:
		return errors.New("shutdown already in progress")
	}
}
