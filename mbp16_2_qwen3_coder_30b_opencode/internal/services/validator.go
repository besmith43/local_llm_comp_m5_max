package services

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

func SanitizeFilename(filename string) string {
	filename = strings.TrimSpace(filename)

	reg := regexp.MustCompile(`[<>:"/\\|？*]+`)
	filename = reg.ReplaceAllString(filename, "")

	filename = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, filename)

	filename = strings.ReplaceAll(filename, "  ", " ")
	return filename
}

func ValidateMovieInput(title, year string) (string, string, bool) {
	if title == "" {
		return "", "", false
	}
	title = SanitizeFilename(title)
	if len(title) == 0 {
		return "", "", false
	}
	if len(title) > 200 {
		return "", "", false
	}

	yearParsed, err := strconv.ParseInt(year, 10, 64)
	if err != nil {
		return "", "", false
	}
	if yearParsed < 1888 || yearParsed > 2100 {
		return "", "", false
	}

	return title, year, true
}

func ValidateTVShowInput(showTitle string, isNewShow bool, newShowTitle string, season, episode int) (string, int, int, bool) {
	if showTitle == "" && (!isNewShow || newShowTitle == "") {
		return "", 0, 0, false
	}

	var title string
	if isNewShow {
		if newShowTitle == "" {
			return "", 0, 0, false
		}
		title = SanitizeFilename(newShowTitle)
	} else {
		title = SanitizeFilename(showTitle)
	}

	if len(title) == 0 {
		return "", 0, 0, false
	}
	if len(title) > 200 {
		return "", 0, 0, false
	}

	if season < 1 || season > 999 {
		return "", 0, 0, false
	}
	if episode < 1 || episode > 999 {
		return "", 0, 0, false
	}

	return title, season, episode, true
}
