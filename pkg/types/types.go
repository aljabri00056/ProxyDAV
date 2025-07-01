package types

import "time"

// FileEntry represents a mapping between a virtual path and a remote URL
type FileEntry struct {
	Path string `json:"path"`
	URL  string `json:"url"`
}

// FileMetadata represents cached metadata about a remote file
type FileMetadata struct {
	URL          string
	Size         int64
	LastModified time.Time
	CachedAt     time.Time
}

// VirtualItem represents an item in the virtual filesystem
type VirtualItem struct {
	Name  string
	Path  string
	URL   string
	IsDir bool
}
