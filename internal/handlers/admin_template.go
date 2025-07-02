package handlers

const adminTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - ProxyDAV Admin</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <style>
        :root {
            --primary-color: #2563eb;
            --secondary-color: #64748b;
            --success-color: #059669;
            --warning-color: #d97706;
            --danger-color: #dc2626;
            --dark-color: #1e293b;
            --light-bg: #f8fafc;
            --border-color: #e2e8f0;
        }
        
        body {
            background-color: var(--light-bg);
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
        }
        
        .sidebar {
            background: linear-gradient(135deg, var(--dark-color) 0%, #334155 100%);
            min-height: 100vh;
            box-shadow: 2px 0 10px rgba(0,0,0,0.1);
        }
        
        .sidebar .nav-link {
            color: #cbd5e1;
            padding: 12px 20px;
            border-radius: 8px;
            margin: 4px 12px;
            transition: all 0.3s ease;
        }
        
        .sidebar .nav-link:hover {
            background-color: rgba(255,255,255,0.1);
            color: white;
            transform: translateX(4px);
        }
        
        .sidebar .nav-link.active {
            background-color: var(--primary-color);
            color: white;
            box-shadow: 0 4px 12px rgba(37, 99, 235, 0.3);
        }
        
        .main-content {
            padding: 2rem;
        }
        
        .card {
            border: none;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            border-radius: 12px;
            transition: transform 0.2s ease, box-shadow 0.2s ease;
        }
        
        .card:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 20px rgba(0,0,0,0.15);
        }
        
        .card-header {
            background: linear-gradient(135deg, var(--primary-color) 0%, #3b82f6 100%);
            color: white;
            border-radius: 12px 12px 0 0 !important;
            border: none;
        }
        
        .stat-card {
            background: linear-gradient(135deg, #fff 0%, #f8fafc 100%);
            border-left: 4px solid var(--primary-color);
        }
        
        .btn-primary {
            background: linear-gradient(135deg, var(--primary-color) 0%, #3b82f6 100%);
            border: none;
            border-radius: 8px;
            padding: 10px 20px;
            font-weight: 500;
            transition: all 0.3s ease;
        }
        
        .btn-primary:hover {
            transform: translateY(-1px);
            box-shadow: 0 4px 12px rgba(37, 99, 235, 0.3);
        }
        
        .form-control, .form-select {
            border-radius: 8px;
            border: 1px solid var(--border-color);
            padding: 12px 16px;
            transition: all 0.3s ease;
        }
        
        .form-control:focus, .form-select:focus {
            border-color: var(--primary-color);
            box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
        }
        
        .table {
            background: white;
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        
        .table thead th {
            background-color: var(--light-bg);
            border: none;
            font-weight: 600;
            color: var(--dark-color);
            padding: 16px;
        }
        
        .table tbody td {
            padding: 16px;
            vertical-align: middle;
            border-color: var(--border-color);
        }
        
        .path-cell {
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            background-color: #f1f5f9;
            padding: 8px 12px;
            border-radius: 6px;
            font-size: 0.9em;
        }
        
        .url-cell {
            max-width: 300px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }
        
        .url-link {
            color: var(--primary-color);
            text-decoration: none;
            transition: color 0.3s ease;
        }
        
        .url-link:hover {
            color: #1d4ed8;
            text-decoration: underline;
        }
        
        .alert {
            border: none;
            border-radius: 8px;
            padding: 16px 20px;
        }
        
        .logo {
            color: white;
            font-size: 1.5rem;
            font-weight: 700;
            margin-bottom: 2rem;
            text-align: center;
            padding: 20px;
        }
        
        .loading {
            opacity: 0.6;
            pointer-events: none;
        }
        
        .htmx-request .loading-spinner {
            display: inline-block;
        }
        
        .loading-spinner {
            display: none;
            margin-left: 8px;
        }
        
        @media (max-width: 768px) {
            .sidebar {
                min-height: auto;
            }
            
            .main-content {
                padding: 1rem;
            }
            
            .url-cell {
                max-width: 200px;
            }
        }
    </style>
</head>
<body>
    <div class="container-fluid">
        <div class="row">
            <!-- Sidebar -->
            <div class="col-md-3 col-lg-2 sidebar">
                <div class="logo">
                    <i class="fas fa-server"></i> ProxyDAV
                </div>
                <nav class="nav flex-column">
                    <a class="nav-link {{if eq .Section "dashboard"}}active{{end}}" href="/admin/">
                        <i class="fas fa-tachometer-alt me-2"></i> Dashboard
                    </a>
                    <a class="nav-link {{if eq .Section "config"}}active{{end}}" href="/admin/config">
                        <i class="fas fa-cog me-2"></i> Configuration
                    </a>
                    <a class="nav-link {{if eq .Section "files"}}active{{end}}" href="/admin/files">
                        <i class="fas fa-file-alt me-2"></i> File Management
                    </a>
                    <a class="nav-link {{if eq .Section "import"}}active{{end}}" href="/admin/import">
                        <i class="fas fa-upload me-2"></i> Import/Export
                    </a>
                </nav>
            </div>
            
            <!-- Main Content -->
            <div class="col-md-9 col-lg-10 main-content">
                {{if eq .Section "dashboard"}}
                    {{template "dashboard" .}}
                {{else if eq .Section "config"}}
                    {{template "config" .}}
                {{else if eq .Section "files"}}
                    {{template "files" .}}
                {{else if eq .Section "import"}}
                    {{template "import" .}}
                {{else}}
                    {{template "dashboard" .}}
                {{end}}
            </div>
        </div>
    </div>
    
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
    <script>
        // Add loading states for HTMX requests
        document.body.addEventListener('htmx:beforeRequest', function(evt) {
            evt.detail.elt.classList.add('loading');
        });
        
        document.body.addEventListener('htmx:afterRequest', function(evt) {
            evt.detail.elt.classList.remove('loading');
        });
        
        // Auto-hide alerts after 5 seconds
        setTimeout(function() {
            var alerts = document.querySelectorAll('.alert');
            alerts.forEach(function(alert) {
                if (alert.classList.contains('alert-success') || alert.classList.contains('alert-info')) {
                    var bsAlert = new bootstrap.Alert(alert);
                    bsAlert.close();
                }
            });
        }, 5000);
    </script>
</body>
</html>

{{define "dashboard"}}
<div class="d-flex justify-content-between align-items-center mb-4">
    <h1 class="h3 mb-0">
        <i class="fas fa-tachometer-alt text-primary me-2"></i>Dashboard
    </h1>
    <div class="text-muted">
        <i class="fas fa-clock me-1"></i>
        <span id="current-time"></span>
    </div>
</div>

<div class="row mb-4">
    <div class="col-md-3 mb-3">
        <div class="card stat-card">
            <div class="card-body">
                <div class="d-flex align-items-center">
                    <div class="flex-grow-1">
                        <h6 class="card-subtitle mb-2 text-muted">Total Files</h6>
                        <h3 class="card-title mb-0">{{.FileCount}}</h3>
                    </div>
                    <div class="text-primary">
                        <i class="fas fa-file-alt fa-2x"></i>
                    </div>
                </div>
            </div>
        </div>
    </div>
    <div class="col-md-3 mb-3">
        <div class="card stat-card">
            <div class="card-body">
                <div class="d-flex align-items-center">
                    <div class="flex-grow-1">
                        <h6 class="card-subtitle mb-2 text-muted">Server Port</h6>
                        <h3 class="card-title mb-0">{{.Config.Port}}</h3>
                    </div>
                    <div class="text-success">
                        <i class="fas fa-network-wired fa-2x"></i>
                    </div>
                </div>
            </div>
        </div>
    </div>
    <div class="col-md-3 mb-3">
        <div class="card stat-card">
            <div class="card-body">
                <div class="d-flex align-items-center">
                    <div class="flex-grow-1">
                        <h6 class="card-subtitle mb-2 text-muted">Authentication</h6>
                        <h3 class="card-title mb-0">
                            {{if .Config.AuthEnabled}}
                                <span class="badge bg-success">Enabled</span>
                            {{else}}
                                <span class="badge bg-warning">Disabled</span>
                            {{end}}
                        </h3>
                    </div>
                    <div class="{{if .Config.AuthEnabled}}text-success{{else}}text-warning{{end}}">
                        <i class="fas fa-shield-alt fa-2x"></i>
                    </div>
                </div>
            </div>
        </div>
    </div>
    <div class="col-md-3 mb-3">
        <div class="card stat-card">
            <div class="card-body">
                <div class="d-flex align-items-center">
                    <div class="flex-grow-1">
                        <h6 class="card-subtitle mb-2 text-muted">Redirect Mode</h6>
                        <h3 class="card-title mb-0">
                            {{if .Config.UseRedirect}}
                                <span class="badge bg-info">On</span>
                            {{else}}
                                <span class="badge bg-secondary">Off</span>
                            {{end}}
                        </h3>
                    </div>
                    <div class="text-info">
                        <i class="fas fa-exchange-alt fa-2x"></i>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>

<div class="row">
    <div class="col-md-8 mb-4">
        <div class="card">
            <div class="card-header">
                <h5 class="mb-0">
                    <i class="fas fa-info-circle me-2"></i>System Information
                </h5>
            </div>
            <div class="card-body">
                <dl class="row">
                    <dt class="col-sm-4">Data Directory:</dt>
                    <dd class="col-sm-8"><code>{{.Config.DataDir}}</code></dd>
                    
                    <dt class="col-sm-4">WebDAV Endpoint:</dt>
                    <dd class="col-sm-8">
                        <a href="http://localhost:{{.Config.Port}}/" target="_blank">
                            http://localhost:{{.Config.Port}}/
                        </a>
                    </dd>
                    
                    <dt class="col-sm-4">API Endpoint:</dt>
                    <dd class="col-sm-8">
                        <a href="http://localhost:{{.Config.Port}}/api/" target="_blank">
                            http://localhost:{{.Config.Port}}/api/
                        </a>
                    </dd>
                    
                    <dt class="col-sm-4">Health Check:</dt>
                    <dd class="col-sm-8">
                        <a href="http://localhost:{{.Config.Port}}/api/health" target="_blank">
                            http://localhost:{{.Config.Port}}/api/health
                        </a>
                    </dd>
                </dl>
            </div>
        </div>
    </div>
    
    <div class="col-md-4 mb-4">
        <div class="card">
            <div class="card-header">
                <h5 class="mb-0">
                    <i class="fas fa-rocket me-2"></i>Quick Actions
                </h5>
            </div>
            <div class="card-body">
                <div class="d-grid gap-2">
                    <a href="/admin/files" class="btn btn-primary">
                        <i class="fas fa-plus me-2"></i>Add Files
                    </a>
                    <a href="/admin/import" class="btn btn-outline-primary">
                        <i class="fas fa-upload me-2"></i>Import Data
                    </a>
                    <a href="/admin/export" class="btn btn-outline-secondary">
                        <i class="fas fa-download me-2"></i>Export Data
                    </a>
                </div>
            </div>
        </div>
    </div>
</div>

<script>
function updateTime() {
    document.getElementById('current-time').textContent = new Date().toLocaleString();
}
updateTime();
setInterval(updateTime, 1000);
</script>
{{end}}

{{define "config"}}
<div class="d-flex justify-content-between align-items-center mb-4">
    <h1 class="h3 mb-0">
        <i class="fas fa-cog text-primary me-2"></i>Server Configuration
    </h1>
</div>

<div id="config-alerts"></div>

<div class="card">
    <div class="card-header">
        <h5 class="mb-0">
            <i class="fas fa-server me-2"></i>Current Configuration
        </h5>
    </div>
    <div class="card-body">
        <form hx-post="/admin/api/config" hx-target="#config-alerts">
            <div class="row">
                <div class="col-md-6 mb-3">
                    <label for="port" class="form-label">Server Port</label>
                    <input type="number" class="form-control" id="port" name="port" value="{{.Config.Port}}" min="1" max="65535">
                    <div class="form-text">Port number for the HTTP server</div>
                </div>
                
                <div class="col-md-6 mb-3">
                    <label for="data_dir" class="form-label">Data Directory</label>
                    <input type="text" class="form-control" id="data_dir" name="data_dir" value="{{.Config.DataDir}}">
                    <div class="form-text">Directory for persistent data storage</div>
                </div>
            </div>
            
            <div class="row">
                <div class="col-md-6 mb-3">
                    <div class="form-check">
                        <input class="form-check-input" type="checkbox" id="use_redirect" name="use_redirect" {{if .Config.UseRedirect}}checked{{end}}>
                        <label class="form-check-label" for="use_redirect">
                            Use 302 Redirects
                        </label>
                        <div class="form-text">Use redirects instead of proxying content</div>
                    </div>
                </div>
                
                <div class="col-md-6 mb-3">
                    <div class="form-check">
                        <input class="form-check-input" type="checkbox" id="auth_enabled" name="auth_enabled" {{if .Config.AuthEnabled}}checked{{end}} onchange="toggleAuthFields()">
                        <label class="form-check-label" for="auth_enabled">
                            HTTP Basic Authentication
                        </label>
                        <div class="form-text">Enable authentication for all endpoints</div>
                    </div>
                </div>
            </div>
            
            <div id="auth-fields" class="row" style="{{if not .Config.AuthEnabled}}display: none;{{end}}">
                <div class="col-md-6 mb-3">
                    <label for="auth_user" class="form-label">Username</label>
                    <input type="text" class="form-control" id="auth_user" name="auth_user" value="{{.Config.AuthUser}}">
                </div>
                
                <div class="col-md-6 mb-3">
                    <label for="auth_pass" class="form-label">Password</label>
                    <input type="password" class="form-control" id="auth_pass" name="auth_pass" placeholder="Leave empty to keep current password">
                    <div class="form-text">Leave empty to keep current password</div>
                </div>
            </div>

            <div class="d-grid gap-2 d-md-flex justify-content-md-end mb-3">
                <button type="submit" class="btn btn-primary">
                    <i class="fas fa-save me-2"></i>Update Configuration
                </button>
            </div>
            
            <div class="alert alert-info" role="alert">
                <i class="fas fa-info-circle me-2"></i>
                <strong>Dynamic Configuration:</strong> Most settings take effect immediately, including:
                <ul class="mb-1 mt-2">
                    <li><strong>Redirect Mode:</strong> Changes apply instantly</li>
                    <li><strong>Authentication:</strong> New credentials take effect immediately</li>
                </ul>
                Settings requiring restart: <strong>Port</strong> and <strong>Data Directory</strong>
            </div>
        </form>
    </div>
</div>

<div id="server-control-alerts"></div>

<div class="card mt-4">
    <div class="card-header">
        <h5 class="mb-0">
            <i class="fas fa-server me-2"></i>Server Control
        </h5>
    </div>
    <div class="card-body">
        <div class="row">
            <div class="col-md-6 mb-3">
                <h6><i class="fas fa-sync me-2 text-primary"></i>Restart Server</h6>
                <p class="text-muted mb-3">
                    Apply port and data directory changes by restarting the server. 
                    The server will restart automatically with the current configuration.
                </p>
                <button class="btn btn-warning" 
                        hx-post="/admin/api/restart" 
                        hx-target="#server-control-alerts"
                        hx-confirm="Are you sure you want to restart the server? This will briefly interrupt all connections.">
                    <i class="fas fa-sync me-2"></i>Restart Server
                </button>
            </div>
            
            <div class="col-md-6 mb-3">
                <h6><i class="fas fa-power-off me-2 text-danger"></i>Shutdown Server</h6>
                <p class="text-muted mb-3">
                    Gracefully shutdown the server. You'll need to manually restart it 
                    from the command line to resume operations.
                </p>
                <button class="btn btn-danger" 
                        hx-post="/admin/api/shutdown" 
                        hx-target="#server-control-alerts"
                        hx-confirm="Are you sure you want to shutdown the server? You'll need to manually restart it.">
                    <i class="fas fa-power-off me-2"></i>Shutdown Server
                </button>
            </div>
        </div>
        
        <div class="alert alert-warning" role="alert">
            <i class="fas fa-exclamation-triangle me-2"></i>
            <strong>Important:</strong> 
            <ul class="mb-0 mt-2">
                <li><strong>Restart:</strong> Server will automatically come back online with any configuration changes</li>
                <li><strong>Shutdown:</strong> Server will stop completely and require manual restart</li>
                <li>Both operations are graceful and will complete existing requests before stopping</li>
            </ul>
        </div>
    </div>
</div>

<script>
function toggleAuthFields() {
    const authEnabled = document.getElementById('auth_enabled').checked;
    const authFields = document.getElementById('auth-fields');
    authFields.style.display = authEnabled ? 'flex' : 'none';
}
</script>
{{end}}

{{define "files"}}
<div class="d-flex justify-content-between align-items-center mb-4">
    <h1 class="h3 mb-0">
        <i class="fas fa-file-alt text-primary me-2"></i>File Management
    </h1>
</div>

<div class="row mb-4">
    <div class="col-md-12">
        <div class="card">
            <div class="card-header">
                <h5 class="mb-0">
                    <i class="fas fa-plus me-2"></i>Add New File
                </h5>
            </div>
            <div class="card-body">
                <form hx-post="/admin/api/files" hx-target="#file-list">
                    <div class="row">
                        <div class="col-md-5 mb-3">
                            <label for="path" class="form-label">Virtual Path</label>
                            <input type="text" class="form-control" id="path" name="path" placeholder="/example/file.pdf" required>
                            <div class="form-text">Path where the file will be accessible in WebDAV</div>
                        </div>
                        
                        <div class="col-md-5 mb-3">
                            <label for="url" class="form-label">Source URL</label>
                            <input type="url" class="form-control" id="url" name="url" placeholder="https://example.com/file.pdf" required>
                            <div class="form-text">URL where the actual file is hosted</div>
                        </div>
                        
                        <div class="col-md-2 mb-3 d-flex align-items-end">
                            <button type="submit" class="btn btn-primary w-100">
                                <i class="fas fa-plus me-2"></i>Add File
                                <span class="loading-spinner">
                                    <i class="fas fa-spinner fa-spin"></i>
                                </span>
                            </button>
                        </div>
                    </div>
                </form>
            </div>
        </div>
    </div>
</div>

<div class="card">
    <div class="card-header">
        <h5 class="mb-0">
            <i class="fas fa-list me-2"></i>Configured Files
        </h5>
    </div>
    <div class="card-body">
        <div class="table-responsive">
            <table class="table table-hover">
                <thead>
                    <tr>
                        <th>Virtual Path</th>
                        <th>Source URL</th>
                        <th width="100">Actions</th>
                    </tr>
                </thead>
                <tbody id="file-list" hx-get="/admin/api/files" hx-trigger="load">
                    <!-- File list will be loaded here -->
                </tbody>
            </table>
        </div>
    </div>
</div>
{{end}}

{{define "import"}}
<div class="d-flex justify-content-between align-items-center mb-4">
    <h1 class="h3 mb-0">
        <i class="fas fa-upload text-primary me-2"></i>Import/Export Data
    </h1>
</div>

<div id="import-alerts"></div>

<div class="row">
    <div class="col-md-6 mb-4">
        <div class="card">
            <div class="card-header">
                <h5 class="mb-0">
                    <i class="fas fa-upload me-2"></i>Import Files
                </h5>
            </div>
            <div class="card-body">
                <form hx-post="/admin/api/import" hx-target="#import-alerts" hx-encoding="multipart/form-data">
                    <div class="mb-3">
                        <label for="import_file" class="form-label">Select JSON File</label>
                        <input class="form-control" type="file" id="import_file" name="import_file" accept=".json" required>
                        <div class="form-text">Choose a JSON file containing file entries to import</div>
                    </div>
                    
                    <button type="submit" class="btn btn-primary">
                        <i class="fas fa-upload me-2"></i>Import Files
                        <span class="loading-spinner">
                            <i class="fas fa-spinner fa-spin"></i>
                        </span>
                    </button>
                </form>
                
                <hr>
                
                <h6>Expected JSON Format:</h6>
                <pre class="bg-light p-3 rounded"><code>{
  "files": [
    {
      "path": "/example/file1.pdf",
      "url": "https://example.com/file1.pdf"
    },
    {
      "path": "/example/file2.pdf",
      "url": "https://example.com/file2.pdf"
    }
  ]
}</code></pre>
            </div>
        </div>
    </div>
    
    <div class="col-md-6 mb-4">
        <div class="card">
            <div class="card-header">
                <h5 class="mb-0">
                    <i class="fas fa-download me-2"></i>Export Files
                </h5>
            </div>
            <div class="card-body">
                <p>Export all currently configured files as a JSON file that can be imported later.</p>
                
                <a href="/admin/export" class="btn btn-outline-primary">
                    <i class="fas fa-download me-2"></i>Download Export
                </a>
                
                <hr>
                
                <h6>Export Information:</h6>
                <ul class="list-unstyled">
                    <li><i class="fas fa-check text-success me-2"></i>All file entries</li>
                    <li><i class="fas fa-check text-success me-2"></i>Export timestamp</li>
                    <li><i class="fas fa-check text-success me-2"></i>File count metadata</li>
                    <li><i class="fas fa-check text-success me-2"></i>Compatible with import function</li>
                </ul>
            </div>
        </div>
    </div>
</div>
{{end}}
`
