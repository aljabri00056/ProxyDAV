package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"proxydav/internal/filesystem"
	"proxydav/internal/storage"
	"proxydav/pkg/types"
)

func createTestVFS(t *testing.T) *filesystem.VirtualFS {
	tempDir := t.TempDir()
	store, err := storage.New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	vfs, err := filesystem.New(store)
	if err != nil {
		t.Fatalf("Failed to create VFS: %v", err)
	}
	return vfs
}

func TestAPIHandler_ListFiles(t *testing.T) {
	vfs := createTestVFS(t)
	handler := NewAPIHandler(vfs)

	// Add some test files
	vfs.AddFile("/test1.txt", "https://example.com/test1.txt")
	vfs.AddFile("/test2.txt", "https://example.com/test2.txt")

	req := httptest.NewRequest("GET", "/api/files", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response APIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}

	files, ok := data["files"].([]interface{})
	if !ok {
		t.Fatal("Expected files to be an array")
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}
}

func TestAPIHandler_AddFiles(t *testing.T) {
	vfs := createTestVFS(t)
	handler := NewAPIHandler(vfs)

	request := AddFilesRequest{
		Files: []types.FileEntry{
			{Path: "/test1.txt", URL: "https://example.com/test1.txt"},
			{Path: "/test2.txt", URL: "https://example.com/test2.txt"},
		},
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/api/files/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	var response APIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}

	// Verify files were added
	if !vfs.Exists("/test1.txt") {
		t.Error("Expected /test1.txt to exist")
	}
	if !vfs.Exists("/test2.txt") {
		t.Error("Expected /test2.txt to exist")
	}
}

func TestAPIHandler_DeleteFiles(t *testing.T) {
	vfs := createTestVFS(t)
	handler := NewAPIHandler(vfs)

	// Add test files first
	vfs.AddFile("/test1.txt", "https://example.com/test1.txt")
	vfs.AddFile("/test2.txt", "https://example.com/test2.txt")

	request := DeleteFilesRequest{
		Files: []types.FileEntry{
			{Path: "/test1.txt"},
			{Path: "/test2.txt"},
		},
	}

	body, _ := json.Marshal(request)
	req := httptest.NewRequest("DELETE", "/api/files/delete", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response APIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}

	// Verify files were deleted
	if vfs.Exists("/test1.txt") {
		t.Error("Expected /test1.txt to be deleted")
	}
	if vfs.Exists("/test2.txt") {
		t.Error("Expected /test2.txt to be deleted")
	}
}

func TestAPIHandler_InvalidEndpoints(t *testing.T) {
	vfs := createTestVFS(t)
	handler := NewAPIHandler(vfs)

	tests := []struct {
		method string
		path   string
		status int
	}{
		{"POST", "/api/files", http.StatusNotFound},                 // Old single file add
		{"PUT", "/api/files/test.txt", http.StatusMethodNotAllowed}, // PUT method not allowed
		{"DELETE", "/api/files/test.txt", http.StatusNotFound},      // Old single file delete
		{"POST", "/api/files/bulk", http.StatusNotFound},            // Old bulk endpoint
		{"GET", "/api/files/something", http.StatusNotFound},        // Invalid GET path
	}

	for _, test := range tests {
		req := httptest.NewRequest(test.method, test.path, nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != test.status {
			t.Errorf("Expected status code %d for %s %s, got %d", test.status, test.method, test.path, w.Code)
		}
	}
}
