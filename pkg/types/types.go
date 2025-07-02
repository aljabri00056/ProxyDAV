package types

import "time"

type FileEntry struct {
	Path string `json:"path"`
	URL  string `json:"url"`
}

type FileMetadata struct {
	URL          string    `json:"url"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
}

type VirtualItem struct {
	Name  string
	Path  string
	URL   string
	IsDir bool
}
