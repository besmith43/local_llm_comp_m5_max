package handler

import (
	"fmt"
	"net/http"
	"plex-importor/internal/mover"
	"strconv"

	"github.com/labstack/echo/v4"
)

// MetadataRequest represents the request body for metadata processing
type MetadataRequest struct {
	ContentType string `json:"contentType"` // "movie" or "tv"
	FolderPath  string `json:"folderPath"`
	Title       string `json:"title"`
	Year        string `json:"year"`
	SeriesName  string `json:"seriesName"`
	Season      string `json:"season"`
	Episode     string `json:"episode"`
}

// MetadataHandler handles metadata-related API requests
type MetadataHandler struct {
	mover *mover.Mover
}

// NewMetadataHandler creates a new MetadataHandler instance
func NewMetadataHandler(mover *mover.Mover) *MetadataHandler {
	return &MetadataHandler{
		mover: mover,
	}
}

// GetFolders handles GET /api/folders
// Returns a list of folders containing video files
func (h *MetadataHandler) GetFolders(c echo.Context) error {
	// For now, return an empty list or implement proper directory scanning
	// In production, you'd use the scanner module
	return c.JSON(http.StatusOK, []interface{}{})
}

// GetConfig handles GET /api/config
// Returns the configuration (source and destination paths)
func (h *MetadataHandler) GetConfig(c echo.Context) error {
	config := map[string]string{
		"sourcePath":      c.Get("sourcePath").(string),
		"destinationPath": c.Get("destinationPath").(string),
	}
	return c.JSON(http.StatusOK, config)
}

// ProcessMetadata handles POST /api/metadata
// Processes metadata and moves the video file to the appropriate destination
func (h *MetadataHandler) ProcessMetadata(c echo.Context) error {
	var req MetadataRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Validate content type
	if req.ContentType != "movie" && req.ContentType != "tv" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "contentType must be 'movie' or 'tv'",
		})
	}

	// Validate required fields based on content type
	if req.ContentType == "movie" {
		if req.Title == "" || req.Year == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "title and year are required for movies",
			})
		}
	} else {
		if req.SeriesName == "" || req.Season == "" || req.Episode == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "seriesName, season, and episode are required for TV shows",
			})
		}
	}

	// Process the metadata and move the file
	result, err := h.mover.MoveFile(
		req.FolderPath,
		req.ContentType,
		req.Title,
		req.Year,
		req.SeriesName,
		req.Season,
		req.Episode,
	)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// Return the result
	if result.Success {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success":   true,
			"message":   result.Message,
			"destPath":  result.DestPath,
			"isDuplicate": result.IsDuplicate,
		})
	} else {
		return c.JSON(http.StatusConflict, map[string]interface{}{
			"success":   false,
			"message":   result.Message,
			"isDuplicate": result.IsDuplicate,
		})
	}
}

// GetTVSeriesList handles GET /api/tv-series
// Returns a list of existing TV series in the destination directory
func (h *MetadataHandler) GetTVSeriesList(c echo.Context) error {
	destinationPath := c.Get("destinationPath").(string)

	// Get all TV series directories
	tvSeriesPath := mover.NewMover(destinationPath).GetDestinationPathForTVSeries("")
	// Note: This is a simplified version - in production you'd want to scan the directory properly

	// For now, return an empty list or implement proper directory scanning
	return c.JSON(http.StatusOK, []string{})
}

// GetSeasonsForSeries handles GET /api/seasons/:seriesName
// Returns a list of seasons for a specific TV series
func (h *MetadataHandler) GetSeasonsForSeries(c echo.Context) error {
	seriesName := c.Param("seriesName")
	destinationPath := c.Get("destinationPath").(string)

	m := mover.NewMover(destinationPath)
	seasonPath := m.GetSeasonPath(seriesName, "")

	// Check if season directory exists
	if m.FileExists(seasonPath) {
		return c.JSON(http.StatusOK, []string{})
	}

	// For now, return a default season list
	// In production, you'd scan the directory to get actual seasons
	return c.JSON(http.StatusOK, []string{"1", "2", "3"})
}

// GetEpisodesForSeason handles GET /api/episodes/:seriesName/:season
// Returns a list of episodes for a specific TV series and season
func (h *MetadataHandler) GetEpisodesForSeason(c echo.Context) error {
	seriesName := c.Param("seriesName")
	season := c.Param("season")
	destinationPath := c.Get("destinationPath").(string)

	m := mover.NewMover(destinationPath)
	seasonPath := m.GetSeasonPath(seriesName, season)

	// Check if season directory exists
	if m.FileExists(seasonPath) {
		return c.JSON(http.StatusOK, []string{})
	}

	// For now, return a default episode list
	// In production, you'd scan the directory to get actual episodes
	return c.JSON(http.StatusOK, []string{"1", "2", "3"})
}

// ParseSeasonEpisode parses season and episode numbers from strings
func ParseSeasonEpisode(seasonStr, episodeStr string) (int, int, error) {
	season, err := strconv.Atoi(seasonStr)
	if err != nil {
		return 0, 0, err
	}

	episode, err := strconv.Atoi(episodeStr)
	if err != nil {
		return 0, 0, err
	}

	return season, episode, nil
}

// FormatSeasonEpisode formats season and episode numbers for display
func FormatSeasonEpisode(season, episode int) string {
	return fmt.Sprintf("S%02dE%02d", season, episode)
}