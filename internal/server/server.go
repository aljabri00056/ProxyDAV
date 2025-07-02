package server

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"proxydav/internal/cache"
	"proxydav/internal/config"
	"proxydav/internal/filesystem"
	"proxydav/internal/handlers"
	"proxydav/pkg/types"
)

// Server represents the ProxyDAV server
type Server struct {
	config         *config.Config
	vfs            *filesystem.VirtualFS
	cache          *cache.MetadataCache
	httpServer     *http.Server
	webdavHandler  *handlers.WebDAVHandler
	browserHandler *handlers.BrowserHandler
	apiHandler     *handlers.APIHandler
}

// New creates a new ProxyDAV server
func New(cfg *config.Config, files []types.FileEntry) *Server {
	// Create virtual filesystem
	vfs := filesystem.New(files)

	// Create metadata cache
	cacheTTL := time.Duration(cfg.CacheTTL) * time.Second
	metadataCache := cache.New(cacheTTL, cfg.MaxCacheSize)

	// Create handlers
	webdavHandler := handlers.NewWebDAVHandler(vfs, metadataCache, cfg.UseRedirect)
	browserHandler := handlers.NewBrowserHandler(vfs)
	apiHandler := handlers.NewAPIHandler(vfs)

	// Create HTTP server
	mux := http.NewServeMux()
	server := &Server{
		config:         cfg,
		vfs:            vfs,
		cache:          metadataCache,
		webdavHandler:  webdavHandler,
		browserHandler: browserHandler,
		apiHandler:     apiHandler,
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Port),
			Handler:      mux,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}

	// Setup routes
	server.setupRoutes(mux)

	return server
}

// setupRoutes sets up the HTTP routes
func (s *Server) setupRoutes(mux *http.ServeMux) {
	// Health check endpoint
	mux.HandleFunc(s.config.HealthPath, s.handleHealth)

	// API endpoints for file management
	apiHandler := s.loggingMiddleware(s.apiHandler.ServeHTTP)
	if s.config.AuthEnabled {
		apiHandler = s.basicAuthMiddleware(apiHandler)
	}
	mux.HandleFunc("/api/", apiHandler)

	// Main handler with middleware
	handler := s.loggingMiddleware(s.routeRequest)
	if s.config.AuthEnabled {
		handler = s.basicAuthMiddleware(handler)
	}

	mux.HandleFunc("/", handler)
}

// routeRequest routes requests to appropriate handlers
func (s *Server) routeRequest(w http.ResponseWriter, r *http.Request) {
	// Route based on Accept header and method
	if r.Method == "GET" && strings.Contains(r.Header.Get("Accept"), "text/html") {
		s.browserHandler.ServeHTTP(w, r)
	} else {
		s.webdavHandler.ServeHTTP(w, r)
	}
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"healthy","cache_size":%d}`, s.cache.Size())
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

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next(wrapped, r)

		duration := time.Since(start)
		log.Printf("%s %s %d %v %s", r.Method, r.URL.Path, wrapped.statusCode, duration, r.UserAgent())
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

// Start starts the server
func (s *Server) Start() error {
	log.Printf("Starting ProxyDAV server on port %d", s.config.Port)
	log.Printf("Cache TTL: %d seconds", s.config.CacheTTL)
	log.Printf("Use redirect: %v", s.config.UseRedirect)
	log.Printf("Authentication: %v", s.config.AuthEnabled)
	log.Printf("Health endpoint: %s", s.config.HealthPath)

	// Start server in a goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	log.Printf("Server started successfully on http://localhost:%d", s.config.Port)

	// Wait for interrupt signal
	return s.waitForShutdown()
}

// waitForShutdown waits for shutdown signals and gracefully shuts down the server
func (s *Server) waitForShutdown() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return err
	}

	// Close cache
	s.cache.Close()

	log.Println("Server shutdown complete")
	return nil
}

// Stop stops the server gracefully
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	s.cache.Close()
	return nil
}
