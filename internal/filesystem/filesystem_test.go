package filesystem

import (
	"testing"

	"proxydav/internal/storage"
	"proxydav/pkg/types"
)

func TestVirtualFS_Creation(t *testing.T) {
	tempDir := t.TempDir()

	store, err := storage.New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	files := []types.FileEntry{
		{Path: "/documents/file1.txt", URL: "https://example.com/file1.txt"},
		{Path: "/documents/subfolder/file2.txt", URL: "https://example.com/file2.txt"},
		{Path: "/images/photo.jpg", URL: "https://example.com/photo.jpg"},
	}

	for _, file := range files {
		err := store.SetFileEntry(&file)
		if err != nil {
			t.Fatalf("Failed to set file entry: %v", err)
		}
	}

	vfs, err := New(store)
	if err != nil {
		t.Fatalf("Failed to create VFS: %v", err)
	}

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
	tempDir := t.TempDir()

	store, err := storage.New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Add test file to store
	file := types.FileEntry{Path: "/test/file.txt", URL: "https://example.com/file.txt"}
	err = store.SetFileEntry(&file)
	if err != nil {
		t.Fatalf("Failed to set file entry: %v", err)
	}

	vfs, err := New(store)
	if err != nil {
		t.Fatalf("Failed to create VFS: %v", err)
	}

	// Test getting a file
	item, exists := vfs.GetItem("/test/file.txt")
	if !exists {
		t.Fatal("File should exist")
	}

	if item.IsDir {
		t.Error("File should not be a directory")
	}

	if item.URL != "https://example.com/file.txt" {
		t.Errorf("Expected URL https://example.com/file.txt, got %s", item.URL)
	}

	// Test getting a directory
	dirItem, exists := vfs.GetItem("/test")
	if !exists {
		t.Fatal("Directory should exist")
	}

	if !dirItem.IsDir {
		t.Error("Directory should be marked as directory")
	}
}

func TestVirtualFS_ListDir(t *testing.T) {
	tempDir := t.TempDir()

	store, err := storage.New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	files := []types.FileEntry{
		{Path: "/folder/file1.txt", URL: "https://example.com/file1.txt"},
		{Path: "/folder/file2.txt", URL: "https://example.com/file2.txt"},
		{Path: "/folder/subfolder/file3.txt", URL: "https://example.com/file3.txt"},
	}

	for _, file := range files {
		err := store.SetFileEntry(&file)
		if err != nil {
			t.Fatalf("Failed to set file entry: %v", err)
		}
	}

	vfs, err := New(store)
	if err != nil {
		t.Fatalf("Failed to create VFS: %v", err)
	}

	// List contents of /folder
	items := vfs.ListDir("/folder")
	if len(items) != 3 { // 2 files + 1 directory
		t.Errorf("Expected 3 items in /folder, got %d", len(items))
	}

	// Check that directories come first (alphabetically sorted)
	foundDir := false
	foundFiles := 0
	for _, item := range items {
		if item.IsDir {
			foundDir = true
			if foundFiles > 0 {
				t.Error("Directories should come before files")
			}
		} else {
			foundFiles++
		}
	}

	if !foundDir {
		t.Error("Should have found subfolder directory")
	}
	if foundFiles != 2 {
		t.Errorf("Should have found 2 files, got %d", foundFiles)
	}
}

func TestVirtualFS_AddFile(t *testing.T) {
	tempDir := t.TempDir()

	store, err := storage.New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	vfs, err := New(store)
	if err != nil {
		t.Fatalf("Failed to create VFS: %v", err)
	}

	// Add a file
	err = vfs.AddFile("/new/file.txt", "https://example.com/new.txt")
	if err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	// Verify it exists
	if !vfs.Exists("/new/file.txt") {
		t.Error("Added file should exist")
	}

	// Verify parent directory was created
	if !vfs.IsDir("/new") {
		t.Error("Parent directory should have been created")
	}

	// Try to add the same file again (should fail)
	err = vfs.AddFile("/new/file.txt", "https://example.com/duplicate.txt")
	if err == nil {
		t.Error("Adding duplicate file should fail")
	}
}

func TestVirtualFS_RemoveFile(t *testing.T) {
	tempDir := t.TempDir()

	store, err := storage.New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Add test file to store
	file := types.FileEntry{Path: "/temp/file.txt", URL: "https://example.com/temp.txt"}
	err = store.SetFileEntry(&file)
	if err != nil {
		t.Fatalf("Failed to set file entry: %v", err)
	}

	vfs, err := New(store)
	if err != nil {
		t.Fatalf("Failed to create VFS: %v", err)
	}

	// Verify file exists
	if !vfs.Exists("/temp/file.txt") {
		t.Fatal("File should exist before removal")
	}

	// Remove the file
	err = vfs.RemoveFile("/temp/file.txt")
	if err != nil {
		t.Fatalf("Failed to remove file: %v", err)
	}

	// Verify file is gone
	if vfs.Exists("/temp/file.txt") {
		t.Error("File should not exist after removal")
	}

	// Verify it's also removed from persistent storage
	retrieved, err := store.GetFileEntry("/temp/file.txt")
	if err != nil {
		t.Fatalf("Failed to check storage: %v", err)
	}
	if retrieved != nil {
		t.Error("File should be removed from persistent storage")
	}
}
