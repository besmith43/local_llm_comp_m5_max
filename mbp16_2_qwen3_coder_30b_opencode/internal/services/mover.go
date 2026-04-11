package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func MoveMovie(videoPath, title, year string) error {
	dest := getEnv("DESTINATION")
	title = SanitizeFilename(title)

	dirPath := filepath.Join(dest, "Movies", fmt.Sprintf("%s (%s)", title, year))
	newFilename := fmt.Sprintf("%s (%s).%s", title, year, filepath.Ext(videoPath))

	return moveFileToPath(videoPath, dirPath, newFilename)
}

func MoveTVShow(videoPath, showTitle, episodeTitle string, season, episode int) error {
	dest := getEnv("DESTINATION")
	showTitle = SanitizeFilename(showTitle)
	episodeTitle = SanitizeFilename(episodeTitle)

	dirPath := filepath.Join(dest, "TV Shows", showTitle, fmt.Sprintf("Season %d", season))
	newFilename := fmt.Sprintf("%s - S%02dE%02d - %s.%s",
		showTitle, season, episode, episodeTitle, filepath.Ext(videoPath))

	return moveFileToPath(videoPath, dirPath, newFilename)
}

func moveFileToPath(sourcePath, destDir, destFilename string) error {
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	destPath := filepath.Join(destDir, destFilename)

	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("file already exists, skipped")
	}

	if err := os.Rename(sourcePath, destPath); err != nil {
		if strings.Contains(err.Error(), "invalid cross-device link") {
			return copyFileThenRemove(sourcePath, destPath)
		}
		return fmt.Errorf("failed to move file: %w", err)
	}

	return nil
}

func copyFileThenRemove(source, dest string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create dest file: %w", err)
	}
	defer destFile.Close()

	if _, err := destFile.ReadFrom(sourceFile); err != nil {
		os.Remove(dest)
		return fmt.Errorf("failed to copy file: %w", err)
	}

	if err := os.Remove(source); err != nil {
		return fmt.Errorf("failed to remove source file: %w", err)
	}

	return nil
}

func GetTVShowFolders() ([]string, error) {
	dest := getEnv("DESTINATION")
	tvShowsPath := filepath.Join(dest, "TV Shows")

	if _, err := os.Stat(tvShowsPath); os.IsNotExist(err) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(tvShowsPath)
	if err != nil {
		return nil, err
	}

	shows := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			shows = append(shows, entry.Name())
		}
	}

	return shows, nil
}
