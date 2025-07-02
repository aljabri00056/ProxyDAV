# ProxyDAV

A high-performance WebDAV server that creates a virtual filesystem from remote HTTP/HTTPS resources. Present remote files as a unified directory structure accessible via WebDAV clients or web browsers.

## Features

- 🌐 **WebDAV Protocol Support** - Full compatibility with WebDAV clients
- 🗂️ **Virtual Filesystem** - Create directory structures from remote files
- 🚀 **High Performance** - Connection pooling, caching, and optimized HTTP handling
- � **REST API** - Complete CRUD API for dynamic file management
- �🔐 **Authentication** - Optional Basic HTTP authentication
- 📱 **Browser Support** - Beautiful web interface for directory browsing
- ⚡ **Caching** - Intelligent metadata caching with TTL
- 🔄 **Two Modes** - Proxy mode (stream files) or redirect mode (302 redirects)
- 🏥 **Health Checks** - Built-in health monitoring endpoint
- 🛡️ **Security** - Input validation, path sanitization, and URL validation
- 📊 **Logging** - Structured request logging with performance metrics
- 🔧 **Configuration** - Environment variables and command-line options
- 📝 **Dynamic Management** - Add, update, or remove files via REST API
- 🔄 **Bulk Operations** - Import/export file mappings via JSON

## Quick Start

### Installation

```bash
curl -sSL aj-get.vercel.app/ProxyDAV | bash
```

Or download the latest release for your platform from the [releases page](https://github.com/aljabri00056/ProxyDAV/releases).

### Build from Source

```bash
# Clone the repository
git clone https://github.com/aljabri00056/ProxyDAV.git
cd ProxyDAV

# Build the application
make build

# Or run directly with Go
go run .
```

### Basic Usage

1. **Start the server**:

```bash
./proxydav
```

2. **Add files via API**:

```bash
# Add a single file
curl -X POST http://localhost:8080/api/files \
  -H "Content-Type: application/json" \
  -d '{"path":"/documents/example.pdf","url":"https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf"}'

# Add multiple files
curl -X POST http://localhost:8080/api/files/bulk \
  -H "Content-Type: application/json" \
  -d '{
    "operation": "add",
    "files": [
      {"path":"/documents/example.pdf","url":"https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf"},
      {"path":"/images/sample.jpg","url":"https://via.placeholder.com/800x600.jpg"}
    ]
  }'
```

3. **Access your files**:
   - **Web Browser**: http://localhost:8080/
   - **WebDAV Client**: webdav://localhost:8080/
   - **REST API**: http://localhost:8080/api/files
   - **Health Check**: http://localhost:8080/health

#### Advanced Usage

```bash
# Custom port
./proxydav -port 9000

# With authentication
./proxydav -auth -user admin -pass secret

# Redirect mode (faster for large files)
./proxydav -redirect
```

## Configuration

### Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-port` | Port to listen on | 8080 |
| `-cache-ttl` | Cache TTL in seconds | 3600 |
| `-redirect` | Use redirects instead of proxying | false |
| `-auth` | Enable basic authentication | false |
| `-user` | Basic auth username | "" |
| `-pass` | Basic auth password | "" |

### Environment Variables

Environment variables override command-line flags:

```bash
export PORT=9000
export CACHE_TTL=600
export USE_REDIRECT=true
export AUTH_ENABLED=true
export AUTH_USER=admin
export AUTH_PASS=secret
```

### File Management

The only way to add files to ProxyDAV is through the REST API. This ensures a consistent, programmatic interface for all file operations.

```bash
# List all files
curl http://localhost:8080/api/files

# Add a file
curl -X POST http://localhost:8080/api/files \
  -H "Content-Type: application/json" \
  -d '{"path":"/docs/file.pdf","url":"https://example.com/file.pdf"}'

# Update a file
curl -X PUT http://localhost:8080/api/files/docs%2Ffile.pdf \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/updated-file.pdf"}'

# Delete a file
curl -X DELETE http://localhost:8080/api/files/docs%2Ffile.pdf

# Bulk operations
curl -X POST http://localhost:8080/api/files/bulk \
  -H "Content-Type: application/json" \
  -d '{
    "operation": "add",
    "files": [
      {"path":"/file1.pdf","url":"https://example.com/file1.pdf"},
      {"path":"/file2.pdf","url":"https://example.com/file2.pdf"}
    ]
  }'
```

See [API.md](API.md) for complete API documentation.


### Redirect Mode

Redirect mode returns HTTP 302 redirects instead of proxying files. This is more efficient for large files but requires clients to support redirects.

```bash
./proxydav -redirect
```


## API Endpoints

### File Management API

Complete REST API for managing files dynamically:

- **GET** `/api/files` - List all files
- **POST** `/api/files` - Add a single file
- **PUT** `/api/files/{path}` - Update a file
- **DELETE** `/api/files/{path}` - Delete a file
- **POST** `/api/files/bulk` - Bulk add/remove operations

See [API.md](API.md) for detailed API documentation with examples.

### Health Check

```http
GET /health
```

Response:
```json
{
    "status": "healthy",
    "cache_size": 150
}
```

### WebDAV Methods

- `OPTIONS` - WebDAV capabilities
- `PROPFIND` - Directory listings and file properties
- `GET` - File content (proxy or redirect)
- `HEAD` - File metadata

