package storage

import (
	"testing"
	"time"

	"proxydav/pkg/types"
)

func TestPersistentStore_FileEntries(t *testing.T) {
	tempDir := t.TempDir()

	store, err := New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	entry := &types.FileEntry{
		Path: "/test/file.txt",
		URL:  "https://example.com/file.txt",
	}

	err = store.SetFileEntry(entry)
	if err != nil {
		t.Fatalf("Failed to set file entry: %v", err)
	}

	retrieved, err := store.GetFileEntry(entry.Path)
	if err != nil {
		t.Fatalf("Failed to get file entry: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected file entry, got nil")
	}

	if retrieved.Path != entry.Path {
		t.Errorf("Expected path %s, got %s", entry.Path, retrieved.Path)
	}

	if retrieved.URL != entry.URL {
		t.Errorf("Expected URL %s, got %s", entry.URL, retrieved.URL)
	}
}

func TestPersistentStore_FileMetadata(t *testing.T) {
	tempDir := t.TempDir()

	store, err := New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	metadata := &types.FileMetadata{
		URL:          "https://example.com/file.txt",
		Size:         1024,
		LastModified: time.Now().Truncate(time.Second), // Truncate for comparison
	}

	err = store.SetFileMetadata(metadata)
	if err != nil {
		t.Fatalf("Failed to set file metadata: %v", err)
	}

	retrieved, err := store.GetFileMetadata(metadata.URL)
	if err != nil {
		t.Fatalf("Failed to get file metadata: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected file metadata, got nil")
	}

	if retrieved.URL != metadata.URL {
		t.Errorf("Expected URL %s, got %s", metadata.URL, retrieved.URL)
	}

	if retrieved.Size != metadata.Size {
		t.Errorf("Expected size %d, got %d", metadata.Size, retrieved.Size)
	}

	if !retrieved.LastModified.Equal(metadata.LastModified) {
		t.Errorf("Expected last modified %v, got %v", metadata.LastModified, retrieved.LastModified)
	}
}

func TestPersistentStore_GetAllFileEntries(t *testing.T) {
	tempDir := t.TempDir()

	store, err := New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	entries := []*types.FileEntry{
		{Path: "/file1.txt", URL: "https://example.com/file1.txt"},
		{Path: "/dir/file2.txt", URL: "https://example.com/file2.txt"},
		{Path: "/file3.txt", URL: "https://example.com/file3.txt"},
	}

	for _, entry := range entries {
		err = store.SetFileEntry(entry)
		if err != nil {
			t.Fatalf("Failed to set file entry: %v", err)
		}
	}

	retrieved, err := store.GetAllFileEntries()
	if err != nil {
		t.Fatalf("Failed to get all file entries: %v", err)
	}

	if len(retrieved) != len(entries) {
		t.Errorf("Expected %d entries, got %d", len(entries), len(retrieved))
	}

	entryMap := make(map[string]string)
	for _, entry := range retrieved {
		entryMap[entry.Path] = entry.URL
	}

	for _, expected := range entries {
		if url, exists := entryMap[expected.Path]; !exists {
			t.Errorf("Expected entry with path %s not found", expected.Path)
		} else if url != expected.URL {
			t.Errorf("Expected URL %s for path %s, got %s", expected.URL, expected.Path, url)
		}
	}
}

func TestPersistentStore_DeleteFileEntry(t *testing.T) {
	tempDir := t.TempDir()

	store, err := New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	entry := &types.FileEntry{
		Path: "/test/file.txt",
		URL:  "https://example.com/file.txt",
	}

	err = store.SetFileEntry(entry)
	if err != nil {
		t.Fatalf("Failed to set file entry: %v", err)
	}

	err = store.DeleteFileEntry(entry.Path)
	if err != nil {
		t.Fatalf("Failed to delete file entry: %v", err)
	}

	retrieved, err := store.GetFileEntry(entry.Path)
	if err != nil {
		t.Fatalf("Failed to get file entry: %v", err)
	}

	if retrieved != nil {
		t.Error("Expected nil after deletion, but got entry")
	}
}

func TestPersistentStore_Persistence(t *testing.T) {
	tempDir := t.TempDir()

	store1, err := New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	entry := &types.FileEntry{
		Path: "/persistent/file.txt",
		URL:  "https://example.com/persistent.txt",
	}

	err = store1.SetFileEntry(entry)
	if err != nil {
		t.Fatalf("Failed to set file entry: %v", err)
	}

	store1.Close()

	store2, err := New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create second store: %v", err)
	}
	defer store2.Close()

	retrieved, err := store2.GetFileEntry(entry.Path)
	if err != nil {
		t.Fatalf("Failed to get file entry from new store: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected persisted entry, got nil")
	}

	if retrieved.Path != entry.Path || retrieved.URL != entry.URL {
		t.Errorf("Persisted data doesn't match: expected %+v, got %+v", entry, retrieved)
	}
}
