package editor

import (
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
)

func (e *Editor) loadFile(uri fyne.URI) {
	content, err := os.ReadFile(uri.Path())
	if err != nil {
		return
	}

	e.currentFile = uri
	e.editComponent.SetContent(string(content))
	e.previewComponent.Update(string(content))
}

func (e *Editor) newFile() {
	if e.currentDir == nil {
		return
	}

	baseName := generateNewFilename()
	newURI, _ := storage.Child(e.currentDir, baseName)

	header := "# " + strings.TrimSuffix(baseName, ".md") + "\n"
	writer, err := storage.Writer(newURI)
	if err != nil {
		return
	}
	defer writer.Close()

	if _, err := writer.Write([]byte(header)); err != nil {
		return
	}

	e.currentFile = newURI
	e.editComponent.SetContent(header)
	e.previewComponent.Update(header)
	e.refreshFileList()
}

func (e *Editor) deleteFile() {
	if e.currentFile == nil {
		return
	}

	dialog.ShowConfirm("Delete File", "Are you sure you want to delete "+e.currentFile.Name()+"?",
		func(ok bool) {
			if ok {
				os.Remove(e.currentFile.Path())
				e.refreshFileList()
				e.currentFile = nil
				e.editComponent.SetContent("")
				e.previewComponent.Update("")
			}
		}, e.window)
}

func (e *Editor) saveFile() {
	if e.currentFile == nil {
		return
	}

	content := e.editComponent.Content()

	lines := strings.SplitN(content, "\n", 2)
	rawTitle := strings.TrimSpace(lines[0])
	cleanTitle := strings.TrimPrefix(rawTitle, "# ")
	newFilename := sanitizeFilename(cleanTitle) + ".md"

	if newFilename != filepath.Base(e.currentFile.Path()) && cleanTitle != "" {
		newURI, _ := storage.Child(e.currentDir, newFilename)

		if err := storage.Move(e.currentFile, newURI); err == nil {
			e.currentFile = newURI
		}
	}

	writer, err := storage.Writer(e.currentFile)
	if err != nil {
		return
	}
	defer writer.Close()

	if _, err := writer.Write([]byte(content)); err != nil {
		return
	}

	e.refreshFileList()
}
