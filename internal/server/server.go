package server

import (
	"context"
	"crypto/subtle"
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

type Server struct {
	config        *config.Config
	vfs           *filesystem.VirtualFS
	store         *storage.PersistentStore
	httpServer    *http.Server
	webdavHandler *handlers.WebDAVHandler
	apiHandler    *handlers.APIHandler
}

func New(cfg *config.Config) (*Server, error) {
	store, err := storage.New(cfg.DataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create persistent store: %w", err)
	}

	log.Printf("💾 Initialized persistent storage in: %s", cfg.DataDir)

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
	}

	server.setupRoutes(mux)

	log.Println("🛠️  HTTP handlers and routes configured")

	return server, nil
}

func (s *Server) setupRoutes(mux *http.ServeMux) {
	apiHandler := s.loggingMiddleware(s.apiHandler.ServeHTTP)
	if s.config.AuthEnabled {
		apiHandler = s.basicAuthMiddleware(apiHandler)
	}
	mux.HandleFunc("/api/", apiHandler)
	mux.HandleFunc("/api/health", s.handleHealth)

	webdavHandler := s.loggingMiddleware(s.webdavHandler.ServeHTTP)
	if s.config.AuthEnabled {
		webdavHandler = s.basicAuthMiddleware(webdavHandler)
	}
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
	// Get file count for startup message
	fileCount, err := s.store.CountFileEntries()
	if err != nil {
		log.Printf("⚠️  Warning: Could not count stored files: %v", err)
		fileCount = -1 // Use -1 to indicate error
	}

	log.Println("📋 Server Configuration:")
	log.Printf("   🌐 Port: %d", s.config.Port)
	log.Printf("   📁 Data Directory: %s", s.config.DataDir)
	log.Printf("   🔄 Redirect Mode: %v", s.config.UseRedirect)
	log.Printf("   🔐 Authentication: %v", s.config.AuthEnabled)
	if s.config.AuthEnabled {
		log.Printf("   👤 Username: %s", s.config.AuthUser)
	}
	log.Printf("   🩺 Health Endpoint: /api/health")
	if fileCount >= 0 {
		if fileCount == 0 {
			log.Printf("   📄 Stored Files: %d (database is empty)", fileCount)
		} else {
			log.Printf("   📄 Stored Files: %d loaded from database", fileCount)
		}
	}
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
	log.Printf("   🩺 Health Check: http://localhost:%d/api/health", s.config.Port)
	log.Println()
	if fileCount == 0 {
		log.Println("💡 Tip: Your virtual filesystem is empty. Add files using:")
		log.Printf("   curl -X POST http://localhost:%d/api/files \\", s.config.Port)
		log.Println("     -H \"Content-Type: application/json\" \\")
		log.Println("     -d '{\"path\":\"/example.pdf\",\"url\":\"https://example.com/file.pdf\"}'")
	} else if fileCount > 0 {
		log.Printf("📚 %d file(s) loaded and ready to serve", fileCount)
	}
	log.Println("🛑 Press Ctrl+C to stop the server")
	log.Println()

	return s.waitForShutdown()
}

// waitForShutdown waits for shutdown signals and gracefully shuts down the server
func (s *Server) waitForShutdown() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println()
	log.Println("🛑 Shutdown signal received. Gracefully shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
		return err
	}

	if err := s.store.Close(); err != nil {
		log.Printf("⚠️  Error closing persistent store: %v", err)
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
