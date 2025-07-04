# ProxyDAV

WebDAV server that creates a virtual filesystem from remote HTTP/HTTPS resources.

## Features

- WebDAV protocol support
- Virtual filesystem from remote files  
- REST API for file management
- Persistent storage with BadgerDB
- Optional authentication
- Proxy or redirect modes

## Quick Start

### Installation

```bash
curl -sSL aj-get.vercel.app/ProxyDAV | bash
```

### Usage

```bash
# Start server
./proxydav

# Add files
curl -X POST http://localhost:8080/api/files/add \
  -H "Content-Type: application/json" \
  -d '{
    "files": [
      {"path":"/docs/file.pdf","url":"https://example.com/file.pdf"}
    ]
  }'
```

## Configuration

### Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-port` | Port to listen on | 8080 |
| `-data-dir` | Data storage directory | ./proxydavData |
| `-redirect` | Use redirects instead of proxying | false |
| `-auth` | Enable basic authentication | false |
| `-user` | Basic auth username | "" |
| `-pass` | Basic auth password | "" |

### Environment Variables

```bash
export PORT=9000
export DATA_DIR=/path/to/data
export USE_REDIRECT=true
export AUTH_ENABLED=true
export AUTH_USER=admin
export AUTH_PASS=secret
```

## API

### File Management

- `GET /api/files` - List all files
- `POST /api/files/add` - Add multiple files
- `DELETE /api/files/delete` - Delete multiple files

### Health Check

```http
GET /api/health
```

See [API.md](API.md) for detailed documentation.

