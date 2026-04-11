package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"plex-importor/internal/views"
)

type movePayload struct {
	Type      string `json:"type"`
	Title     string `json:"title,omitempty"`
	Year      int    `json:"year,omitempty"`
	Series    string `json:"series,omitempty"`
	Season    int    `json:"season,omitempty"`
	Episode   int    `json:"episode,omitempty"`
	SourcePath string `json:"source_path"`
	Ext       string `json:"ext"`
}
type moveResult struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
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

	// Serve index.html at root (redirect to assets)
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, "/assets/index.html")
	})

	// POST /scan – return list of folders with video files
	e.POST("/scan", func(c echo.Context) error {
		src := os.Getenv("SOURCE")
		if src == "" {
			src = "./source"
		}
		srcAbs, err := filepath.Abs(src)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		entries, err := os.ReadDir(srcAbs)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		type folderInfo struct {
			FolderName string `json:"folder_name"`
			Videos []struct {
				RelativePath string `json:"relative_path"`
				FileName     string `json:"file_name"`
			} `json:"videos"`
		}
		var folders []folderInfo
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			folderPath := filepath.Join(srcAbs, e.Name())
			var videos []struct {
				RelativePath string `json:"relative_path"`
				FileName     string `json:"file_name"`
			}
			subEntries, err := os.ReadDir(folderPath)
			if err != nil {
				continue
			}
			for _, sub := range subEntries {
				if sub.IsDir() {
					continue
				}
				ext := strings.ToLower(filepath.Ext(sub.Name()))
				if ext == ".mp4" || ext == ".mkv" || ext == ".avi" || ext == ".mov" {
					rel, _ := filepath.Rel(srcAbs, filepath.Join(folderPath, sub.Name()))
					videos = append(videos, struct {
						RelativePath string `json:"relative_path"`
						FileName     string `json:"file_name"`
					}{RelativePath: rel, FileName: sub.Name()})
				}
			}
			if len(videos) > 0 {
				folders = append(folders, folderInfo{
					FolderName: e.Name(),
					Videos: videos,
				})
			}
		}
		return c.JSON(http.StatusOK, folders)
	})

	// POST /move – rename/move the selected file according to Plex naming
	e.POST("/move", func(c echo.Context) error {
		var p movePayload
		if err := c.Bind(&p); err != nil {
			return c.JSON(http.StatusBadRequest, moveResult{Error: "invalid payload"})
		}
		// Resolve source path relative to SOURCE
		srcRoot := os.Getenv("SOURCE")
		if srcRoot == "" {
			srcRoot = "./source"
		}
		srcAbs, err := filepath.Abs(filepath.Join(srcRoot, p.SourcePath))
		if err != nil {
			return c.JSON(http.StatusBadRequest, moveResult{Error: "invalid source path"})
		}
		if _, err := os.Stat(srcAbs); os.IsNotExist(err) {
			return c.JSON(http.StatusBadRequest, moveResult{Error: "source file does not exist"})
		}
		// Normalise extension
		if !strings.HasPrefix(p.Ext, ".") {
			p.Ext = "." + p.Ext
		}
		// Destination root
		destRoot := os.Getenv("DESTINATION")
		if destRoot == "" {
			destRoot = "./dest"
		}
		destAbs, err := filepath.Abs(destRoot)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, moveResult{Error: "invalid destination"})
		}
		var dest string
		switch p.Type {
		case "movie":
			if p.Title == "" || p.Year == 0 {
				return c.JSON(http.StatusBadRequest, moveResult{Error: "title and year required for movies"})
			}
			title := strings.ReplaceAll(p.Title, "/", "")
			dest = filepath.Join(destAbs, fmt.Sprintf("%s (%d)%s", strings.TrimSpace(title), p.Year, p.Ext))
		case "tv_show":
			if p.Series == "" || p.Season == 0 || p.Episode == 0 {
				return c.JSON(http.StatusBadRequest, moveResult{Error: "series, season, and episode required for TV shows"})
			}
			series := strings.ReplaceAll(p.Series, "/", "")
			seasonDir := filepath.Join(destAbs, series, fmt.Sprintf("Season %02d", p.Season))
			if err := os.MkdirAll(seasonDir, 0755); err != nil {
				return c.JSON(http.StatusInternalServerError, moveResult{Error: err.Error()})
			}
			dest = filepath.Join(seasonDir, fmt.Sprintf("Episode %02d%s", p.Episode, p.Ext))
		default:
			return c.JSON(http.StatusBadRequest, moveResult{Error: "invalid type"})
		}
		// Ensure destination directory exists
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return c.JSON(http.StatusInternalServerError, moveResult{Error: err.Error()})
		}
		// Move the file
		if err := os.Rename(srcAbs, dest); err != nil {
			return c.JSON(http.StatusInternalServerError, moveResult{Error: err.Error()})
		}
		return c.JSON(http.StatusOK, moveResult{Status: "moved"})
	})

	// GET /series – list existing TV series folders under DESTINATION/TV Show
	e.GET("/series", func(c echo.Context) error {
		destRoot := os.Getenv("DESTINATION")
		if destRoot == "" {
			destRoot = "./dest"
		}
		destAbs, err := filepath.Abs(filepath.Join(destRoot, "TV Show"))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		entries, err := os.ReadDir(destAbs)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		var series []string
		for _, e := range entries {
			if e.IsDir() {
				series = append(series, e.Name())
			}
		}
		return c.JSON(http.StatusOK, series)
	})

	return e
}
