package editor

import (
	"errors"
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

var (
	ErrFilenameGenerationDirNil      = errors.New("directory cannot be nil for filename generation")
	ErrFilenameGenerationURI         = errors.New("failed to create URI for unique filename check")
	ErrFilenameGenerationExistsCheck = errors.New("failed to check existence of potential filename")
	ErrFilenameGenerationMaxAttempts = errors.New("failed to generate unique filename after multiple attempts")
)

func generateNewFilename(dir fyne.ListableURI) (string, error) {
	if dir == nil {
		return "", ErrFilenameGenerationDirNil
	}

	timestamp := time.Now().Format(filenameTimestampFormat)
	baseName := fmt.Sprintf("%s%s.md", filenameBasePrefix, timestamp)
	uniqueName := baseName
	counter := 1

	for {
		newURI, err := storage.Child(dir, uniqueName)
		if err != nil {
			return "", fmt.Errorf("%w: creating URI for '%s': %v", ErrFilenameGenerationURI, uniqueName, err)
		}

		_, err = storage.Reader(newURI)

		if err != nil {
			if errors.Is(err, storage.ErrNotExists) {
				return uniqueName, nil
			}
			return "", fmt.Errorf("%w: checking existence of '%s' in '%s': %v", ErrFilenameGenerationExistsCheck, uniqueName, dir.Path(), err)
		}

		uniqueName = fmt.Sprintf("%s%s-%d.md", filenameBasePrefix, timestamp, counter)
		counter++

		if counter > 1000 {
			return "", fmt.Errorf("%w: in directory '%s'", ErrFilenameGenerationMaxAttempts, dir.Path())
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
