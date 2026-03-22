package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"plex-importor/internal/models"
	"plex-importor/internal/services"
	"plex-importor/internal/views"
)

type Handler struct {
	fileService *services.FileService
}

func NewHandler() *Handler {
	return &Handler{
		fileService: services.NewFileService(),
	}
}

// Index handles the main page request
func (h *Handler) Index(c echo.Context) error {
	return views.Plex().Render(c.Request().Context(), c.Response())
}

// GetFolders returns a JSON list of folders containing video files
func (h *Handler) GetFolders(c echo.Context) error {
	folders, err := h.fileService.ScanSourceFolders()
	if err != nil {
		log.Printf("Error scanning folders: %v", err)
		return c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to scan source directory",
		})
	}

	return c.JSON(http.StatusOK, models.FoldersResponse{
		APIResponse: models.APIResponse{
			Success: true,
		},
		Folders: folders,
	})
}

// GetTVShows returns a JSON list of existing TV shows in the destination directory
func (h *Handler) GetTVShows(c echo.Context) error {
	shows, err := h.fileService.GetExistingTVShows()
	if err != nil {
		log.Printf("Error getting TV shows: %v", err)
		return c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to get TV shows",
		})
	}

	return c.JSON(http.StatusOK, models.TVShowsResponse{
		APIResponse: models.APIResponse{
			Success: true,
		},
		Shows: shows,
	})
}

// MoveFile handles the form submission to move a video file
func (h *Handler) MoveFile(c echo.Context) error {
	var req models.MoveRequest

	// Bind the request body to the MoveRequest struct
	if err := c.Bind(&req); err != nil {
		log.Printf("Error binding request: %v", err)
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request data",
		})
	}

	// Validate the request
	if err := h.fileService.ValidateMoveRequest(req); err != nil {
		log.Printf("Validation error: %v", err)
		return c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	// Move the file
	if err := h.fileService.MoveVideoFile(req); err != nil {
		log.Printf("Error moving file: %v", err)
		return c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "File moved successfully",
	})
}

// Helper function to send JSON response
func jsonResponse(c echo.Context, statusCode int, data interface{}) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.Response().WriteHeader(statusCode)
	return json.NewEncoder(c.Response()).Encode(data)
}
