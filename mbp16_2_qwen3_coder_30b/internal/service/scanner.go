package service

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Video extensions to scan for
var videoExtensions = map[string]bool{
	".mp4": true, ".mkv": true, ".avi": true,
	".mov": true, ".wmv": true, ".mpeg": true, ".mpg": true,
}

type VideoFile struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type VideoFolder struct {
	ID    string      `json:"id"` // URL-safe encoded path
	Name  string      `json:"name"`
	Path  string      `json:"path"`
	Files []VideoFile `json:"files"`
}

// ScanSourceDirectory walks the source directory and returns folders containing video files
func ScanSourceDirectory(sourcePath string) ([]VideoFolder, error) {
	var folders []VideoFolder

	err := filepath.WalkDir(sourcePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip non-directories
		if !d.IsDir() {
			return nil
		}

		// Skip the root source directory itself
		if path == sourcePath {
			return nil
		}

		// Check if this directory contains video files
		videoFiles, err := scanDirectoryForVideos(path)
		if err != nil {
			return err
		}

		// Only include folders that have video files
		if len(videoFiles) > 0 {
			folder := VideoFolder{
				ID:    encodePath(path),
				Name:  d.Name(),
				Path:  path,
				Files: videoFiles,
			}
			folders = append(folders, folder)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort folders by name for consistent display
	sort.Slice(folders, func(i, j int) bool {
		return strings.ToLower(folders[i].Name) < strings.ToLower(folders[j].Name)
	})

	return folders, nil
}

// scanDirectoryForVideos scans a single directory for video files
func scanDirectoryForVideos(dirPath string) ([]VideoFile, error) {
	var videoFiles []VideoFile

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if videoExtensions[ext] {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			videoFiles = append(videoFiles, VideoFile{
				Path: filepath.Join(dirPath, entry.Name()),
				Name: entry.Name(),
				Size: info.Size(),
			})
		}
	}

	// Sort by size descending (largest first)
	sort.Slice(videoFiles, func(i, j int) bool {
		return videoFiles[i].Size > videoFiles[j].Size
	})

	return videoFiles, nil
}
