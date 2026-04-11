package handlers

import (
	"net/http"
	"plex-importor/internal/services"
	"plex-importor/internal/views"
	"strings"

	"github.com/labstack/echo/v4"
)

// Index renders the main page
func Index(c echo.Context) error {
	return views.Plex().Render(c.Request().Context(), c.Response())
}

func GetFolders(c echo.Context) error {
	folders, err := services.ScanSourceDir()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to scan source directory",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"folders": folders,
	})
}

func GetTVShows(c echo.Context) error {
	shows, err := services.GetTVShowFolders()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get TV show folders",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"tv_shows": shows,
	})
}

type MoveRequest struct {
	VideoPath     string `json:"video_path"`
	Type          string `json:"type"`
	MovieTitle    string `json:"movie_title"`
	MovieYear     string `json:"movie_year"`
	TVShowTitle   string `json:"tv_show_title"`
	IsNewShow     bool   `json:"is_new_show"`
	NewShowTitle  string `json:"new_show_title"`
	SeasonNumber  int    `json:"season_number"`
	EpisodeNumber int    `json:"episode_number"`
	EpisodeTitle  string `json:"episode_title"`
}

func MoveVideo(c echo.Context) error {
	var req MoveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request data",
		})
	}

	if req.VideoPath == "" || req.Type == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Video path and type are required",
		})
	}

	switch req.Type {
	case "movie":
		title, year, valid := services.ValidateMovieInput(req.MovieTitle, req.MovieYear)
		if !valid {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid movie title or year. Year must be between 1888 and 2100.",
			})
		}

		err := services.MoveMovie(req.VideoPath, title, year)
		if err != nil {
			if err.Error() == "file already exists, skipped" {
				return c.JSON(http.StatusConflict, map[string]string{
					"error": "A movie with that title already exists. Skipped.",
				})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to move movie: " + err.Error(),
			})
		}

	case "tv_show":
		showTitle, season, episode, valid := services.ValidateTVShowInput(
			req.TVShowTitle, req.IsNewShow, req.NewShowTitle, req.SeasonNumber, req.EpisodeNumber,
		)
		if !valid {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid TV show input. Check show name, season, and episode numbers.",
			})
		}

		episodeTitle := req.EpisodeTitle
		if episodeTitle == "" {
			episodeTitle = req.VideoPath
		}
		episodeTitle = sanitizeEpisodeTitle(episodeTitle)

		err := services.MoveTVShow(req.VideoPath, showTitle, episodeTitle, season, episode)
		if err != nil {
			if err.Error() == "file already exists, skipped" {
				return c.JSON(http.StatusConflict, map[string]string{
					"error": "An episode with that name already exists. Skipped.",
				})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to move TV show: " + err.Error(),
			})
		}

	default:
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid type. Must be 'movie' or 'tv_show'",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Success",
	})
}

func sanitizeEpisodeTitle(title string) string {
	ext := ""
	if dotIdx := strings.LastIndex(title, "."); dotIdx > 0 {
		title = title[:dotIdx]
		ext = title[dotIdx:]
	}
	return services.SanitizeFilename(title) + ext
}
