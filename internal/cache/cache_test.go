package cache

import (
	"fmt"
	"testing"
	"time"

	"proxydav/pkg/types"
)

func TestMetadataCache_SetAndGet(t *testing.T) {
	cache := New(time.Minute, 10)
	defer cache.Close()

	url := "https://example.com/file.txt"
	metadata := &types.FileMetadata{
		URL:          url,
		Size:         1024,
		LastModified: time.Now(),
	}

	// Test set and get
	cache.Set(url, metadata)
	retrieved := cache.Get(url)

	if retrieved == nil {
		t.Fatal("Expected metadata, got nil")
	}

	if retrieved.URL != metadata.URL {
		t.Errorf("Expected URL %s, got %s", metadata.URL, retrieved.URL)
	}

	if retrieved.Size != metadata.Size {
		t.Errorf("Expected size %d, got %d", metadata.Size, retrieved.Size)
	}
}

func TestMetadataCache_Expiration(t *testing.T) {
	cache := New(50*time.Millisecond, 10)
	defer cache.Close()

	url := "https://example.com/file.txt"
	metadata := &types.FileMetadata{
		URL:          url,
		Size:         1024,
		LastModified: time.Now(),
	}

	// Set metadata
	cache.Set(url, metadata)

	// Should be available immediately
	if retrieved := cache.Get(url); retrieved == nil {
		t.Error("Expected metadata to be available immediately")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired
	if retrieved := cache.Get(url); retrieved != nil {
		t.Error("Expected metadata to be expired")
	}
}

func TestMetadataCache_MaxSize(t *testing.T) {
	maxSize := 3
	cache := New(time.Minute, maxSize)
	defer cache.Close()

	// Add items up to max size
	for i := 0; i < maxSize; i++ {
		url := fmt.Sprintf("https://example.com/file%d.txt", i)
		metadata := &types.FileMetadata{
			URL:          url,
			Size:         int64(i),
			LastModified: time.Now(),
		}
		cache.Set(url, metadata)
	}

	if cache.Size() != maxSize {
		t.Errorf("Expected cache size %d, got %d", maxSize, cache.Size())
	}

	// Add one more item (should evict oldest)
	newURL := "https://example.com/newfile.txt"
	newMetadata := &types.FileMetadata{
		URL:          newURL,
		Size:         999,
		LastModified: time.Now(),
	}
	cache.Set(newURL, newMetadata)

	// Cache should still be at max size
	if cache.Size() != maxSize {
		t.Errorf("Expected cache size %d after eviction, got %d", maxSize, cache.Size())
	}

	// New item should be in cache
	if retrieved := cache.Get(newURL); retrieved == nil {
		t.Error("Expected new item to be in cache")
	}
}

func TestMetadataCache_Clear(t *testing.T) {
	cache := New(time.Minute, 10)
	defer cache.Close()

	// Add some items
	for i := 0; i < 5; i++ {
		url := fmt.Sprintf("https://example.com/file%d.txt", i)
		metadata := &types.FileMetadata{
			URL:          url,
			Size:         int64(i),
			LastModified: time.Now(),
		}
		cache.Set(url, metadata)
	}

	if cache.Size() != 5 {
		t.Errorf("Expected cache size 5, got %d", cache.Size())
	}

	// Clear cache
	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", cache.Size())
	}
}
