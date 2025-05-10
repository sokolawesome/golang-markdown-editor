package fileservice

import "fyne.io/fyne/v2"

type FileOperations interface {
	ReadFile(uri fyne.URI) ([]byte, error)
	WriteFile(uri fyne.URI, content []byte) error
	DeleteFile(uri fyne.URI) error
	RenameFile(oldURI, newURI fyne.URI) error
	FileExists(uri fyne.URI) (bool, error)
	ListDirectory(dir fyne.ListableURI) ([]fyne.URI, error)
	CreateDirectoryAll(path string) error
	GenerateUniqueFilename(dir fyne.ListableURI, basePrefix, extension string) (string, error)
	SanitizeFilenameComponent(input string) string
}
