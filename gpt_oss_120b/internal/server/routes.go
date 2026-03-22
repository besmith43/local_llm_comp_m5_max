package server

import (
	"net/http"
	"plex-importor/internal/views"
	"plex-importor/internal/handlers"


	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

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

	// Register API routes
	handler := "github.com/plex-importor/internal/handlers"
	e.GET("/api/folders", handler.ListSourceFolders)
	e.GET("/api/tvseries", handler.GetTVSeries)
	e.POST("/api/submit", handler.ProcessSelection)
	return e
}
