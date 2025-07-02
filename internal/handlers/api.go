package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	"proxydav/internal/filesystem"
	"proxydav/pkg/types"
)

// APIHandler handles REST API requests for file management
type APIHandler struct {
	vfs *filesystem.VirtualFS
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(vfs *filesystem.VirtualFS) *APIHandler {
	return &APIHandler{
		vfs: vfs,
	}
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// BulkOperation represents a bulk file operation request
type BulkOperation struct {
	Operation string            `json:"operation"` // "add" or "remove"
	Files     []types.FileEntry `json:"files"`
}

// FileListResponse represents the response for listing files
type FileListResponse struct {
	Files []types.FileEntry `json:"files"`
	Total int               `json:"total"`
}

// ServeHTTP handles API requests
func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse the path to determine the operation
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 || pathParts[0] != "api" || pathParts[1] != "files" {
		h.sendError(w, http.StatusNotFound, "Invalid API endpoint")
		return
	}

	switch r.Method {
	case "GET":
		h.handleListFiles(w, r)
	case "POST":
		if len(pathParts) >= 3 && pathParts[2] == "bulk" {
			h.handleBulkOperation(w, r)
		} else {
			h.handleAddFile(w, r)
		}
	case "PUT":
		if len(pathParts) >= 3 {
			h.handleUpdateFile(w, r, strings.Join(pathParts[2:], "/"))
		} else {
			h.sendError(w, http.StatusBadRequest, "File path required for PUT operation")
		}
	case "DELETE":
		if len(pathParts) >= 3 {
			h.handleDeleteFile(w, r, strings.Join(pathParts[2:], "/"))
		} else {
			h.sendError(w, http.StatusBadRequest, "File path required for DELETE operation")
		}
	default:
		h.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleListFiles handles GET /api/files - list all files
func (h *APIHandler) handleListFiles(w http.ResponseWriter, r *http.Request) {
	files := h.vfs.GetAllFiles()

	response := FileListResponse{
		Files: files,
		Total: len(files),
	}

	h.sendSuccess(w, http.StatusOK, "Files retrieved successfully", response)
}

// handleAddFile handles POST /api/files - add a single file
func (h *APIHandler) handleAddFile(w http.ResponseWriter, r *http.Request) {
	var file types.FileEntry
	if err := json.NewDecoder(r.Body).Decode(&file); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
		return
	}

	if err := h.validateFileEntry(file); err != nil {
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Normalize path
	file.Path = path.Clean("/" + strings.TrimPrefix(file.Path, "/"))

	if err := h.vfs.AddFile(file.Path, file.URL); err != nil {
		h.sendError(w, http.StatusConflict, "Failed to add file: "+err.Error())
		return
	}

	h.sendSuccess(w, http.StatusCreated, "File added successfully", file)
}

// handleUpdateFile handles PUT /api/files/{path} - update a single file
func (h *APIHandler) handleUpdateFile(w http.ResponseWriter, r *http.Request, filePath string) {
	// Decode path parameter
	filePath = "/" + filePath
	filePath = path.Clean(filePath)

	var updateData struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
		return
	}

	if updateData.URL == "" {
		h.sendError(w, http.StatusBadRequest, "url is required")
		return
	}

	if !strings.HasPrefix(updateData.URL, "http://") && !strings.HasPrefix(updateData.URL, "https://") {
		h.sendError(w, http.StatusBadRequest, "url must be a valid HTTP or HTTPS URL")
		return
	}

	if !h.vfs.Exists(filePath) {
		h.sendError(w, http.StatusNotFound, "File not found")
		return
	}

	if h.vfs.IsDir(filePath) {
		h.sendError(w, http.StatusBadRequest, "Cannot update directory")
		return
	}

	if err := h.vfs.UpdateFile(filePath, updateData.URL); err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to update file: "+err.Error())
		return
	}

	// Create response with updated file entry
	file := types.FileEntry{
		Path: filePath,
		URL:  updateData.URL,
	}

	h.sendSuccess(w, http.StatusOK, "File updated successfully", file)
}

// handleDeleteFile handles DELETE /api/files/{path} - delete a single file
func (h *APIHandler) handleDeleteFile(w http.ResponseWriter, r *http.Request, filePath string) {
	// Decode path parameter
	filePath = "/" + filePath
	filePath = path.Clean(filePath)

	if !h.vfs.Exists(filePath) {
		h.sendError(w, http.StatusNotFound, "File not found")
		return
	}

	if h.vfs.IsDir(filePath) {
		h.sendError(w, http.StatusBadRequest, "Cannot delete directory")
		return
	}

	if err := h.vfs.RemoveFile(filePath); err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to delete file: "+err.Error())
		return
	}

	h.sendSuccess(w, http.StatusOK, "File deleted successfully", map[string]string{"path": filePath})
}

// handleBulkOperation handles POST /api/files/bulk - bulk operations
func (h *APIHandler) handleBulkOperation(w http.ResponseWriter, r *http.Request) {
	var operation BulkOperation
	if err := json.NewDecoder(r.Body).Decode(&operation); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
		return
	}

	if operation.Operation != "add" && operation.Operation != "remove" {
		h.sendError(w, http.StatusBadRequest, "Invalid operation. Must be 'add' or 'remove'")
		return
	}

	results := make(map[string]interface{})
	successful := 0
	failed := 0
	errors := make(map[string]string)

	for _, file := range operation.Files {
		if err := h.validateFileEntry(file); err != nil {
			errors[file.Path] = err.Error()
			failed++
			continue
		}

		// Normalize path
		file.Path = path.Clean("/" + strings.TrimPrefix(file.Path, "/"))

		var err error
		switch operation.Operation {
		case "add":
			err = h.vfs.AddFile(file.Path, file.URL)
		case "remove":
			err = h.vfs.RemoveFile(file.Path)
		}

		if err != nil {
			errors[file.Path] = err.Error()
			failed++
		} else {
			successful++
		}
	}

	results["successful"] = successful
	results["failed"] = failed
	if len(errors) > 0 {
		results["errors"] = errors
	}

	message := fmt.Sprintf("Bulk %s operation completed: %d successful, %d failed",
		operation.Operation, successful, failed)

	if failed == 0 {
		h.sendSuccess(w, http.StatusOK, message, results)
	} else {
		h.sendSuccess(w, http.StatusPartialContent, message, results)
	}
}

// validateFileEntry validates a file entry
func (h *APIHandler) validateFileEntry(file types.FileEntry) error {
	if file.Path == "" {
		return fmt.Errorf("path is required")
	}
	if file.URL == "" {
		return fmt.Errorf("url is required")
	}
	if !strings.HasPrefix(file.URL, "http://") && !strings.HasPrefix(file.URL, "https://") {
		return fmt.Errorf("url must be a valid HTTP or HTTPS URL")
	}
	return nil
}

// sendSuccess sends a successful API response
func (h *APIHandler) sendSuccess(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.WriteHeader(statusCode)
	response := APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	json.NewEncoder(w).Encode(response)
}

// sendError sends an error API response
func (h *APIHandler) sendError(w http.ResponseWriter, statusCode int, errorMsg string) {
	w.WriteHeader(statusCode)
	response := APIResponse{
		Success: false,
		Error:   errorMsg,
	}
	json.NewEncoder(w).Encode(response)
}
