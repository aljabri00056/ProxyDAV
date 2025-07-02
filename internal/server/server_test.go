package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"proxydav/internal/config"
)

func TestNew(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		Port:        8080,
		DataDir:     tempDir,
		UseRedirect: false,
		AuthEnabled: false,
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer server.Stop()

	if server.config != cfg {
		t.Error("Server config not set correctly")
	}

	if server.vfs == nil {
		t.Error("VFS not initialized")
	}

	if server.store == nil {
		t.Error("Store not initialized")
	}

	if server.httpServer == nil {
		t.Error("HTTP server not initialized")
	}

	if server.webdavHandler == nil {
		t.Error("WebDAV handler not initialized")
	}

	if server.apiHandler == nil {
		t.Error("API handler not initialized")
	}

	expectedAddr := ":8080"
	if server.httpServer.Addr != expectedAddr {
		t.Errorf("Expected server address %s, got %s", expectedAddr, server.httpServer.Addr)
	}
}

func TestNew_InvalidDataDir(t *testing.T) {
	cfg := &config.Config{
		Port:        8080,
		DataDir:     "/invalid/path/that/does/not/exist/and/cannot/be/created/due/to/permissions",
		UseRedirect: false,
		AuthEnabled: false,
	}

	server, err := New(cfg)
	if err == nil {
		if server != nil {
			server.Stop()
		}
		t.Error("Expected error when creating server with invalid data directory")
	}
}

func TestServer_HealthEndpoint(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		Port:        8080,
		DataDir:     tempDir,
		UseRedirect: false,
		AuthEnabled: false,
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer server.Stop()

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode health response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status healthy, got %v", response["status"])
	}

	if response["data_dir"] != tempDir {
		t.Errorf("Expected data_dir %s, got %v", tempDir, response["data_dir"])
	}
}

func TestServer_BasicAuthMiddleware(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		Port:        8080,
		DataDir:     tempDir,
		UseRedirect: false,
		AuthEnabled: true,
		AuthUser:    "testuser",
		AuthPass:    "testpass",
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer server.Stop()

	handlerCalled := false
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}

	authHandler := server.basicAuthMiddleware(testHandler)

	tests := []struct {
		name           string
		username       string
		password       string
		expectedStatus int
		expectHandler  bool
	}{
		{
			name:           "valid credentials",
			username:       "testuser",
			password:       "testpass",
			expectedStatus: http.StatusOK,
			expectHandler:  true,
		},
		{
			name:           "invalid username",
			username:       "wronguser",
			password:       "testpass",
			expectedStatus: http.StatusUnauthorized,
			expectHandler:  false,
		},
		{
			name:           "invalid password",
			username:       "testuser",
			password:       "wrongpass",
			expectedStatus: http.StatusUnauthorized,
			expectHandler:  false,
		},
		{
			name:           "empty credentials",
			username:       "",
			password:       "",
			expectedStatus: http.StatusUnauthorized,
			expectHandler:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled = false
			req := httptest.NewRequest("GET", "/", nil)
			if tt.username != "" || tt.password != "" {
				req.SetBasicAuth(tt.username, tt.password)
			}
			w := httptest.NewRecorder()

			authHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			if handlerCalled != tt.expectHandler {
				t.Errorf("Expected handler called: %v, got: %v", tt.expectHandler, handlerCalled)
			}

			if tt.expectedStatus == http.StatusUnauthorized {
				authHeader := w.Header().Get("WWW-Authenticate")
				if !strings.Contains(authHeader, "Basic realm") {
					t.Errorf("Expected WWW-Authenticate header with Basic realm, got: %s", authHeader)
				}
			}
		})
	}
}

func TestServer_BasicAuthMiddleware_NoAuth(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		Port:        8080,
		DataDir:     tempDir,
		UseRedirect: false,
		AuthEnabled: false, // Auth disabled
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer server.Stop()

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	// Test that health endpoint works without auth when auth is disabled
	server.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestServer_LoggingMiddleware(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		Port:        8080,
		DataDir:     tempDir,
		UseRedirect: false,
		AuthEnabled: false,
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer server.Stop()

	handlerCalled := false
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusCreated)
	}

	loggedHandler := server.loggingMiddleware(testHandler)

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()

	loggedHandler(w, req)

	if !handlerCalled {
		t.Error("Expected underlying handler to be called")
	}

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestResponseWriter(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

	// Test default status code
	if rw.statusCode != http.StatusOK {
		t.Errorf("Expected default status code %d, got %d", http.StatusOK, rw.statusCode)
	}

	// Test WriteHeader
	rw.WriteHeader(http.StatusCreated)
	if rw.statusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, rw.statusCode)
	}

	if w.Code != http.StatusCreated {
		t.Errorf("Expected underlying ResponseWriter status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestServer_Stop(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		Port:        8080,
		DataDir:     tempDir,
		UseRedirect: false,
		AuthEnabled: false,
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start server in background
	go func() {
		server.httpServer.ListenAndServe()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Stop server
	err = server.Stop()
	if err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}
}

func TestServer_RouteSetup(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		authEnabled bool
		path        string
		method      string
	}{
		{
			name:        "health endpoint without auth",
			authEnabled: false,
			path:        "/api/health",
			method:      "GET",
		},
		{
			name:        "health endpoint with auth",
			authEnabled: true,
			path:        "/api/health",
			method:      "GET",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Port:        8080,
				DataDir:     tempDir,
				UseRedirect: false,
				AuthEnabled: tt.authEnabled,
				AuthUser:    "testuser",
				AuthPass:    "testpass",
			}

			server, err := New(cfg)
			if err != nil {
				t.Fatalf("Failed to create server: %v", err)
			}
			defer server.Stop()

			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.authEnabled {
				req.SetBasicAuth("testuser", "testpass")
			}
			w := httptest.NewRecorder()

			server.httpServer.Handler.ServeHTTP(w, req)

			// Health endpoint should always return 200 with valid auth
			if tt.path == "/api/health" && w.Code != http.StatusOK {
				t.Errorf("Expected status code %d for health endpoint, got %d", http.StatusOK, w.Code)
			}
		})
	}
}
