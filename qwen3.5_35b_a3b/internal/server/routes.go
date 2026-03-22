package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"plex-importor/internal/views"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Video extensions to consider valid
var videoExtensions = map[string]bool{
	".mp4": true, ".mkv": true, ".avi": true,
}

// API Response structs
type DirectoryResponse struct {
	Directories []string `json:"directories"`
}

type TVShowsResponse struct {
	Shows []string `json:"shows"`
}

type MovieRequest struct {
	SourceDir  string `json:"sourceDir"`
	Filename   string `json:"filename"`
	Title      string `json:"title"`
	Year       string `json:"year"`
	Extension  string `json:"extension"`
}

type TVShowRequest struct {
	SourceDir    string `json:"sourceDir"`
	Filename     string `json:"filename"`
	SeriesTitle  string `json:"seriesTitle"`
	Season       int    `json:"season"`
	Episode      int    `json:"episode"`
	NewSeries    bool   `json:"newSeries"`
	Extension    string `json:"extension"`
}

// isValidVideoFile checks if a file has a valid video extension
func isValidVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return videoExtensions[ext]
}

// getDirectories returns all subdirectories in source that contain video files
func getDirectories(sourcePath string) ([]string, error) {
	entries, err := os.ReadDir(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read source directory: %w", err)
	}

	var dirs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Check if directory contains video files
		dirPath := filepath.Join(sourcePath, entry.Name())
		dirEntries, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}
		for _, fileEntry := range dirEntries {
			if !fileEntry.IsDir() && isValidVideoFile(fileEntry.Name()) {
				dirs = append(dirs, entry.Name())
				break
			}
		}
	}
	return dirs, nil
}

// getTVShows returns existing TV show directories in destination
func getTVShows(destPath string) ([]string, error) {
	tvShowsPath := filepath.Join(destPath, "TV Shows")
	entries, err := os.ReadDir(tvShowsPath)
	if err != nil {
		return []string{}, nil // Return empty if directory doesn't exist
	}

	var shows []string
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "placeholder" {
			shows = append(shows, entry.Name())
		}
	}
	return shows, nil
}

// moveFile moves a file from source to destination (handles cross-device)
func moveFile(srcPath, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create destination directory if needed
	dstDir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Remove existing file at destination
	os.Remove(dstPath)

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Remove source file after successful copy
	os.Remove(srcPath)
	return nil
}

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		HTML5:      true,
		Root:       "assets",
		Filesystem: http.FS(views.Files),
	}))

	// Main Page Route
	e.GET("/", func(c echo.Context) error {
		w := &bytes.Buffer{}
		views.Plex().Render(c.Request().Context(), w)
		return c.HTMLBlob(http.StatusOK, w.Bytes())
	})

	// API Routes
	e.GET("/api/directories", func(c echo.Context) error {
		sourcePath := os.Getenv("SOURCE")
		dirs, err := getDirectories(sourcePath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, DirectoryResponse{Directories: dirs})
	})

	e.GET("/api/tv-shows", func(c echo.Context) error {
		destPath := os.Getenv("DESTINATION")
		shows, err := getTVShows(destPath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, TVShowsResponse{Shows: shows})
	})

	e.GET("/api/files", func(c echo.Context) error {
		dir := c.QueryParam("dir")
		if dir == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Directory parameter required"})
		}

		sourcePath := os.Getenv("SOURCE")
		dirPath := filepath.Join(sourcePath, dir)

		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to read directory"})
		}

		var files []string
		for _, entry := range entries {
			if !entry.IsDir() && isValidVideoFile(entry.Name()) {
				files = append(files, entry.Name())
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"files": files,
		})
	})

	e.POST("/api/move-movie", func(c echo.Context) error {
		var req MovieRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}

		sourcePath := os.Getenv("SOURCE")
		destPath := os.Getenv("DESTINATION")

		srcFile := filepath.Join(sourcePath, req.SourceDir, req.Filename)
		dstFilename := fmt.Sprintf("%s (%s).%s", sanitizeFileName(req.Title), req.Year, req.Extension)
		dstFile := filepath.Join(destPath, "Movies", dstFilename)

		if err := moveFile(srcFile, dstFile); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Movie moved successfully"})
	})

	e.POST("/api/move-tvshow", func(c echo.Context) error {
		var req TVShowRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}

		sourcePath := os.Getenv("SOURCE")
		destPath := os.Getenv("DESTINATION")

		seriesTitle := req.SeriesTitle
		if req.NewSeries && seriesTitle == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Series title required for new series"})
		}

		srcFile := filepath.Join(sourcePath, req.SourceDir, req.Filename)
		newSeriesName := sanitizeFileName(seriesTitle)
		dstFilename := fmt.Sprintf("%s S%02dE%02d.%s", newSeriesName, req.Season, req.Episode, req.Extension)
		dstFile := filepath.Join(destPath, "TV Shows", newSeriesName, fmt.Sprintf("Season %02d", req.Season), dstFilename)

		if err := moveFile(srcFile, dstFile); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "TV show episode moved successfully"})
	})

	return e
}

// sanitizeFileName removes or replaces invalid filename characters
func sanitizeFileName(name string) string {
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, ch := range invalidChars {
		result = strings.ReplaceAll(result, ch, "")
	}
	return strings.TrimSpace(result)
}
