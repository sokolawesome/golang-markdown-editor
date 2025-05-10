package fileservice

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

var (
	ErrFileServiceReadFailed     = errors.New("fileservice: read failed")
	ErrFileServiceWriteFailed    = errors.New("fileservice: write failed")
	ErrFileServiceCloseFailed    = errors.New("fileservice: close failed")
	ErrFileServiceDeleteFailed   = errors.New("fileservice: delete failed")
	ErrFileServiceRenameFailed   = errors.New("fileservice: rename failed")
	ErrFileServiceExistenceCheck = errors.New("fileservice: existence check failed")
	ErrFileServiceListDirFailed  = errors.New("fileservice: list directory failed")
	ErrFileServiceMkdirAllFailed = errors.New("fileservice: mkdirall failed")
	ErrFilenameGenDirNil         = errors.New("fileservice: directory cannot be nil for filename generation")
	ErrFilenameGenURI            = errors.New("fileservice: failed to create URI for unique filename check")
	ErrFilenameGenExistsCheck    = errors.New("fileservice: failed to check existence of potential filename")
	ErrFilenameGenMaxAttempts    = errors.New("fileservice: failed to generate unique filename after multiple attempts")
)

const (
	filenameLeadingHeadingRegex = `^#+\s*`
	filenameInvalidCharsRegex   = `[<>:"/\\|?*]`
	filenameSpaceReplacement    = "-"
	filenameMaxLength           = 50
	filenameTimestampFormat     = "20060102_150405"
)

var (
	leadingHeadingRegexFS = regexp.MustCompile(filenameLeadingHeadingRegex)
	invalidCharsRegexFS   = regexp.MustCompile(filenameInvalidCharsRegex)
)

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) ReadFile(uri fyne.URI) ([]byte, error) {
	content, err := os.ReadFile(uri.Path())
	if err != nil {
		return nil, fmt.Errorf("%w: reading '%s': %v", ErrFileServiceReadFailed, uri.Path(), err)
	}
	return content, nil
}

func (s *Service) WriteFile(uri fyne.URI, content []byte) error {
	writer, err := storage.Writer(uri)
	if err != nil {
		return fmt.Errorf("%w: opening writer for '%s': %v", ErrFileServiceWriteFailed, uri.Path(), err)
	}
	defer func() {
		if closeErr := writer.Close(); closeErr != nil {
			fmt.Printf("fileservice: info: problem closing writer for '%s' after write: %v\n", uri.Path(), closeErr)
		}
	}()

	if _, err := writer.Write(content); err != nil {
		_ = writer.Close()
		return fmt.Errorf("%w: writing to '%s': %v", ErrFileServiceWriteFailed, uri.Path(), err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("%w: closing writer for '%s': %v", ErrFileServiceCloseFailed, uri.Path(), err)
	}
	return nil
}

func (s *Service) DeleteFile(uri fyne.URI) error {
	if err := os.Remove(uri.Path()); err != nil {
		return fmt.Errorf("%w: removing '%s': %v", ErrFileServiceDeleteFailed, uri.Path(), err)
	}
	return nil
}

func (s *Service) RenameFile(oldURI, newURI fyne.URI) error {
	if err := storage.Move(oldURI, newURI); err != nil {
		return fmt.Errorf("%w: moving '%s' to '%s': %v", ErrFileServiceRenameFailed, oldURI.Path(), newURI.Path(), err)
	}
	return nil
}

func (s *Service) FileExists(uri fyne.URI) (bool, error) {
	result, err := storage.Exists(uri)
	if err != nil {
		return false, fmt.Errorf("%w: checking '%s': %v", ErrFileServiceExistenceCheck, uri.Path(), err)
	}

	return result, nil
}

func (s *Service) ListDirectory(dir fyne.ListableURI) ([]fyne.URI, error) {
	items, err := dir.List()
	if err != nil {
		return nil, fmt.Errorf("%w: listing '%s': %v", ErrFileServiceListDirFailed, dir.Path(), err)
	}

	return items, nil
}

func (s *Service) CreateDirectoryAll(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("%w: creating directory '%s': %v", ErrFileServiceMkdirAllFailed, path, err)
	}
	return nil
}

func (s *Service) GenerateUniqueFilename(dir fyne.ListableURI, basePrefix, extension string) (string, error) {
	if dir == nil {
		return "", ErrFilenameGenDirNil
	}

	timestamp := time.Now().Format(filenameTimestampFormat)
	baseName := fmt.Sprintf("%s%s%s", basePrefix, timestamp, extension)
	uniqueName := baseName
	counter := 1

	for {
		newURI, err := storage.Child(dir, uniqueName)
		if err != nil {
			return "", fmt.Errorf("%w: creating URI for '%s': %v", ErrFilenameGenURI, uniqueName, err)
		}

		exists, err := s.FileExists(newURI)
		if err != nil {
			return "", fmt.Errorf("%w: checking existence of '%s' in '%s': %v", ErrFilenameGenExistsCheck, uniqueName, dir.Path(), err)
		}
		if !exists {
			return uniqueName, nil
		}

		uniqueName = fmt.Sprintf("%s%s-%d%s", basePrefix, timestamp, counter, extension)
		counter++

		if counter > 1000 {
			return "", fmt.Errorf("%w: in directory '%s' with prefix '%s'", ErrFilenameGenMaxAttempts, dir.Path(), basePrefix)
		}
	}
}

func (s *Service) SanitizeFilenameComponent(input string) string {
	cleaned := leadingHeadingRegexFS.ReplaceAllString(input, "")
	cleaned = invalidCharsRegexFS.ReplaceAllString(cleaned, "")
	cleaned = strings.ReplaceAll(cleaned, " ", filenameSpaceReplacement)
	cleaned = strings.TrimSpace(cleaned)

	if cleaned == "" {
		return ""
	}

	if len(cleaned) > filenameMaxLength {
		cleaned = cleaned[:filenameMaxLength]
	}
	return strings.ToLower(cleaned)
}
