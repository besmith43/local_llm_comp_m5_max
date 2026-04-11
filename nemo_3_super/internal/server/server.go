package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"plex-importor/internal/views"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type Server struct {
	port int
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	NewServer := &Server{
		port: port,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}

// getFoldersHandler returns a list of folders in the source directory that contain video files
func (s *Server) getFoldersHandler(c echo.Context) error {
	sourcePath := os.Getenv("SOURCE")
	if sourcePath == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "SOURCE environment variable not set"})
	}

	folders, err := scanSourceFolders(sourcePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, folders)
}

// getTVShowsHandler returns a list of existing TV show series from the destination directory
func (s *Server) getTVShowsHandler(c echo.Context) error {
	destPath := os.Getenv("DESTINATION")
	if destPath == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "DESTINATION environment variable not set"})
	}

	tvShowsPath := filepath.Join(destPath, "TV Shows")
	series, err := listTVShowSeries(tvShowsPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, series)
}

// moveFileHandler processes a request to move a video file to the Plex-organized destination
func (s *Server) moveFileHandler(c echo.Context) error {
	var req moveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	sourcePath := os.Getenv("SOURCE")
	if sourcePath == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "SOURCE environment variable not set"})
	}

	destPath := os.Getenv("DESTINATION")
	if destPath == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "DESTINATION environment variable not set"})
	}

	folderPath := filepath.Join(sourcePath, req.FolderPath)
	videoFiles, err := getVideoFilesInFolder(folderPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if len(videoFiles) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No video files found in folder"})
	}

	// For simplicity, we'll move the first video file found
	// In a more advanced version, we might handle multiple files or let user choose
	sourceFile := filepath.Join(folderPath, videoFiles[0])

	var destFile string
	var moveErr error

	switch req.MediaType {
	case "movie":
		destFile, moveErr = moveMovie(sourceFile, req.Title, req.Year, destPath)
	case "tvshow":
		destFile, moveErr = moveTVShow(sourceFile, req.SeriesTitle, req.Season, req.Episode, destPath, req.IsNewSeries)
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid media type"})
	}

	if moveErr != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": moveErr.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message":     "File moved successfully",
		"destination": destFile,
	})
}

// Helper functions for file operations
func scanSourceFolders(sourcePath string) ([]folderInfo, error) {
	var folders []folderInfo

	entries, err := os.ReadDir(sourcePath)
	if err != nil {
		return folders, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			folderPath := filepath.Join(sourcePath, entry.Name())
			videoFiles, err := getVideoFilesInFolder(folderPath)
			if err != nil {
				return folders, err
			}
			if len(videoFiles) > 0 {
				folders = append(folders, folderInfo{
					Path:       entry.Name(),
					Name:       entry.Name(),
					VideoFiles: videoFiles,
				})
			}
		}
	}

	return folders, nil
}

func getVideoFilesInFolder(folderPath string) ([]string, error) {
	var videoFiles []string
	videoExts := map[string]bool{
		".mkv": true, ".mp4": true, ".avi": true, ".mov": true,
		".wmv": true, ".flv": true, ".webm": true,
	}

	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return videoFiles, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if videoExts[ext] {
				videoFiles = append(videoFiles, entry.Name())
			}
		}
	}

	return videoFiles, nil
}

func listTVShowSeries(tvShowsPath string) ([]string, error) {
	var series []string

	entries, err := os.ReadDir(tvShowsPath)
	if err != nil {
		// If directory doesn't exist, return empty list (no series yet)
		if os.IsNotExist(err) {
			return series, nil
		}
		return series, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			series = append(series, entry.Name())
		}
	}

	return series, nil
}

func moveMovie(sourceFile, title, year string, destPath string) (string, error) {
	// Format: destination/Movies/<Title> (<Year>)/<Title> (<Year>).<ext>
	movieDir := filepath.Join(destPath, "Movies", fmt.Sprintf("%s (%s)", title, year))
	fileExt := filepath.Ext(sourceFile)
	destFile := filepath.Join(movieDir, fmt.Sprintf("%s (%s)%s", title, year, fileExt))

	// Create directory if it doesn't exist
	if err := os.MkdirAll(movieDir, 0755); err != nil {
		return "", err
	}

	// Move the file
	if err := os.Rename(sourceFile, destFile); err != nil {
		return "", err
	}

	return destFile, nil
}

func moveTVShow(sourceFile, seriesTitle string, season, episode int, destPath string, isNewSeries bool) (string, error) {
	// Format: destination/TV Shows/<Series>/Season <Season>/<Series> - S<season>E<episode>.<ext>
	seasonDir := filepath.Join(destPath, "TV Shows", seriesTitle, fmt.Sprintf("Season %d", season))
	fileExt := filepath.Ext(sourceFile)
	destFile := filepath.Join(seasonDir, fmt.Sprintf("%s - S%02dE%02d%s", seriesTitle, season, episode, fileExt))

	// Create directory if it doesn't exist
	if err := os.MkdirAll(seasonDir, 0755); err != nil {
		return "", err
	}

	// Move the file
	if err := os.Rename(sourceFile, destFile); err != nil {
		return "", err
	}

	return destFile, nil
}

func (s *Server) PlexHandler(c echo.Context) error {
	c.Response().Writer.WriteHeader(http.StatusOK)
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return views.Plex().Render(c.Request().Context(), c.Response().Writer)
}

// Request structures
type folderInfo struct {
	Path       string   `json:"path"`
	Name       string   `json:"name"`
	VideoFiles []string `json:"videoFiles"`
}

type moveRequest struct {
	FolderPath  string `json:"folderPath"`
	MediaType   string `json:"mediaType"` // "movie" or "tvshow"
	Title       string `json:"title,omitempty"`
	Year        string `json:"year,omitempty"`
	SeriesTitle string `json:"seriesTitle,omitempty"`
	Season      int    `json:"season,omitempty"`
	Episode     int    `json:"episode,omitempty"`
	IsNewSeries bool   `json:"isNewSeries,omitempty"`
}
