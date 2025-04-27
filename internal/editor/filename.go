package editor

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

const (
	filenameLeadingHeadingRegex = `^#+\s*`
	filenameInvalidCharsRegex   = `[<>:"/\\|?*]`
	filenameSpaceReplacement    = "-"
	filenameDefault             = "untitled"
	filenameMaxLength           = 50
	filenameTimestampFormat     = "20060102_150405"
	filenameBasePrefix          = "note-"
)

var (
	leadingHeadingRegex = regexp.MustCompile(filenameLeadingHeadingRegex)
	invalidCharsRegex   = regexp.MustCompile(filenameInvalidCharsRegex)
)

func generateNewFilename(dir fyne.ListableURI) (string, error) {
	if dir == nil {
		return "", fmt.Errorf("directory cannot be nil")
	}

	timestamp := time.Now().Format(filenameTimestampFormat)
	baseName := fmt.Sprintf("%s%s.md", filenameBasePrefix, timestamp)
	uniqueName := baseName
	counter := 1

	for {
		newURI, err := storage.Child(dir, uniqueName)
		if err != nil {
			return "", fmt.Errorf("failed to create URI for %s: %w", uniqueName, err)
		}

		_, err = storage.Reader(newURI)

		if err != nil {
			if err == storage.ErrNotExists {
				return uniqueName, nil
			}
			return "", fmt.Errorf("failed to check existence of %s in directory %s: %w", uniqueName, dir.Path(), err)
		}

		uniqueName = fmt.Sprintf("%s%s-%d.md", filenameBasePrefix, timestamp, counter)
		counter++

		if counter > 1000 {
			return "", fmt.Errorf("failed to generate unique filename after 1000 attempts in directory %s", dir.Path())
		}
	}
}

func sanitizeFilename(input string) string {
	cleaned := leadingHeadingRegex.ReplaceAllString(input, "")
	cleaned = invalidCharsRegex.ReplaceAllString(cleaned, "")
	cleaned = strings.ReplaceAll(cleaned, " ", filenameSpaceReplacement)
	cleaned = strings.TrimSpace(cleaned)

	if cleaned == "" {
		return filenameDefault
	}

	if len(cleaned) > filenameMaxLength {
		cleaned = cleaned[:filenameMaxLength]
	}

	return strings.ToLower(cleaned)
}
