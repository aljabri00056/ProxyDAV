# ProxyDAV

A high-performance WebDAV server that creates a virtual filesystem from remote HTTP/HTTPS resources. Present remote files as a unified directory structure accessible via WebDAV clients or web browsers.

## Features

- 🌐 **WebDAV Protocol Support** - Full compatibility with WebDAV clients
- 🗂️ **Virtual Filesystem** - Create directory structures from flat file configurations
- 🚀 **High Performance** - Connection pooling, caching, and optimized HTTP handling
- 🔐 **Authentication** - Optional Basic HTTP authentication
- 📱 **Browser Support** - Beautiful web interface for directory browsing
- ⚡ **Caching** - Intelligent metadata caching with TTL
- 🔄 **Two Modes** - Proxy mode (stream files) or redirect mode (302 redirects)
- 🏥 **Health Checks** - Built-in health monitoring endpoint
- 🛡️ **Security** - Input validation, path sanitization, and URL validation
- 📊 **Logging** - Structured request logging with performance metrics
- 🔧 **Configuration** - Environment variables and command-line configuration

## Quick Start

### Installation

```bash
curl aj-get.vercel.app/ProxyDAV | bash
```

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

1. **Create a configuration file** (`files.json`):

```json
[
    {
        "path": "/documents/example.pdf",
        "url": "https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf"
    },
    {
        "path": "/images/sample.jpg",
        "url": "https://via.placeholder.com/800x600.jpg"
    },
    {
        "path": "/nested/folder/deep/file.txt",
        "url": "https://www.w3.org/TR/PNG/iso_8859-1.txt"
    }
]
```

2. **Start the server**:

```bash
# Default settings (port 8080)
./proxydav

# Custom port and configuration
./proxydav -port 9000 -config myfiles.json

# With authentication
./proxydav -auth -auth-user admin -auth-pass secret

# Redirect mode (faster for large files)
./proxydav -redirect
```

3. **Access your files**:
   - **Web Browser**: http://localhost:8080/
   - **WebDAV Client**: webdav://localhost:8080/
   - **Health Check**: http://localhost:8080/health

## Configuration

### Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-port` | Port to listen on | 8080 |
| `-config` | JSON configuration file | files.json |
| `-cache-ttl` | Cache TTL in seconds | 3600 |
| `-redirect` | Use redirects instead of proxying | false |
| `-auth` | Enable basic authentication | false |
| `-auth-user` | Basic auth username | "" |
| `-auth-pass` | Basic auth password | "" |

### Environment Variables

Environment variables override command-line flags:

```bash
export PROXYDAV_PORT=9000
export PROXYDAV_CONFIG=myfiles.json
export PROXYDAV_CACHE_TTL=600
export PROXYDAV_USE_REDIRECT=true
export PROXYDAV_AUTH_ENABLED=true
export PROXYDAV_AUTH_USER=admin
export PROXYDAV_AUTH_PASS=secret
```

### File Configuration Format

The configuration file is a JSON array of file entries:

```json
[
    {
        "path": "/virtual/path/to/file.ext",
        "url": "https://remote-server.com/actual/file.ext"
    }
]
```

**Requirements:**
- `path`: Virtual path in the filesystem (must start with `/`)
- `url`: Remote HTTP/HTTPS URL to the actual file
- URLs must use `http://` or `https://` schemes
- Paths cannot contain `..` sequences


### Redirect Mode

Redirect mode returns HTTP 302 redirects instead of proxying files. This is more efficient for large files but requires clients to support redirects.

```bash
./proxydav -redirect
```


## API Endpoints

### Health Check

```http
GET /health
```

Response:
```json
{
    "status": "ok",
    "timestamp": 1672531200,
    "version": "1.0.0"
}
```

### WebDAV Methods

- `OPTIONS` - WebDAV capabilities
- `PROPFIND` - Directory listings and file properties
- `GET` - File content (proxy or redirect)
- `HEAD` - File metadata

