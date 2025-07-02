# ProxyDAV API Documentation

This document describes the REST API endpoints for managing files in ProxyDAV. All API endpoints are prefixed with `/api` and require the same authentication as the WebDAV interface (if enabled).

## Base URL
```
http://localhost:8080/api
```

## Authentication
If authentication is enabled on the server, all API endpoints require HTTP Basic Authentication using the same credentials as the WebDAV interface.

## Content Type
All requests and responses use `application/json` content type.

## Response Format
All API responses follow this standard format:

```json
{
  "success": boolean,
  "message": "string (optional)",
  "data": object (optional),
  "error": "string (optional)"
}
```

## Endpoints

### 1. List All Files
**GET** `/api/files`

Returns a list of all files currently managed by the server.

#### Response
```json
{
  "success": true,
  "message": "Files retrieved successfully",
  "data": {
    "files": [
      {
        "path": "/documents/file1.pdf",
        "url": "https://example.com/file1.pdf"
      },
      {
        "path": "/images/photo.jpg",
        "url": "https://example.com/photo.jpg"
      }
    ],
    "total": 2
  }
}
```

### 2. Add Files
**POST** `/api/files/add`

Adds multiple files to the virtual filesystem.

#### Request Body
```json
{
  "files": [
    {
      "path": "/documents/file1.pdf",
      "url": "https://example.com/file1.pdf"
    },
    {
      "path": "/documents/file2.pdf",
      "url": "https://example.com/file2.pdf"
    }
  ]
}
```

#### Response (All Successful)
```json
{
  "success": true,
  "message": "Add operation completed: 2 successful, 0 failed",
  "data": {
    "successful": 2,
    "failed": 0,
    "files": [
      {
        "path": "/documents/file1.pdf",
        "url": "https://example.com/file1.pdf"
      },
      {
        "path": "/documents/file2.pdf",
        "url": "https://example.com/file2.pdf"
      }
    ]
  }
}
```

#### Response (Partial Success)
```json
{
  "success": true,
  "message": "Add operation completed: 1 successful, 1 failed",
  "data": {
    "successful": 1,
    "failed": 1,
    "files": [
      {
        "path": "/documents/file1.pdf",
        "url": "https://example.com/file1.pdf"
      }
    ],
    "errors": {
      "/documents/file2.pdf": "file already exists at path: /documents/file2.pdf"
    }
  }
}
```

### 3. Delete Files
**DELETE** `/api/files/delete`

Removes multiple files from the virtual filesystem.

#### Request Body
```json
{
  "files": [
    {
      "path": "/documents/file1.pdf"
    },
    {
      "path": "/documents/file2.pdf"
    }
  ]
}
```

#### Response (All Successful)
```json
{
  "success": true,
  "message": "Delete operation completed: 2 successful, 0 failed",
  "data": {
    "successful": 2,
    "failed": 0,
    "files": [
      {
        "path": "/documents/file1.pdf"
      },
      {
        "path": "/documents/file2.pdf"
      }
    ]
  }
}
```

#### Response (Partial Success)
```json
{
  "success": true,
  "message": "Delete operation completed: 1 successful, 1 failed",
  "data": {
    "successful": 1,
    "failed": 1,
    "files": [
      {
        "path": "/documents/file1.pdf"
      }
    ],
    "errors": {
      "/documents/file2.pdf": "File not found"
    }
  }
}
```

## Error Codes

- **400 Bad Request**: Invalid JSON payload, missing required fields, or invalid data
- **401 Unauthorized**: Authentication required or invalid credentials
- **404 Not Found**: File or endpoint not found
- **405 Method Not Allowed**: HTTP method not supported for the endpoint
- **409 Conflict**: File already exists (for POST operations)
- **500 Internal Server Error**: Server-side error

## Usage Examples

### Using cURL

#### List all files
```bash
curl http://localhost:8080/api/files
```

#### Add files
```bash
curl -X POST http://localhost:8080/api/files/add \
  -H "Content-Type: application/json" \
  -d '{
    "files": [
      {"path":"/documents/report.pdf","url":"https://example.com/report.pdf"},
      {"path":"/documents/manual.pdf","url":"https://example.com/manual.pdf"}
    ]
  }'
```

#### Delete files
```bash
curl -X DELETE http://localhost:8080/api/files/delete \
  -H "Content-Type: application/json" \
  -d '{
    "files": [
      {"path":"/documents/report.pdf"},
      {"path":"/documents/manual.pdf"}
    ]
  }'
```

### With Authentication
If authentication is enabled, include basic auth credentials:

```bash
curl -u username:password -X POST http://localhost:8080/api/files/add \
  -H "Content-Type: application/json" \
  -d '{
    "files": [
      {"path":"/documents/report.pdf","url":"https://example.com/report.pdf"}
    ]
  }'
```

## Notes
- All file paths are automatically normalized (e.g., `/path/` becomes `/path`)
- URLs must be valid HTTP or HTTPS URLs for add operations
- The URL field is optional for delete operations
- The API automatically creates parent directories as needed
- Empty directories are automatically cleaned up when the last file is removed
- The virtual filesystem is thread-safe and supports concurrent operations
- Both add and delete operations support batch processing and return detailed results including success/failure counts and error details
