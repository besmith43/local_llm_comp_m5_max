package services

import (
	"os"
	"path/filepath"
	"strings"
)

var videoExtensions = map[string]bool{
	".mp4": true, ".mkv": true, ".avi": true,
	".mov": true, ".webm": true, ".m4v": true,
	".flv": true, ".wmv": true, ".mpg": true,
	".mpeg": true, ".3gp": true, ".vob": true,
}

type VideoFile struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Ext      string `json:"ext"`
}

type FolderInfo struct {
	Path        string      `json:"path"`
	DisplayName string      `json:"display_name"`
	Videos      []VideoFile `json:"videos"`
}

func ScanSourceDir() ([]FolderInfo, error) {
	sourcePath := getEnv("SOURCE")
	if sourcePath == "" {
		return nil, nil
	}

	foldersMap := make(map[string][]VideoFile)

	err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if isVideoFile(info.Name()) {
			parentDir := filepath.Dir(path)
			size := info.Size()
			ext := filepath.Ext(info.Name())

			video := VideoFile{
				Filename: info.Name(),
				Size:     size,
				Ext:      strings.TrimPrefix(ext, "."),
			}

			foldersMap[parentDir] = append(foldersMap[parentDir], video)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	folders := make([]FolderInfo, 0, len(foldersMap))
	for path, videos := range foldersMap {
		displayName := filepath.Base(path)
		folders = append(folders, FolderInfo{
			Path:        path,
			DisplayName: displayName,
			Videos:      videos,
		})
	}

	return folders, nil
}

func isVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return videoExtensions[ext]
}
