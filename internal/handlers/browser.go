package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strings"

	"proxydav/internal/filesystem"
)

const htmlTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>ProxyDAV - {{.Path}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f7;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 12px;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px 30px;
        }
        .header h1 {
            margin: 0;
            font-size: 24px;
            font-weight: 600;
        }
        .breadcrumb {
            background: #f8f9fa;
            padding: 15px 30px;
            border-bottom: 1px solid #e9ecef;
        }
        .breadcrumb a {
            color: #0066cc;
            text-decoration: none;
            margin-right: 5px;
        }
        .breadcrumb a:hover {
            text-decoration: underline;
        }
        .file-list {
            padding: 0;
        }
        .file-item {
            display: flex;
            align-items: center;
            padding: 12px 30px;
            border-bottom: 1px solid #f0f0f0;
            transition: background-color 0.2s;
        }
        .file-item:hover {
            background-color: #f8f9fa;
        }
        .file-item:last-child {
            border-bottom: none;
        }
        .file-icon {
            width: 24px;
            height: 24px;
            margin-right: 12px;
            flex-shrink: 0;
        }
        .file-name {
            flex: 1;
            font-weight: 500;
        }
        .file-name a {
            color: #333;
            text-decoration: none;
        }
        .file-name a:hover {
            color: #0066cc;
        }
        .directory {
            color: #0066cc;
        }
        .file-size {
            color: #666;
            font-size: 14px;
            min-width: 80px;
            text-align: right;
        }
        .empty-state {
            text-align: center;
            padding: 60px 30px;
            color: #666;
        }
        .footer {
            background: #f8f9fa;
            padding: 20px 30px;
            text-align: center;
            color: #666;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📁 ProxyDAV Server</h1>
        </div>
        
        <div class="breadcrumb">
            {{range .Breadcrumbs}}
                <a href="{{.URL}}">{{.Name}}</a> /
            {{end}}
        </div>
        
        <div class="file-list">
            {{if .Items}}
                {{range .Items}}
                <div class="file-item">
                    <div class="file-icon">
                        {{if .IsDir}}📁{{else}}📄{{end}}
                    </div>
                    <div class="file-name {{if .IsDir}}directory{{end}}">
                        <a href="{{.Path}}">{{.Name}}</a>
                    </div>
                    <div class="file-size">
                        {{if not .IsDir}}{{.Size}}{{end}}
                    </div>
                </div>
                {{end}}
            {{else}}
                <div class="empty-state">
                    <h3>📂 Empty Directory</h3>
                    <p>This directory contains no files or subdirectories.</p>
                </div>
            {{end}}
        </div>
        
        <div class="footer">
            Powered by ProxyDAV - Virtual WebDAV Server
        </div>
    </div>
</body>
</html>`

// BrowserHandler handles browser requests for directory listing
type BrowserHandler struct {
	vfs      *filesystem.VirtualFS
	template *template.Template
}

// NewBrowserHandler creates a new browser handler
func NewBrowserHandler(vfs *filesystem.VirtualFS) *BrowserHandler {
	tmpl := template.Must(template.New("directory").Parse(htmlTemplate))
	return &BrowserHandler{
		vfs:      vfs,
		template: tmpl,
	}
}

// ServeHTTP handles browser requests
func (h *BrowserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestPath := r.URL.Path

	if !h.vfs.Exists(requestPath) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	item, exists := h.vfs.GetItem(requestPath)
	if exists && !item.IsDir {
		// Redirect to the actual file URL for browser viewing
		http.Redirect(w, r, item.URL, http.StatusFound)
		return
	}

	// Generate directory listing
	h.renderDirectoryListing(w, requestPath)
}

// BreadcrumbItem represents a breadcrumb item
type BreadcrumbItem struct {
	Name string
	URL  string
}

// TemplateData represents data for the HTML template
type TemplateData struct {
	Path        string
	Breadcrumbs []BreadcrumbItem
	Items       []TemplateItem
}

// TemplateItem represents an item in the directory listing
type TemplateItem struct {
	Name  string
	Path  string
	IsDir bool
	Size  string
}

// renderDirectoryListing renders the directory listing HTML
func (h *BrowserHandler) renderDirectoryListing(w http.ResponseWriter, requestPath string) {
	// Generate breadcrumbs
	breadcrumbs := h.generateBreadcrumbs(requestPath)

	// Get directory contents
	items := h.vfs.ListDir(requestPath)
	templateItems := make([]TemplateItem, len(items))

	for i, item := range items {
		templateItems[i] = TemplateItem{
			Name:  item.Name,
			Path:  item.Path,
			IsDir: item.IsDir,
			Size:  h.formatSize(0), // Size will be fetched if needed
		}
	}

	data := TemplateData{
		Path:        requestPath,
		Breadcrumbs: breadcrumbs,
		Items:       templateItems,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.template.Execute(w, data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// generateBreadcrumbs generates breadcrumb navigation
func (h *BrowserHandler) generateBreadcrumbs(requestPath string) []BreadcrumbItem {
	var breadcrumbs []BreadcrumbItem

	// Add root
	breadcrumbs = append(breadcrumbs, BreadcrumbItem{
		Name: "Home",
		URL:  "/",
	})

	if requestPath == "/" {
		return breadcrumbs
	}

	// Split path and create breadcrumbs
	parts := strings.Split(strings.Trim(requestPath, "/"), "/")
	currentPath := ""

	for _, part := range parts {
		if part == "" {
			continue
		}
		currentPath = path.Join(currentPath, part)
		breadcrumbs = append(breadcrumbs, BreadcrumbItem{
			Name: part,
			URL:  "/" + currentPath,
		})
	}

	return breadcrumbs
}

// formatSize formats file size for display
func (h *BrowserHandler) formatSize(size int64) string {
	if size == 0 {
		return ""
	}

	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(size)/float64(div), units[exp])
}
