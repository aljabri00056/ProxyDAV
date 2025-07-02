package handlers

import (
	"context"
	"encoding/xml"
	"io"
	"log"
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"proxydav/internal/filesystem"
	"proxydav/internal/storage"
	"proxydav/internal/webdav"
	"proxydav/pkg/types"
)

type WebDAVHandler struct {
	vfs         *filesystem.VirtualFS
	store       *storage.PersistentStore
	useRedirect bool
	client      *http.Client
}

func NewWebDAVHandler(vfs *filesystem.VirtualFS, store *storage.PersistentStore, useRedirect bool) *WebDAVHandler {
	return &WebDAVHandler{
		vfs:         vfs,
		store:       store,
		useRedirect: useRedirect,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (h *WebDAVHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "OPTIONS":
		h.handleOptions(w, r)
	case "PROPFIND":
		h.handlePropFind(w, r)
	case "GET", "HEAD":
		h.handleGetHead(w, r)
	case "DELETE":
		h.handleDelete(w, r)
	case "MOVE":
		h.handleMove(w, r)
	case "COPY":
		h.handleCopy(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *WebDAVHandler) handleOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "OPTIONS, PROPFIND, GET, HEAD, DELETE, MOVE, COPY")
	w.Header().Set("DAV", "1")
	w.Header().Set("MS-Author-Via", "DAV")
	w.WriteHeader(http.StatusOK)
}

func (h *WebDAVHandler) handlePropFind(w http.ResponseWriter, r *http.Request) {
	requestPath := r.URL.Path
	normalizedPath := path.Clean("/" + strings.TrimPrefix(requestPath, "/"))
	if !h.vfs.Exists(normalizedPath) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	depth := r.Header.Get("Depth")
	if depth == "" {
		depth = "1"
	}

	var responses []webdav.Response

	// Add response for the requested path itself
	if response := h.createResponse(normalizedPath); response != nil {
		responses = append(responses, *response)
	}

	// If it's a directory and depth allows, add children
	if depth != "0" && h.vfs.IsDir(normalizedPath) {
		children := h.vfs.ListDir(normalizedPath)
		for _, child := range children {
			if response := h.createResponse(child.Path); response != nil {
				responses = append(responses, *response)
			}
		}
	}

	multistatus := webdav.Multistatus{
		Responses: responses,
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusMultiStatus)

	xmlData, err := xml.MarshalIndent(multistatus, "", "  ")
	if err != nil {
		log.Printf("Error marshaling XML: %v", err)
		return
	}

	w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>` + "\n"))
	w.Write(xmlData)
}

// createResponse creates a WebDAV response for a given path
func (h *WebDAVHandler) createResponse(requestPath string) *webdav.Response {
	item, exists := h.vfs.GetItem(requestPath)
	if !exists && !h.vfs.IsDir(requestPath) {
		return nil
	}

	// For WebDAV compatibility, directories should have trailing slashes in href
	href := requestPath
	if h.vfs.IsDir(requestPath) && !strings.HasSuffix(href, "/") && href != "/" {
		href += "/"
	}

	response := &webdav.Response{
		Href: href,
		Propstat: webdav.Propstat{
			Status: "HTTP/1.1 200 OK",
		},
	}

	if item != nil && !item.IsDir {
		// It's a file
		response.Propstat.Prop = webdav.Prop{
			DisplayName:  item.Name,
			ResourceType: nil, // Files don't have resource type
			ContentType:  mime.TypeByExtension(filepath.Ext(item.Name)),
		}

		// Try to get metadata from persistent store or fetch it
		metadata := h.getFileMetadata(item.URL)
		if metadata != nil {
			response.Propstat.Prop.ContentLength = &metadata.Size
			response.Propstat.Prop.LastModified = webdav.FormatTime(metadata.LastModified)
			response.Propstat.Prop.ETag = webdav.GenerateETag(metadata.URL, metadata.LastModified)
		}
	} else {
		// It's a directory
		displayName := path.Base(requestPath)
		if displayName == "/" || displayName == "." {
			displayName = "Root"
		}

		response.Propstat.Prop = webdav.Prop{
			DisplayName: displayName,
			ResourceType: &webdav.ResourceType{
				Collection: &webdav.Collection{},
			},
		}
	}

	return response
}

// getFileMetadata gets file metadata from persistent store or by making a HEAD request
func (h *WebDAVHandler) getFileMetadata(url string) *types.FileMetadata {
	// Try persistent store first
	if metadata, err := h.store.GetFileMetadata(url); err == nil && metadata != nil {
		return metadata
	}

	// Make HEAD request to get metadata
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		log.Printf("Error creating HEAD request for %s: %v", url, err)
		return nil
	}

	resp, err := h.client.Do(req)
	if err != nil {
		log.Printf("Error making HEAD request for %s: %v", url, err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("HEAD request for %s returned status %d", url, resp.StatusCode)
		return nil
	}

	// Parse metadata
	metadata := &types.FileMetadata{
		URL: url,
	}

	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		if size, err := strconv.ParseInt(contentLength, 10, 64); err == nil {
			metadata.Size = size
		}
	}

	if lastModified := resp.Header.Get("Last-Modified"); lastModified != "" {
		if t, err := time.Parse(time.RFC1123, lastModified); err == nil {
			metadata.LastModified = t
		} else {
			metadata.LastModified = time.Now()
		}
	} else {
		metadata.LastModified = time.Now()
	}

	// Store the metadata persistently
	if err := h.store.SetFileMetadata(metadata); err != nil {
		log.Printf("Failed to store metadata for %s: %v", url, err)
	}

	return metadata
}

// handleGetHead handles GET and HEAD requests
func (h *WebDAVHandler) handleGetHead(w http.ResponseWriter, r *http.Request) {
	requestPath := r.URL.Path

	normalizedPath := path.Clean("/" + strings.TrimPrefix(requestPath, "/"))
	item, exists := h.vfs.GetItem(normalizedPath)
	if !exists {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if item.IsDir {
		http.Error(w, "Cannot GET directory", http.StatusBadRequest)
		return
	}

	if h.useRedirect {
		http.Redirect(w, r, item.URL, http.StatusFound)
		return
	}

	// Proxy the content
	h.proxyContent(w, r, item.URL)
}

// proxyContent proxies content from the remote URL
func (h *WebDAVHandler) proxyContent(w http.ResponseWriter, r *http.Request, url string) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, r.Method, url, nil)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Copy relevant headers
	for name, values := range r.Header {
		if name == "Host" || strings.HasPrefix(name, "X-") {
			continue
		}
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	resp, err := h.client.Do(req)
	if err != nil {
		log.Printf("Error proxying request to %s: %v", url, err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	if r.Method != "HEAD" {
		_, err := io.Copy(w, resp.Body)
		if err != nil {
			log.Printf("Error copying response body: %v", err)
		}
	}
}

func (h *WebDAVHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	requestPath := r.URL.Path
	normalizedPath := path.Clean("/" + strings.TrimPrefix(requestPath, "/"))

	if !h.vfs.Exists(normalizedPath) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	var err error
	if h.vfs.IsDir(normalizedPath) {
		err = h.vfs.RemoveDirectory(normalizedPath)
	} else {
		err = h.vfs.RemoveFile(normalizedPath)
	}

	if err != nil {
		log.Printf("Error deleting %s: %v", normalizedPath, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *WebDAVHandler) handleMove(w http.ResponseWriter, r *http.Request) {
	sourcePath := r.URL.Path
	normalizedSource := path.Clean("/" + strings.TrimPrefix(sourcePath, "/"))

	destination := r.Header.Get("Destination")
	if destination == "" {
		http.Error(w, "Missing Destination header", http.StatusBadRequest)
		return
	}

	destPath, err := h.parseDestinationPath(destination)
	if err != nil {
		log.Printf("Error parsing destination %s: %v", destination, err)
		http.Error(w, "Bad Destination", http.StatusBadRequest)
		return
	}

	normalizedDest := path.Clean("/" + strings.TrimPrefix(destPath, "/"))

	if !h.vfs.Exists(normalizedSource) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	overwrite := r.Header.Get("Overwrite")
	if overwrite == "" {
		overwrite = "T" // Default is to overwrite
	}

	destExists := h.vfs.Exists(normalizedDest)
	if destExists && overwrite == "F" {
		http.Error(w, "Destination exists and overwrite is forbidden", http.StatusPreconditionFailed)
		return
	}

	if destExists && overwrite == "T" {
		var deleteErr error
		if h.vfs.IsDir(normalizedDest) {
			deleteErr = h.vfs.RemoveDirectory(normalizedDest)
		} else {
			deleteErr = h.vfs.RemoveFile(normalizedDest)
		}
		if deleteErr != nil {
			log.Printf("Error deleting destination %s: %v", normalizedDest, deleteErr)
			http.Error(w, "Failed to overwrite destination", http.StatusInternalServerError)
			return
		}
	}

	var moveErr error
	if h.vfs.IsDir(normalizedSource) {
		moveErr = h.vfs.MoveDirectory(normalizedSource, normalizedDest)
	} else {
		moveErr = h.vfs.MoveFile(normalizedSource, normalizedDest)
	}

	if moveErr != nil {
		log.Printf("Error moving %s to %s: %v", normalizedSource, normalizedDest, moveErr)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if destExists {
		w.WriteHeader(http.StatusNoContent) // Replaced existing resource
	} else {
		w.WriteHeader(http.StatusCreated) // Created new resource
	}
}

func (h *WebDAVHandler) handleCopy(w http.ResponseWriter, r *http.Request) {
	sourcePath := r.URL.Path
	normalizedSource := path.Clean("/" + strings.TrimPrefix(sourcePath, "/"))

	destination := r.Header.Get("Destination")
	if destination == "" {
		http.Error(w, "Missing Destination header", http.StatusBadRequest)
		return
	}

	destPath, err := h.parseDestinationPath(destination)
	if err != nil {
		log.Printf("Error parsing destination %s: %v", destination, err)
		http.Error(w, "Bad Destination", http.StatusBadRequest)
		return
	}

	normalizedDest := path.Clean("/" + strings.TrimPrefix(destPath, "/"))

	if !h.vfs.Exists(normalizedSource) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	overwrite := r.Header.Get("Overwrite")
	if overwrite == "" {
		overwrite = "T" // Default is to overwrite
	}

	destExists := h.vfs.Exists(normalizedDest)
	if destExists && overwrite == "F" {
		http.Error(w, "Destination exists and overwrite is forbidden", http.StatusPreconditionFailed)
		return
	}

	if destExists && overwrite == "T" {
		var deleteErr error
		if h.vfs.IsDir(normalizedDest) {
			deleteErr = h.vfs.RemoveDirectory(normalizedDest)
		} else {
			deleteErr = h.vfs.RemoveFile(normalizedDest)
		}
		if deleteErr != nil {
			log.Printf("Error deleting destination %s: %v", normalizedDest, deleteErr)
			http.Error(w, "Failed to overwrite destination", http.StatusInternalServerError)
			return
		}
	}

	var copyErr error
	if h.vfs.IsDir(normalizedSource) {
		copyErr = h.vfs.CopyDirectory(normalizedSource, normalizedDest)
	} else {
		copyErr = h.vfs.CopyFile(normalizedSource, normalizedDest)
	}

	if copyErr != nil {
		log.Printf("Error copying %s to %s: %v", normalizedSource, normalizedDest, copyErr)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if destExists {
		w.WriteHeader(http.StatusNoContent) // Replaced existing resource
	} else {
		w.WriteHeader(http.StatusCreated) // Created new resource
	}
}

func (h *WebDAVHandler) parseDestinationPath(destination string) (string, error) {
	// The destination can be a full URL or just a path
	// We need to extract just the path part

	// If it starts with http:// or https://, parse as URL
	if strings.HasPrefix(destination, "http://") || strings.HasPrefix(destination, "https://") {
		// Find the path part after the host
		parts := strings.SplitN(destination, "/", 4)
		if len(parts) < 4 {
			return "/", nil
		}
		return "/" + parts[3], nil
	}

	// Otherwise, treat as path
	return destination, nil
}
