package server

import (
	"net/http"
	"plex-importor/internal/views"

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

	// API routes
	e.GET("/api/config", s.handler.GetConfig)
	e.GET("/api/folders", s.handler.GetFolders)
	e.POST("/api/metadata", s.handler.ProcessMetadata)
	e.GET("/api/tv-series", s.handler.GetTVSeriesList)
	e.GET("/api/seasons/:seriesName", s.handler.GetSeasonsForSeries)
	e.GET("/api/episodes/:seriesName/:season", s.handler.GetEpisodesForSeason)

	// Static file serving
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		HTML5:      true,
		Root:       "assets",
		Filesystem: http.FS(views.Files),
	}))

	return e
}
