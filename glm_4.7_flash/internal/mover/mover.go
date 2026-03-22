package mover

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MoveResult represents the result of a file move operation
type MoveResult struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	SourcePath   string `json:"sourcePath"`
	DestPath     string `json:"destPath"`
	IsDuplicate  bool   `json:"isDuplicate"`
}

// Mover handles file movement operations
type Mover struct {
	destinationPath string
}

// NewMover creates a new Mover instance
func NewMover(destinationPath string) *Mover {
	return &Mover{
		destinationPath: destinationPath,
	}
}

// MoveFile moves a video file to the appropriate Plex directory structure
func (m *Mover) MoveFile(sourcePath, contentType, title, year, seriesName, season, episode string) (*MoveResult, error) {
	// Validate inputs
	if contentType == "" {
		return &MoveResult{
			Success: false,
			Message: "Content type is required",
		}, nil
	}

	// Get file info
	fileInfo, err := os.Stat(sourcePath)
	if err != nil {
		return &MoveResult{
			Success: false,
			Message: fmt.Sprintf("Error accessing source file: %v", err),
		}, nil
	}

	// Get destination path based on content type
	destPath, err := m.getDestinationPath(contentType, title, year, seriesName, season, episode, fileInfo.Name())
	if err != nil {
		return &MoveResult{
			Success: false,
			Message: fmt.Sprintf("Error creating destination path: %v", err),
		}, nil
	}

	// Check if file already exists
	if m.FileExists(destPath) {
		return &MoveResult{
			Success:    false,
			Message:    "File already exists in destination",
			IsDuplicate: true,
			DestPath:   destPath,
		}, nil
	}

	// Create destination directory if needed
	err = os.MkdirAll(filepath.Dir(destPath), 0755)
	if err != nil {
		return &MoveResult{
			Success: false,
			Message: fmt.Sprintf("Error creating destination directory: %v", err),
		}, nil
	}

	// Move the file
	err = os.Rename(sourcePath, destPath)
	if err != nil {
		return &MoveResult{
			Success: false,
			Message: fmt.Sprintf("Error moving file: %v", err),
		}, nil
	}

	return &MoveResult{
		Success:   true,
		Message:   "File moved successfully",
		SourcePath: sourcePath,
		DestPath:   destPath,
	}, nil
}

// getDestinationPath generates the Plex-compatible destination path
func (m *Mover) getDestinationPath(contentType, title, year, seriesName, season, episode string, filename string) (string, error) {
	// Sanitize filenames
	title = sanitizeFilename(title)
	seriesName = sanitizeFilename(seriesName)

	switch contentType {
	case "movie":
		// Movies: {Title} ({Year})/{Title} ({Year}){Extension}
		if title == "" || year == "" {
			return "", fmt.Errorf("title and year are required for movies")
		}

		movieDir := filepath.Join(m.destinationPath, fmt.Sprintf("%s (%s)", title, year))
		destFile := filepath.Join(movieDir, fmt.Sprintf("%s (%s)%s", title, year, filepath.Ext(filename)))

		return destFile, nil

	case "tv":
		// TV Shows: TV Shows/{Series Name}/Season {Season}/{Series Name} - S{Season}E{Episode}{Extension}
		if seriesName == "" || season == "" || episode == "" {
			return "", fmt.Errorf("series name, season, and episode are required for TV shows")
		}

		tvDir := filepath.Join(m.destinationPath, "TV Shows", seriesName, fmt.Sprintf("Season %s", season))
		destFile := filepath.Join(tvDir, fmt.Sprintf("%s - S%sE%s%s", seriesName, season, episode, filepath.Ext(filename)))

		return destFile, nil

	default:
		return "", fmt.Errorf("invalid content type: %s", contentType)
	}
}

// FileExists checks if a file exists at the given path
func (m *Mover) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// sanitizeFilename removes or replaces characters that are invalid in filenames
func sanitizeFilename(name string) string {
	// Replace invalid characters with underscores
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		name = strings.ReplaceAll(name, char, "_")
	}

	// Trim leading/trailing spaces and dots
	name = strings.Trim(name, " .")

	return name
}

// GetDestinationPathForTVSeries returns the path for a TV series directory
func (m *Mover) GetDestinationPathForTVSeries(seriesName string) string {
	seriesName = sanitizeFilename(seriesName)
	return filepath.Join(m.destinationPath, "TV Shows", seriesName)
}

// GetSeasonPath returns the path for a TV season directory
func (m *Mover) GetSeasonPath(seriesName, season string) string {
	seriesName = sanitizeFilename(seriesName)
	return filepath.Join(m.destinationPath, "TV Shows", seriesName, fmt.Sprintf("Season %s", season))
}