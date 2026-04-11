package server

import (
	"net/http"
	"plex-importor/internal/handlers"
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

	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		HTML5:      true,
		Root:       "assets",
		Filesystem: http.FS(views.Files),
	}))

	// API routes
	e.GET("/api/folders", handlers.GetFolders)
	e.GET("/api/tv-shows", handlers.GetTVShows)
	e.POST("/api/move", handlers.MoveVideo)

	// Main page
	e.GET("/", handlers.Index)

	return e
}
