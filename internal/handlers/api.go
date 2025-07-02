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

type APIHandler struct {
	vfs *filesystem.VirtualFS
}

func NewAPIHandler(vfs *filesystem.VirtualFS) *APIHandler {
	return &APIHandler{
		vfs: vfs,
	}
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type AddFilesRequest struct {
	Files []types.FileEntry `json:"files"`
}

type DeleteFilesRequest struct {
	Files []types.FileEntry `json:"files"`
}

type FileListResponse struct {
	Files []types.FileEntry `json:"files"`
	Total int               `json:"total"`
}

func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse the path to determine the operation
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 || pathParts[0] != "api" || pathParts[1] != "files" {
		h.sendError(w, http.StatusNotFound, "Invalid API endpoint")
		return
	}

	if len(pathParts) != 2 {
		h.sendError(w, http.StatusNotFound, "Invalid API endpoint")
		return
	}

	switch r.Method {
	case "GET":
		h.handleListFiles(w, r)
	case "POST":
		h.handleAddFiles(w, r)
	case "DELETE":
		h.handleDeleteFiles(w, r)
	default:
		h.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// GET /api/files - list all files
func (h *APIHandler) handleListFiles(w http.ResponseWriter, r *http.Request) {
	files := h.vfs.GetAllFiles()

	response := FileListResponse{
		Files: files,
		Total: len(files),
	}

	h.sendSuccess(w, http.StatusOK, "Files retrieved successfully", response)
}

// POST /api/files - add multiple files
func (h *APIHandler) handleAddFiles(w http.ResponseWriter, r *http.Request) {
	var request AddFilesRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
		return
	}

	if len(request.Files) == 0 {
		h.sendError(w, http.StatusBadRequest, "files array cannot be empty")
		return
	}

	results := make(map[string]interface{})
	successful := 0
	failed := 0
	errors := make(map[string]string)
	successfulFiles := make([]types.FileEntry, 0)

	for _, file := range request.Files {
		if err := h.validateFileEntry(file); err != nil {
			errors[file.Path] = err.Error()
			failed++
			continue
		}

		file.Path = path.Clean("/" + strings.TrimPrefix(file.Path, "/"))

		if err := h.vfs.AddFile(file.Path, file.URL); err != nil {
			errors[file.Path] = err.Error()
			failed++
		} else {
			successful++
			successfulFiles = append(successfulFiles, file)
		}
	}

	results["successful"] = successful
	results["failed"] = failed
	results["files"] = successfulFiles
	if len(errors) > 0 {
		results["errors"] = errors
	}

	message := fmt.Sprintf("Add operation completed: %d successful, %d failed", successful, failed)

	if failed == 0 {
		h.sendSuccess(w, http.StatusCreated, message, results)
	} else {
		h.sendSuccess(w, http.StatusPartialContent, message, results)
	}
}

// DELETE /api/files - delete multiple files
func (h *APIHandler) handleDeleteFiles(w http.ResponseWriter, r *http.Request) {
	var request DeleteFilesRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
		return
	}

	if len(request.Files) == 0 {
		h.sendError(w, http.StatusBadRequest, "files array cannot be empty")
		return
	}

	results := make(map[string]interface{})
	successful := 0
	failed := 0
	errors := make(map[string]string)
	deletedFiles := make([]types.FileEntry, 0)

	for _, file := range request.Files {
		if file.Path == "" {
			errors[file.Path] = "path is required"
			failed++
			continue
		}

		filePath := path.Clean("/" + strings.TrimPrefix(file.Path, "/"))

		if !h.vfs.Exists(filePath) {
			errors[filePath] = "File not found"
			failed++
			continue
		}

		if h.vfs.IsDir(filePath) {
			errors[filePath] = "Cannot delete directory"
			failed++
			continue
		}

		if err := h.vfs.RemoveFile(filePath); err != nil {
			errors[filePath] = err.Error()
			failed++
		} else {
			successful++
			deletedFiles = append(deletedFiles, types.FileEntry{Path: filePath, URL: file.URL})
		}
	}

	results["successful"] = successful
	results["failed"] = failed
	results["files"] = deletedFiles
	if len(errors) > 0 {
		results["errors"] = errors
	}

	message := fmt.Sprintf("Delete operation completed: %d successful, %d failed", successful, failed)

	if failed == 0 {
		h.sendSuccess(w, http.StatusOK, message, results)
	} else {
		h.sendSuccess(w, http.StatusPartialContent, message, results)
	}
}

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

func (h *APIHandler) sendSuccess(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.WriteHeader(statusCode)
	response := APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	json.NewEncoder(w).Encode(response)
}

func (h *APIHandler) sendError(w http.ResponseWriter, statusCode int, errorMsg string) {
	w.WriteHeader(statusCode)
	response := APIResponse{
		Success: false,
		Error:   errorMsg,
	}
	json.NewEncoder(w).Encode(response)
}
