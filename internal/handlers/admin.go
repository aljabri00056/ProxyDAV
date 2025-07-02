package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"proxydav/internal/config"
	"proxydav/internal/filesystem"
	"proxydav/internal/storage"
	"proxydav/pkg/types"
)

type AdminHandler struct {
	vfs           *filesystem.VirtualFS
	store         *storage.PersistentStore
	config        *config.Config
	configUpdater config.ConfigUpdater
	template      *template.Template
}

// ServerController interface for restart/shutdown operations
type ServerController interface {
	Restart() error
	Shutdown() error
}

func NewAdminHandler(vfs *filesystem.VirtualFS, store *storage.PersistentStore, cfg *config.Config, configUpdater config.ConfigUpdater) *AdminHandler {
	tmpl := template.Must(template.New("admin").Funcs(template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"formatSize": func(size int64) string {
			const unit = 1024
			if size < unit {
				return fmt.Sprintf("%d B", size)
			}
			div, exp := int64(unit), 0
			for n := size / unit; n >= unit; n /= unit {
				div *= unit
				exp++
			}
			return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
		},
	}).Parse(adminTemplate))

	return &AdminHandler{
		vfs:           vfs,
		store:         store,
		config:        cfg,
		configUpdater: configUpdater,
		template:      tmpl,
	}
}

func (h *AdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/admin")

	switch {
	case path == "" || path == "/":
		h.handleDashboard(w, r)
	case path == "/config":
		h.handleConfig(w, r)
	case path == "/files":
		h.handleFiles(w, r)
	case path == "/import":
		h.handleImport(w, r)
	case path == "/export":
		h.handleExport(w, r)
	case path == "/api/config":
		h.handleConfigAPI(w, r)
	case path == "/api/files":
		h.handleFilesAPI(w, r)
	case path == "/api/import":
		h.handleImportAPI(w, r)
	case path == "/api/delete-file":
		h.handleDeleteFileAPI(w, r)
	case path == "/api/restart":
		h.handleRestartAPI(w, r)
	case path == "/api/shutdown":
		h.handleShutdownAPI(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *AdminHandler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	fileCount, _ := h.store.CountFileEntries()

	data := struct {
		Title     string
		FileCount int
		Config    *config.Config
		Section   string
	}{
		Title:     "ProxyDAV Admin Dashboard",
		FileCount: fileCount,
		Config:    h.config,
		Section:   "dashboard",
	}

	h.renderTemplate(w, "dashboard", data)
}

func (h *AdminHandler) handleConfig(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title   string
		Config  *config.Config
		Section string
	}{
		Title:   "Server Configuration",
		Config:  h.config,
		Section: "config",
	}

	h.renderTemplate(w, "config", data)
}

func (h *AdminHandler) handleFiles(w http.ResponseWriter, r *http.Request) {
	entries, _ := h.store.GetAllFileEntries()

	data := struct {
		Title   string
		Files   []types.FileEntry
		Section string
	}{
		Title:   "File Management",
		Files:   entries,
		Section: "files",
	}

	h.renderTemplate(w, "files", data)
}

func (h *AdminHandler) handleImport(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title   string
		Section string
	}{
		Title:   "Import Files",
		Section: "import",
	}

	h.renderTemplate(w, "import", data)
}

func (h *AdminHandler) handleExport(w http.ResponseWriter, r *http.Request) {
	entries, err := h.store.GetAllFileEntries()
	if err != nil {
		http.Error(w, "Failed to retrieve files", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=proxydav-export.json")

	exportData := struct {
		ExportDate time.Time         `json:"export_date"`
		FileCount  int               `json:"file_count"`
		Files      []types.FileEntry `json:"files"`
	}{
		ExportDate: time.Now(),
		FileCount:  len(entries),
		Files:      entries,
	}

	json.NewEncoder(w).Encode(exportData)
}

func (h *AdminHandler) handleConfigAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetConfig(w, r)
	case http.MethodPost:
		h.handleUpdateConfig(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *AdminHandler) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.config)
}

func (h *AdminHandler) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Create a copy of current config to validate changes
	newConfig := *h.config

	var errors []string

	if portStr := r.FormValue("port"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err != nil {
			errors = append(errors, "Invalid port number")
		} else if port < 1 || port > 65535 {
			errors = append(errors, "Port must be between 1 and 65535")
		} else {
			newConfig.Port = port
		}
	}

	if dataDir := r.FormValue("data_dir"); dataDir != "" {
		if dataDir == "" {
			errors = append(errors, "Data directory cannot be empty")
		} else {
			newConfig.DataDir = dataDir
		}
	}

	newConfig.UseRedirect = r.FormValue("use_redirect") == "on"
	newConfig.AuthEnabled = r.FormValue("auth_enabled") == "on"

	if newConfig.AuthEnabled {
		if authUser := r.FormValue("auth_user"); authUser != "" {
			newConfig.AuthUser = authUser
		}
		if authPass := r.FormValue("auth_pass"); authPass != "" && authPass != "••••••••" {
			newConfig.AuthPass = authPass
		}

		if newConfig.AuthUser == "" || newConfig.AuthPass == "" {
			errors = append(errors, "Authentication requires both username and password")
		}
	}

	if err := newConfig.Validate(); err != nil {
		errors = append(errors, err.Error())
	}

	w.Header().Set("Content-Type", "text/html")

	if len(errors) > 0 {
		response := `<div class="alert alert-danger" role="alert">
			<strong>Error:</strong> Configuration validation failed:<ul>`
		for _, err := range errors {
			response += fmt.Sprintf("<li>%s</li>", err)
		}
		response += `</ul></div>`
		w.Write([]byte(response))
		return
	}

	// Apply configuration changes dynamically
	// Store original values for comparison
	originalPort := h.config.Port
	originalDataDir := h.config.DataDir

	var response string
	if err := h.configUpdater.UpdateConfig(&newConfig); err != nil {
		response = fmt.Sprintf(`<div class="alert alert-danger" role="alert">
			<strong>Error:</strong> Failed to apply configuration changes: %s
		</div>`, err.Error())
	} else {
		// Update local config reference
		h.config = h.configUpdater.GetConfig()

		// Determine what requires restart
		needsRestart := []string{}
		if originalPort != newConfig.Port {
			needsRestart = append(needsRestart, "Port change")
		}
		if originalDataDir != newConfig.DataDir {
			needsRestart = append(needsRestart, "Data directory change")
		}

		if len(needsRestart) > 0 {
			response = fmt.Sprintf(`<div class="alert alert-warning" role="alert">
				<i class="fas fa-exclamation-triangle me-2"></i>
				<strong>Configuration Updated:</strong> Most changes applied successfully!<br>
				<strong>Restart required for:</strong> %s
			</div>`, strings.Join(needsRestart, ", "))
		} else {
			response = `<div class="alert alert-success" role="alert">
				<i class="fas fa-check-circle me-2"></i>
				<strong>Configuration Updated:</strong> All changes applied successfully and are now active!
			</div>`
		}
	}

	w.Write([]byte(response))
}

func (h *AdminHandler) handleFilesAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.handleAddFile(w, r)
	case http.MethodGet:
		h.handleListFiles(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *AdminHandler) handleAddFile(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	path := r.FormValue("path")
	url := r.FormValue("url")

	if path == "" || url == "" {
		http.Error(w, "Path and URL are required", http.StatusBadRequest)
		return
	}

	entry := types.FileEntry{Path: path, URL: url}

	if err := h.store.SetFileEntry(&entry); err != nil {
		http.Error(w, "Failed to add file", http.StatusInternalServerError)
		return
	}

	entries, _ := h.store.GetAllFileEntries()
	h.renderFileList(w, entries)
}

func (h *AdminHandler) handleListFiles(w http.ResponseWriter, r *http.Request) {
	entries, _ := h.store.GetAllFileEntries()
	h.renderFileList(w, entries)
}

func (h *AdminHandler) handleDeleteFileAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path parameter required", http.StatusBadRequest)
		return
	}

	if err := h.store.DeleteFileEntry(path); err != nil {
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	entries, _ := h.store.GetAllFileEntries()
	h.renderFileList(w, entries)
}

func (h *AdminHandler) handleImportAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("import_file")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	var importData struct {
		Files []types.FileEntry `json:"files"`
	}

	if err := json.NewDecoder(file).Decode(&importData); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	successCount := 0
	for _, entry := range importData.Files {
		if err := h.store.SetFileEntry(&entry); err == nil {
			successCount++
		}
	}

	response := fmt.Sprintf(`<div class="alert alert-success" role="alert">
		<strong>Success:</strong> Imported %d of %d files successfully.
	</div>`, successCount, len(importData.Files))

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(response))
}

func (h *AdminHandler) renderFileList(w http.ResponseWriter, files []types.FileEntry) {
	fileListTemplate := `
	{{range .}}
	<tr>
		<td class="path-cell">{{.Path}}</td>
		<td class="url-cell">
			<a href="{{.URL}}" target="_blank" class="url-link">{{.URL}}</a>
		</td>
		<td>
			<button class="btn btn-outline-danger btn-sm" 
					hx-delete="/admin/api/delete-file?path={{.Path}}" 
					hx-target="#file-list"
					hx-confirm="Are you sure you want to delete this file?"
					onclick="this.disabled=true">
				<i class="fas fa-trash"></i>
			</button>
		</td>
	</tr>
	{{else}}
	<tr>
		<td colspan="3" class="text-center text-muted">No files configured</td>
	</tr>
	{{end}}`

	tmpl := template.Must(template.New("filelist").Parse(fileListTemplate))
	tmpl.Execute(w, files)
}

func (h *AdminHandler) renderTemplate(w http.ResponseWriter, section string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.template.Execute(w, data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (h *AdminHandler) handleRestartAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get server controller for restart
	serverController := h.configUpdater.(ServerController)
	if err := serverController.Restart(); err != nil {
		response := fmt.Sprintf(`<div class="alert alert-danger" role="alert">
			<i class="fas fa-exclamation-triangle me-2"></i>
			<strong>Restart Failed:</strong> %s
		</div>`, err.Error())
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(response))
		return
	}

	response := `<div class="alert alert-success" role="alert">
		<i class="fas fa-sync me-2"></i>
		<strong>Restart Initiated:</strong> Server is restarting with the new configuration...
		<div class="mt-2">
			<div class="spinner-border spinner-border-sm me-2" role="status">
				<span class="visually-hidden">Loading...</span>
			</div>
			This page will automatically reload once the server is back online.
		</div>
	</div>
	<script>
		setTimeout(function() {
			// Reload the page after a short delay to allow restart
			window.location.reload();
		}, 3000);
	</script>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(response))
}

func (h *AdminHandler) handleShutdownAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Get server controller for shutdown
	serverController := h.configUpdater.(ServerController)
	if err := serverController.Shutdown(); err != nil {
		response := fmt.Sprintf(`<div class="alert alert-danger" role="alert">
			<i class="fas fa-exclamation-triangle me-2"></i>
			<strong>Shutdown Failed:</strong> %s
		</div>`, err.Error())
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(response))
		return
	}

	response := `<div class="alert alert-warning" role="alert">
		<i class="fas fa-power-off me-2"></i>
		<strong>Shutdown Initiated:</strong> Server is shutting down gracefully...
		<div class="mt-2">
			<div class="spinner-border spinner-border-sm me-2" role="status">
				<span class="visually-hidden">Loading...</span>
			</div>
			The server will stop in a few seconds.
		</div>
	</div>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(response))
}
