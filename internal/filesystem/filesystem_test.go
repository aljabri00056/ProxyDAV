package filesystem

import (
	"testing"

	"proxydav/pkg/types"
)

func TestVirtualFS_Creation(t *testing.T) {
	files := []types.FileEntry{
		{Path: "/documents/file1.txt", URL: "https://example.com/file1.txt"},
		{Path: "/documents/subfolder/file2.txt", URL: "https://example.com/file2.txt"},
		{Path: "/images/photo.jpg", URL: "https://example.com/photo.jpg"},
	}

	vfs := New(files)

	// Test that all files exist
	for _, file := range files {
		if !vfs.Exists(file.Path) {
			t.Errorf("File %s should exist", file.Path)
		}
	}

	// Test that directories were created
	expectedDirs := []string{"/", "/documents", "/documents/subfolder", "/images"}
	for _, dir := range expectedDirs {
		if !vfs.Exists(dir) {
			t.Errorf("Directory %s should exist", dir)
		}
		if !vfs.IsDir(dir) {
			t.Errorf("Path %s should be a directory", dir)
		}
	}
}

func TestVirtualFS_GetItem(t *testing.T) {
	files := []types.FileEntry{
		{Path: "/test/file.txt", URL: "https://example.com/file.txt"},
	}

	vfs := New(files)

	// Test getting a file
	item, exists := vfs.GetItem("/test/file.txt")
	if !exists {
		t.Fatal("File should exist")
	}

	if item.IsDir {
		t.Error("File should not be a directory")
	}

	if item.URL != "https://example.com/file.txt" {
		t.Errorf("Expected URL 'https://example.com/file.txt', got '%s'", item.URL)
	}

	// Test getting a directory
	item, exists = vfs.GetItem("/test")
	if !exists {
		t.Fatal("Directory should exist")
	}

	if !item.IsDir {
		t.Error("Directory should be marked as directory")
	}
}

func TestVirtualFS_ListDir(t *testing.T) {
	files := []types.FileEntry{
		{Path: "/docs/file1.txt", URL: "https://example.com/file1.txt"},
		{Path: "/docs/file2.txt", URL: "https://example.com/file2.txt"},
		{Path: "/docs/subfolder/file3.txt", URL: "https://example.com/file3.txt"},
		{Path: "/images/photo.jpg", URL: "https://example.com/photo.jpg"},
	}

	vfs := New(files)

	// Test listing root directory
	rootItems := vfs.ListDir("/")
	if len(rootItems) != 2 { // docs and images
		t.Errorf("Expected 2 items in root, got %d", len(rootItems))
	}

	// Test listing docs directory
	docsItems := vfs.ListDir("/docs")
	if len(docsItems) != 3 { // file1.txt, file2.txt, subfolder
		t.Errorf("Expected 3 items in /docs, got %d", len(docsItems))
	}

	// Check that directories come first
	if !docsItems[0].IsDir {
		t.Error("First item should be a directory (subfolder)")
	}

	// Test listing non-existent directory
	nonExistentItems := vfs.ListDir("/nonexistent")
	if nonExistentItems != nil {
		t.Error("Non-existent directory should return nil")
	}
}

func TestVirtualFS_GetAllPaths(t *testing.T) {
	files := []types.FileEntry{
		{Path: "/test/file.txt", URL: "https://example.com/file.txt"},
		{Path: "/other/file.txt", URL: "https://example.com/other.txt"},
	}

	vfs := New(files)

	paths := vfs.GetAllPaths()

	// Should include all files and directories
	expectedPaths := []string{"/", "/other", "/other/file.txt", "/test", "/test/file.txt"}
	
	if len(paths) != len(expectedPaths) {
		t.Errorf("Expected %d paths, got %d", len(expectedPaths), len(paths))
	}

	// Check that paths are sorted
	for i := 1; i < len(paths); i++ {
		if paths[i-1] > paths[i] {
			t.Error("Paths should be sorted")
			break
		}
	}
}
