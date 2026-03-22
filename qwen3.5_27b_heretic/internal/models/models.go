package models

// VideoExtensions are the recognized video file extensions
var VideoExtensions = []string{".mp4", ".mkv", ".avi", ".mov", ".wmv"}

// VideoFile represents a video file found in a folder
type VideoFile struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}

// FolderInfo represents a folder containing video files
type FolderInfo struct {
	Path        string      `json:"path"`
	Name        string      `json:"name"`
	VideoFiles  []VideoFile `json:"video_files"`
	FileCount   int         `json:"file_count"`
	TotalSize   int64       `json:"total_size"`
}

// MoveRequest represents the form data for moving a video file
type MoveRequest struct {
	SourceType    string `json:"source_type" form:"source_type"`       // "movie" or "tv_show"
	SourcePath    string `json:"source_path" form:"source_path"`       // Full path to source video file
	MovieTitle    string `json:"movie_title" form:"movie_title"`       // Movie title
	MovieYear     int    `json:"movie_year" form:"movie_year"`         // Movie release year
	TVShowTitle   string `json:"tv_show_title" form:"tv_show_title"`   // TV show title (for existing shows)
	NewShowTitle  string `json:"new_show_title" form:"new_show_title"` // New TV show title (if creating new)
	SeasonNumber  int    `json:"season_number" form:"season_number"`   // Season number
	EpisodeNumber int    `json:"episode_number" form:"episode_number"` // Episode number
}

// APIResponse is a standard JSON response wrapper
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// FoldersResponse represents the response for the /api/folders endpoint
type FoldersResponse struct {
	APIResponse
	Folders []FolderInfo `json:"folders"`
}

// TVShowsResponse represents the response for listing existing TV shows
type TVShowsResponse struct {
	APIResponse
	Shows []string `json:"shows"`
}
