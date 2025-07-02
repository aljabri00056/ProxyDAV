# ProxyDAV

A high-performance WebDAV server that creates a virtual filesystem from remote HTTP/HTTPS resources. Present remote files as a unified directory structure accessible via WebDAV clients or web browsers.

## Features

- 🌐 **WebDAV Protocol Support** - Full compatibility with WebDAV clients
- 🗂️ **Virtual Filesystem** - Create directory structures from remote files
- 🚀 **High Performance** - Connection pooling, persistent storage, and optimized HTTP handling
- 🛠️ **REST API** - Complete CRUD API for dynamic file management
- 🔐 **Authentication** - Optional Basic HTTP authentication
- 📱 **Browser Support** - Beautiful web interface for directory browsing
- 💾 **Persistent Storage** - BadgerDB-based persistence with automatic data recovery
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

# Custom data directory for persistent storage
./proxydav -data-dir /path/to/data

# With authentication
./proxydav -auth -user admin -pass secret

# Redirect mode (faster for large files)
./proxydav -redirect
```

## Persistent Storage

ProxyDAV uses BadgerDB for persistent storage of file entries and metadata. All data is automatically saved and restored across server restarts.

- **File Entries**: Virtual path to URL mappings are permanently stored
- **Metadata**: File size and modification times are cached persistently
- **Auto-Recovery**: On startup, the server automatically loads all existing files from the database
- **Data Directory**: Configurable storage location (default: `./data`)

## Data Persistence

ProxyDAV uses BadgerDB, a high-performance embedded database, for data persistence. This ensures that all your file mappings and metadata survive server restarts and system crashes.

### Data Storage

- **Location**: Configurable via `-data-dir` flag (default: `./data`)
- **Format**: BadgerDB database files
- **Backup**: Copy the entire data directory to backup your configuration
- **Migration**: Simply move the data directory to migrate between servers

### What Gets Persisted

1. **File Entries**: All virtual path to URL mappings
2. **File Metadata**: Size and last-modified times for performance
3. **Directory Structure**: Automatically reconstructed from file entries

### Data Recovery

On startup, ProxyDAV automatically:
1. Opens the BadgerDB database
2. Loads all existing file entries
3. Reconstructs the virtual filesystem in memory
4. Resumes normal operation with all previous data intact

### Example: Backing Up Data

```bash
# Stop the server
pkill proxydav

# Backup the data directory
cp -r ./data ./data-backup-$(date +%Y%m%d)

# Restart the server
./proxydav
```

### Example: Migrating to a New Server

```bash
# On old server - backup data
tar -czf proxydav-data.tar.gz ./data

# On new server - restore data
tar -xzf proxydav-data.tar.gz
./proxydav -data-dir ./data
```

## Configuration

### Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-port` | Port to listen on | 8080 |
| `-data-dir` | Directory for persistent data storage | ./data |
| `-redirect` | Use redirects instead of proxying | false |
| `-auth` | Enable basic authentication | false |
| `-user` | Basic auth username | "" |
| `-pass` | Basic auth password | "" |

### Environment Variables

Environment variables override command-line flags:

```bash
export PORT=9000
export DATA_DIR=/path/to/data
export USE_REDIRECT=true
export AUTH_ENABLED=true
export AUTH_USER=admin
export AUTH_PASS=secret
```

### File Management

ProxyDAV provides a REST API for dynamic file management. All file entries are automatically persisted to BadgerDB and survive server restarts.

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

**Note**: Files added via the API are immediately persisted to BadgerDB and will be available after server restarts.

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
    "data_dir": "./data"
}
```

### WebDAV Methods

- `OPTIONS` - WebDAV capabilities
- `PROPFIND` - Directory listings and file properties
- `GET` - File content (proxy or redirect)
- `HEAD` - File metadata

