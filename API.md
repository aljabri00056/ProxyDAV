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

### 2. Add Single File
**POST** `/api/files`

Adds a single file to the virtual filesystem.

#### Request Body
```json
{
  "path": "/documents/newfile.pdf",
  "url": "https://example.com/newfile.pdf"
}
```

#### Response
```json
{
  "success": true,
  "message": "File added successfully",
  "data": {
    "path": "/documents/newfile.pdf",
    "url": "https://example.com/newfile.pdf"
  }
}
```

#### Error Response (Conflict)
```json
{
  "success": false,
  "error": "Failed to add file: file already exists at path: /documents/newfile.pdf"
}
```

### 3. Update Single File
**PUT** `/api/files/{path}`

Updates the URL of an existing file. The path parameter should be URL-encoded.

#### Request Body
```json
{
  "url": "https://example.com/updated-file.pdf"
}
```

#### Response
```json
{
  "success": true,
  "message": "File updated successfully",
  "data": {
    "path": "/documents/file.pdf",
    "url": "https://example.com/updated-file.pdf"
  }
}
```

#### Error Response (Not Found)
```json
{
  "success": false,
  "error": "File not found"
}
```

### 4. Delete Single File
**DELETE** `/api/files/{path}`

Removes a file from the virtual filesystem. The path parameter should be URL-encoded.

#### Response
```json
{
  "success": true,
  "message": "File deleted successfully",
  "data": {
    "path": "/documents/file.pdf"
  }
}
```

#### Error Response (Not Found)
```json
{
  "success": false,
  "error": "File not found"
}
```

### 5. Bulk Operations
**POST** `/api/files/bulk`

Performs bulk add or remove operations on multiple files.

#### Request Body (Bulk Add)
```json
{
  "operation": "add",
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

#### Request Body (Bulk Remove)
```json
{
  "operation": "remove",
  "files": [
    {
      "path": "/documents/file1.pdf",
      "url": ""
    },
    {
      "path": "/documents/file2.pdf", 
      "url": ""
    }
  ]
}
```

#### Response (All Successful)
```json
{
  "success": true,
  "message": "Bulk add operation completed: 2 successful, 0 failed",
  "data": {
    "successful": 2,
    "failed": 0
  }
}
```

#### Response (Partial Success)
```json
{
  "success": true,
  "message": "Bulk add operation completed: 1 successful, 1 failed",
  "data": {
    "successful": 1,
    "failed": 1,
    "errors": {
      "/documents/file1.pdf": "file already exists at path: /documents/file1.pdf"
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

#### Add a single file
```bash
curl -X POST http://localhost:8080/api/files \
  -H "Content-Type: application/json" \
  -d '{"path":"/documents/report.pdf","url":"https://example.com/report.pdf"}'
```

#### List all files
```bash
curl http://localhost:8080/api/files
```

#### Update a file (note URL encoding for path)
```bash
curl -X PUT http://localhost:8080/api/files/documents%2Freport.pdf \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/updated-report.pdf"}'
```

#### Delete a file
```bash
curl -X DELETE http://localhost:8080/api/files/documents%2Freport.pdf
```

#### Bulk add files
```bash
curl -X POST http://localhost:8080/api/files/bulk \
  -H "Content-Type: application/json" \
  -d '{
    "operation": "add",
    "files": [
      {"path":"/docs/file1.pdf","url":"https://example.com/file1.pdf"},
      {"path":"/docs/file2.pdf","url":"https://example.com/file2.pdf"}
    ]
  }'
```

### With Authentication
If authentication is enabled, include basic auth credentials:

```bash
curl -u username:password -X POST http://localhost:8080/api/files \
  -H "Content-Type: application/json" \
  -d '{"path":"/documents/report.pdf","url":"https://example.com/report.pdf"}'
```

## Path Encoding
When using file paths in URLs (for PUT and DELETE operations), make sure to properly URL-encode the path. For example:
- `/documents/file.pdf` becomes `documents%2Ffile.pdf`
- `/folder with spaces/file.pdf` becomes `folder%20with%20spaces%2Ffile.pdf`

## Notes
- All file paths are automatically normalized (e.g., `/path/` becomes `/path`)
- URLs must be valid HTTP or HTTPS URLs
- The API automatically creates parent directories as needed
- Empty directories are automatically cleaned up when the last file is removed
- The virtual filesystem is thread-safe and supports concurrent operations
