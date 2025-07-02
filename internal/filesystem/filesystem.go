package filesystem

import (
	"fmt"
	"path"
	"sort"
	"strings"
	"sync"

	"proxydav/pkg/types"
)

// VirtualFS represents the virtual filesystem structure
type VirtualFS struct {
	items map[string]*types.VirtualItem
	dirs  map[string]bool
	mutex sync.RWMutex // Add mutex for thread safety
}

// New creates a new virtual filesystem from file entries
func New(files []types.FileEntry) *VirtualFS {
	vfs := &VirtualFS{
		items: make(map[string]*types.VirtualItem),
		dirs:  make(map[string]bool),
	}

	// Add root directory
	vfs.dirs["/"] = true

	// Process each file entry
	for _, file := range files {
		vfs.addFile(file.Path, file.URL)
	}

	return vfs
}

// addFile adds a file to the virtual filesystem
func (vfs *VirtualFS) addFile(filePath, fileURL string) {
	// Normalize path
	filePath = path.Clean("/" + strings.TrimPrefix(filePath, "/"))

	// Add the file itself
	vfs.items[filePath] = &types.VirtualItem{
		Name:  path.Base(filePath),
		Path:  filePath,
		URL:   fileURL,
		IsDir: false,
	}

	// Add all parent directories
	dir := path.Dir(filePath)
	for dir != "/" && dir != "." {
		if !vfs.dirs[dir] {
			vfs.dirs[dir] = true
			vfs.items[dir] = &types.VirtualItem{
				Name:  path.Base(dir),
				Path:  dir,
				URL:   "",
				IsDir: true,
			}
		}
		dir = path.Dir(dir)
	}
}

// Exists checks if a path exists in the virtual filesystem
func (vfs *VirtualFS) Exists(path string) bool {
	vfs.mutex.RLock()
	defer vfs.mutex.RUnlock()
	_, exists := vfs.items[path]
	return exists || vfs.dirs[path]
}

// IsDir checks if a path is a directory
func (vfs *VirtualFS) IsDir(path string) bool {
	vfs.mutex.RLock()
	defer vfs.mutex.RUnlock()
	if item, exists := vfs.items[path]; exists {
		return item.IsDir
	}
	return vfs.dirs[path]
}

// GetItem returns the virtual item at the given path
func (vfs *VirtualFS) GetItem(path string) (*types.VirtualItem, bool) {
	vfs.mutex.RLock()
	defer vfs.mutex.RUnlock()
	item, exists := vfs.items[path]
	return item, exists
}

// ListDir returns the contents of a directory
func (vfs *VirtualFS) ListDir(dirPath string) []*types.VirtualItem {
	vfs.mutex.RLock()
	defer vfs.mutex.RUnlock()

	if !vfs.isDir(dirPath) {
		return nil
	}

	// Normalize directory path
	dirPath = path.Clean("/" + strings.TrimPrefix(dirPath, "/"))
	if dirPath != "/" {
		dirPath = strings.TrimSuffix(dirPath, "/")
	}

	var items []*types.VirtualItem

	// Find all direct children
	for itemPath, item := range vfs.items {
		itemDir := path.Dir(itemPath)
		if itemDir == dirPath {
			items = append(items, item)
		}
	}

	// Sort items: directories first, then files, both alphabetically
	sort.Slice(items, func(i, j int) bool {
		if items[i].IsDir != items[j].IsDir {
			return items[i].IsDir
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	return items
}

// isDir is an internal helper method that doesn't acquire locks
func (vfs *VirtualFS) isDir(path string) bool {
	if item, exists := vfs.items[path]; exists {
		return item.IsDir
	}
	return vfs.dirs[path]
}

// GetAllPaths returns all paths in the filesystem
func (vfs *VirtualFS) GetAllPaths() []string {
	vfs.mutex.RLock()
	defer vfs.mutex.RUnlock()

	var paths []string
	for path := range vfs.items {
		paths = append(paths, path)
	}
	for dir := range vfs.dirs {
		if _, exists := vfs.items[dir]; !exists {
			paths = append(paths, dir)
		}
	}
	sort.Strings(paths)
	return paths
}

// AddFile adds a new file to the virtual filesystem
func (vfs *VirtualFS) AddFile(filePath, fileURL string) error {
	vfs.mutex.Lock()
	defer vfs.mutex.Unlock()

	// Normalize path
	filePath = path.Clean("/" + strings.TrimPrefix(filePath, "/"))

	// Check if file already exists
	if _, exists := vfs.items[filePath]; exists {
		return fmt.Errorf("file already exists at path: %s", filePath)
	}

	// Check if there's a directory at this path
	if vfs.dirs[filePath] {
		return fmt.Errorf("directory exists at path: %s", filePath)
	}

	vfs.addFile(filePath, fileURL)
	return nil
}

// UpdateFile updates an existing file in the virtual filesystem
func (vfs *VirtualFS) UpdateFile(filePath, fileURL string) error {
	vfs.mutex.Lock()
	defer vfs.mutex.Unlock()

	// Normalize path
	filePath = path.Clean("/" + strings.TrimPrefix(filePath, "/"))

	// Check if file exists
	item, exists := vfs.items[filePath]
	if !exists {
		return fmt.Errorf("file not found at path: %s", filePath)
	}

	// Check if it's actually a file
	if item.IsDir {
		return fmt.Errorf("cannot update directory at path: %s", filePath)
	}

	// Update the URL
	item.URL = fileURL
	return nil
}

// RemoveFile removes a file from the virtual filesystem
func (vfs *VirtualFS) RemoveFile(filePath string) error {
	vfs.mutex.Lock()
	defer vfs.mutex.Unlock()

	// Normalize path
	filePath = path.Clean("/" + strings.TrimPrefix(filePath, "/"))

	// Check if file exists
	item, exists := vfs.items[filePath]
	if !exists {
		return fmt.Errorf("file not found at path: %s", filePath)
	}

	// Check if it's actually a file
	if item.IsDir {
		return fmt.Errorf("cannot remove directory at path: %s", filePath)
	}

	// Remove the file
	delete(vfs.items, filePath)

	// Clean up empty parent directories
	vfs.cleanupEmptyDirectories(filePath)
	return nil
}

// GetAllFiles returns all files (not directories) in the filesystem
func (vfs *VirtualFS) GetAllFiles() []types.FileEntry {
	vfs.mutex.RLock()
	defer vfs.mutex.RUnlock()

	var files []types.FileEntry
	for _, item := range vfs.items {
		if !item.IsDir {
			files = append(files, types.FileEntry{
				Path: item.Path,
				URL:  item.URL,
			})
		}
	}

	// Sort files by path
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	return files
}

// cleanupEmptyDirectories removes empty parent directories after file removal
func (vfs *VirtualFS) cleanupEmptyDirectories(filePath string) {
	dir := path.Dir(filePath)

	for dir != "/" && dir != "." {
		// Check if directory has any children
		hasChildren := false
		for itemPath := range vfs.items {
			if path.Dir(itemPath) == dir {
				hasChildren = true
				break
			}
		}

		// If no children, remove the directory
		if !hasChildren {
			delete(vfs.items, dir)
			delete(vfs.dirs, dir)
			dir = path.Dir(dir)
		} else {
			break
		}
	}
}
