package filesystem

import (
	"path"
	"sort"
	"strings"

	"proxydav/pkg/types"
)

// VirtualFS represents the virtual filesystem structure
type VirtualFS struct {
	items map[string]*types.VirtualItem
	dirs  map[string]bool
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
	_, exists := vfs.items[path]
	return exists || vfs.dirs[path]
}

// IsDir checks if a path is a directory
func (vfs *VirtualFS) IsDir(path string) bool {
	if item, exists := vfs.items[path]; exists {
		return item.IsDir
	}
	return vfs.dirs[path]
}

// GetItem returns the virtual item at the given path
func (vfs *VirtualFS) GetItem(path string) (*types.VirtualItem, bool) {
	item, exists := vfs.items[path]
	return item, exists
}

// ListDir returns the contents of a directory
func (vfs *VirtualFS) ListDir(dirPath string) []*types.VirtualItem {
	if !vfs.IsDir(dirPath) {
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

// GetAllPaths returns all paths in the filesystem
func (vfs *VirtualFS) GetAllPaths() []string {
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
