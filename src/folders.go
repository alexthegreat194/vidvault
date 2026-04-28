package main

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var foldersLog = fileLogger("folders")

// folderInfo describes a subdirectory under the media root.
// It is serialised to JSON and returned by the /api/folders endpoint.
type folderInfo struct {
	Name          string `json:"name"`            // slash-separated path relative to root
	HasOtherFiles bool   `json:"has_other_files"` // true if the directory contains non-video files
}

func getFolderMetadata(root string) ([]folderInfo, error) {
	foldersLog.Debug("scanning folder metadata", "root", root)
	// map of existing folder metadata
	seen := map[string]*folderInfo{}

	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || path == root {
			return err
		}

		rel, _ := filepath.Rel(root, path)
		// Replace all \ with / for windows edge cases
		slashRel := filepath.ToSlash(rel)

		if d.IsDir() {
			// Add folder meta data to seen map
			exists := seen[slashRel]
			if exists == nil {
				seen[slashRel] = &folderInfo{Name: slashRel}
			}
		} else {
			// check for existing files in directory
			ext := strings.ToLower(filepath.Ext(path))
			parent := filepath.ToSlash(filepath.Dir(rel))
			inSubDir := parent != "."

			if !isValidVideoExtention(ext) && inSubDir {
				parentDirDetails := seen[parent]
				if parentDirDetails != nil {
					parentDirDetails.HasOtherFiles = true
				} else {
					seen[parent] = &folderInfo{Name: parent, HasOtherFiles: true}
				}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	folders := make([]folderInfo, 0, len(seen))
	for _, fi := range seen {
		folders = append(folders, *fi)
	}

	// Sort by name
	sort.Slice(folders, func(i, j int) bool {
		return folders[i].Name < folders[j].Name
	})

	foldersLog.Debug("folder metadata scan completed", "root", root, "count", len(folders))
	return folders, nil
}

// Make a new directory in root directory with verification
func makeDirectory(root string, dirName string) error {
	foldersLog.Debug("create directory requested", "root", root, "dir", dirName)
	clean := filepath.Clean(filepath.FromSlash(dirName))
	if strings.HasPrefix(clean, "..") {
		return errors.New("Directory starts with ..")
	}
	rootAbs, _ := filepath.Abs(root)
	destAbs, err := filepath.Abs(filepath.Join(root, clean))
	if err != nil || !strings.HasPrefix(destAbs, rootAbs+string(os.PathSeparator)) {
		return errors.New("Directory not in root directory")
	}
	if err := os.MkdirAll(destAbs, 0755); err != nil {
		return errors.New("Error creating directory")
	}
	foldersLog.Info("directory created", "path", filepath.ToSlash(clean))
	return nil
}

func removeDirectory(root string, dirName string) error {
	foldersLog.Debug("remove directory requested", "root", root, "dir", dirName)
	clean := filepath.Clean(filepath.FromSlash(dirName))
	if clean == "." || strings.HasPrefix(clean, "..") {
		return errors.New("Directory starts with ..")
	}
	rootAbs, _ := filepath.Abs(root)
	dirAbs, err := filepath.Abs(filepath.Join(root, clean))
	if err != nil || !strings.HasPrefix(dirAbs, rootAbs+string(os.PathSeparator)) {
		return errors.New("Directory not in root directory")
	}

	// Move all files inside the directory to root
	err = filepath.WalkDir(dirAbs, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		dest := filepath.Join(root, filepath.Base(path))
		// Avoid overwriting existing files at root
		if _, statErr := os.Stat(dest); statErr == nil {
			ext := filepath.Ext(filepath.Base(path))
			base := strings.TrimSuffix(filepath.Base(path), ext)
			dest = filepath.Join(root, base+"_1"+ext)
		}
		return os.Rename(path, dest)
	})
	if err != nil {
		return err
	}

	if err := os.RemoveAll(dirAbs); err != nil {
		return err
	}

	foldersLog.Info("directory removed", "path", filepath.ToSlash(clean))
	return nil
}

func moveFileToDirectory(root string, path string, dest string) error {
	foldersLog.Debug("move file requested", "root", root, "path", path, "dest", dest)
	srcClean := filepath.Clean(filepath.FromSlash(path))
	if strings.HasPrefix(srcClean, "..") {
		return errors.New("forbidden: path attempts directory traversal")
	}
	rootAbs, _ := filepath.Abs(root)
	srcAbs, err := filepath.Abs(filepath.Join(root, srcClean))
	if err != nil || !strings.HasPrefix(srcAbs, rootAbs+string(os.PathSeparator)) {
		return errors.New("forbidden: source path not in root")
	}

	destFolder := filepath.Clean(filepath.FromSlash(dest))
	if destFolder == "." || destFolder == "/" {
		destFolder = ""
	}
	if strings.HasPrefix(destFolder, "..") {
		return errors.New("forbidden: dest attempts directory traversal")
	}
	destDir := filepath.Join(root, destFolder)
	destDirAbs, err := filepath.Abs(destDir)
	if err != nil || (!strings.HasPrefix(destDirAbs, rootAbs+string(os.PathSeparator)) && destDirAbs != rootAbs) {
		return errors.New("forbidden: dest dir not in root")
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return errors.New("cannot create destination directory")
	}
	destAbs := filepath.Join(destDir, filepath.Base(srcAbs))
	if err := os.Rename(srcAbs, destAbs); err != nil {
		return err
	}
	foldersLog.Info("file moved", "from", filepath.ToSlash(srcClean), "to", filepath.ToSlash(filepath.Join(destFolder, filepath.Base(srcAbs))))
	return nil
}
