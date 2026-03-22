package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Video extensions supported
var videoExtensions = map[string]bool{
	"mp4":  true,
	"mkv":  true,
	"avi":  true,
}

// FolderInfo represents a folder containing video files
type FolderInfo struct {
	Name      string   `json:"name"`
	Path      string   `json:"path"`
	VideoFile string   `json:"videoFile"`
}

// ScanFolders scans the source directory recursively and returns folders containing video files
func ScanFolders(sourcePath string) ([]FolderInfo, error) {
	var folders []FolderInfo

	// Check if source path exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("source path does not exist: %s", sourcePath)
	}

	// Walk through the directory recursively
	err := filepath.WalkDir(sourcePath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == sourcePath {
			return nil
		}

		// Only process directories
		if !d.IsDir() {
			return nil
		}

		// Check if directory contains any video files
		videoFiles, err := findVideoFiles(path)
		if err != nil {
			return nil // Skip directories with errors
		}

		if len(videoFiles) > 0 {
			// Get the folder name
			folderName := filepath.Base(path)

			// Add to folders list
			folders = append(folders, FolderInfo{
				Name:      folderName,
				Path:      path,
				VideoFile: videoFiles[0], // Use first video file
			})
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error scanning directories: %w", err)
	}

	return folders, nil
}

// findVideoFiles finds all video files in a directory
func findVideoFiles(dirPath string) ([]string, error) {
	var videoFiles []string

	err := filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Check if file has a video extension
		ext := strings.ToLower(filepath.Ext(path))
		if videoExtensions[ext] {
			videoFiles = append(videoFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return videoFiles, nil
}

// GetVideoExtensions returns the list of supported video extensions
func GetVideoExtensions() []string {
	extensions := make([]string, 0, len(videoExtensions))
	for ext := range videoExtensions {
		extensions = append(extensions, ext)
	}
	return extensions
}

// Scanner is the main scanner type
type Scanner struct {
	sourcePath string
}

// NewScanner creates a new Scanner instance
func NewScanner(sourcePath string) *Scanner {
	return &Scanner{
		sourcePath: sourcePath,
	}
}

// ScanFolders scans the source directory recursively and returns folders containing video files
func (s *Scanner) ScanFolders() ([]scanner.FolderInfo, error) {
	return ScanFolders(s.sourcePath)
}