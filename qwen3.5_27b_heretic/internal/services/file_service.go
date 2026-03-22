package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"plex-importor/internal/models"
)

// FileService handles file operations for the Plex Importor
type FileService struct {
	sourceDir      string
	destinationDir string
}

// NewFileService creates a new FileService instance
func NewFileService() *FileService {
	return &FileService{
		sourceDir:      os.Getenv("SOURCE"),
		destinationDir: os.Getenv("DESTINATION"),
	}
}

// ScanSourceFolders recursively scans the source directory and returns folders containing video files
func (fs *FileService) ScanSourceFolders() ([]models.FolderInfo, error) {
	var folders []models.FolderInfo

	err := filepath.WalkDir(fs.sourceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root source directory itself
		if path == fs.sourceDir {
			return nil
		}

		// Only process directories
		if !d.IsDir() {
			return nil
		}

		folderInfo, hasVideos := fs.scanFolderForVideos(path)
		if hasVideos {
			folders = append(folders, folderInfo)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error scanning source directory: %w", err)
	}

	return folders, nil
}

// scanFolderForVideos scans a single folder for video files
func (fs *FileService) scanFolderForVideos(folderPath string) (models.FolderInfo, bool) {
	folderName := filepath.Base(folderPath)
	var videoFiles []models.VideoFile
	var totalSize int64

	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return models.FolderInfo{}, false
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		ext := strings.ToLower(filepath.Ext(fileName))

		for _, videoExt := range models.VideoExtensions {
			if ext == videoExt {
				info, err := entry.Info()
				if err != nil {
					continue
				}

				videoFiles = append(videoFiles, models.VideoFile{
					Name: fileName,
					Path: filepath.Join(folderPath, fileName),
					Size: info.Size(),
				})
				totalSize += info.Size()
				break
			}
		}
	}

	if len(videoFiles) > 0 {
		return models.FolderInfo{
			Path:       folderPath,
			Name:       folderName,
			VideoFiles: videoFiles,
			FileCount:  len(videoFiles),
			TotalSize:  totalSize,
		}, true
	}

	return models.FolderInfo{}, false
}

// GetExistingTVShows returns a list of existing TV show directories from the destination
func (fs *FileService) GetExistingTVShows() ([]string, error) {
	tvShowsDir := filepath.Join(fs.destinationDir, "TV Shows")

	// Check if TV Shows directory exists
	if _, err := os.Stat(tvShowsDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	var shows []string

	entries, err := os.ReadDir(tvShowsDir)
	if err != nil {
		return nil, fmt.Errorf("error reading TV Shows directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			shows = append(shows, entry.Name())
		}
	}

	return shows, nil
}

// MoveVideoFile moves a video file to the destination with proper Plex naming
func (fs *FileService) MoveVideoFile(req models.MoveRequest) error {
	var destPath string
	var err error

	switch req.SourceType {
	case "movie":
		destPath, err = fs.moveMovie(req)
	case "tv_show":
		destPath, err = fs.moveTVShow(req)
	default:
		return fmt.Errorf("unknown source type: %s", req.SourceType)
	}

	if err != nil {
		return err
	}

	// Move the file
	err = os.Rename(req.SourcePath, destPath)
	if err != nil {
		return fmt.Errorf("error moving file: %w", err)
	}

	return nil
}

// moveMovie handles moving a movie file to the Movies directory
func (fs *FileService) moveMovie(req models.MoveRequest) (string, error) {
	moviesDir := filepath.Join(fs.destinationDir, "Movies")

	// Create Movies directory if it doesn't exist
	err := os.MkdirAll(moviesDir, 0755)
	if err != nil {
		return "", fmt.Errorf("error creating Movies directory: %w", err)
	}

	// Get the file extension
	ext := filepath.Ext(req.SourcePath)

	// Create the destination filename: "Movie Title (Year).ext"
	destFileName := fmt.Sprintf("%s (%d)%s", req.MovieTitle, req.MovieYear, ext)
	destPath := filepath.Join(moviesDir, destFileName)

	// Handle duplicate filenames
	if _, err := os.Stat(destPath); err == nil {
		destPath = fs.getUniqueFilePath(destPath)
	}

	return destPath, nil
}

// moveTVShow handles moving a TV show episode to the proper directory structure
func (fs *FileService) moveTVShow(req models.MoveRequest) (string, error) {
	// Determine the show title (use new show title if provided, otherwise use existing)
	showTitle := req.TVShowTitle
	if req.NewShowTitle != "" {
		showTitle = req.NewShowTitle
	}

	// Create the TV show directory structure: "TV Shows/Show Name/Season XX"
	seasonDir := filepath.Join(fs.destinationDir, "TV Shows", showTitle, fmt.Sprintf("Season %02d", req.SeasonNumber))

	err := os.MkdirAll(seasonDir, 0755)
	if err != nil {
		return "", fmt.Errorf("error creating TV show directory: %w", err)
	}

	// Get the file extension
	ext := filepath.Ext(req.SourcePath)

	// Create the destination filename: "E0X - Episode Title.ext"
	episodePrefix := fmt.Sprintf("E%02d", req.EpisodeNumber)
	destFileName := episodePrefix + ext
	destPath := filepath.Join(seasonDir, destFileName)

	// Handle duplicate filenames
	if _, err := os.Stat(destPath); err == nil {
		destPath = fs.getUniqueFilePath(destPath)
	}

	return destPath, nil
}

// getUniqueFilePath generates a unique file path by appending a counter if the file exists
func (fs *FileService) getUniqueFilePath(path string) string {
	dir := filepath.Dir(path)
	baseName := filepath.Base(path)
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)

	counter := 2
	for {
		newName := fmt.Sprintf("%s %d%s", nameWithoutExt, counter, ext)
		newPath := filepath.Join(dir, newName)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
		counter++
	}
}

// ValidateMoveRequest validates the move request data
func (fs *FileService) ValidateMoveRequest(req models.MoveRequest) error {
	if req.SourceType != "movie" && req.SourceType != "tv_show" {
		return fmt.Errorf("invalid source type")
	}

	if req.SourcePath == "" {
		return fmt.Errorf("source path is required")
	}

	// Check if source file exists
	if _, err := os.Stat(req.SourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist")
	}

	if req.SourceType == "movie" {
		if strings.TrimSpace(req.MovieTitle) == "" {
			return fmt.Errorf("movie title is required")
		}
		if req.MovieYear < 1888 || req.MovieYear > 2030 {
			return fmt.Errorf("invalid movie year")
		}
	} else if req.SourceType == "tv_show" {
		if req.TVShowTitle == "" && req.NewShowTitle == "" {
			return fmt.Errorf("TV show title is required")
		}
		if req.SeasonNumber < 1 {
			return fmt.Errorf("season number must be at least 1")
		}
		if req.EpisodeNumber < 1 {
			return fmt.Errorf("episode number must be at least 1")
		}
	}

	return nil
}

// FormatBytes converts bytes to a human-readable string
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
