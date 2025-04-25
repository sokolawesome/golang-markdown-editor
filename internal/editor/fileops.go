package editor

import (
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
)

func (e *Editor) loadFile(uri fyne.URI) error {
	content, err := os.ReadFile(uri.Path())
	if err != nil {
		return err
	}

	e.currentFile = uri
	e.editComponent.SetContent(string(content))
	e.previewComponent.Update(string(content))
	return nil
}

func (e *Editor) newFile() {
	if e.currentDir == nil {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Error",
			Content: "No workspace directory selected",
		})
		return
	}

	baseName := generateNewFilename()
	newURI, err := storage.Child(e.currentDir, baseName)
	if err != nil {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Error",
			Content: "Failed to create file URI: " + err.Error(),
		})
		return
	}

	header := "# " + strings.TrimSuffix(baseName, ".md") + "\n"
	writer, err := storage.Writer(newURI)
	if err != nil {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Error",
			Content: "Failed to open file writer: " + err.Error(),
		})
		return
	}
	defer func() {
		if err := writer.Close(); err != nil {
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Error",
				Content: "Failed to close file writer: " + err.Error(),
			})
		}
	}()

	if _, err := writer.Write([]byte(header)); err != nil {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Error",
			Content: "Failed to write file: " + err.Error(),
		})
		return
	}

	e.currentFile = newURI
	e.editComponent.SetContent(header)
	e.previewComponent.Update(header)
	e.editorMode = false
	e.toggleMode()
	e.refreshFileList()
}

func (e *Editor) deleteFile(file fyne.URI, index int) {
	dialog.ShowConfirm("Delete File", "Are you sure you want to delete "+file.Name()+"?",
		func(ok bool) {
			if ok {
				if err := os.Remove(file.Path()); err != nil {
					fyne.CurrentApp().SendNotification(&fyne.Notification{
						Title:   "Error",
						Content: "Failed to delete file: " + err.Error(),
					})
					return
				}
				e.refreshFileList()

				if len(e.files) > 0 {
					newIndex := index
					if newIndex >= len(e.files) {
						newIndex = len(e.files) - 1
					}

					if newIndex < 0 {
						newIndex = 0
					}

					e.fileList.Select(newIndex)
					if err := e.loadFile(e.files[newIndex]); err != nil {
						fyne.CurrentApp().SendNotification(&fyne.Notification{
							Title:   "Error",
							Content: "Failed to load file: " + err.Error(),
						})
					}
				} else {
					e.currentFile = nil
					e.editComponent.SetContent("")
					e.previewComponent.Update("")
				}
			}
		}, e.window)
}

func (e *Editor) saveFile() {
	if e.currentFile == nil {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Error",
			Content: "No file selected to save",
		})
		return
	}

	content := e.editComponent.Content()

	lines := strings.SplitN(content, "\n", 2)
	rawTitle := strings.TrimSpace(lines[0])
	cleanTitle := strings.TrimPrefix(rawTitle, "# ")
	newFilename := sanitizeFilename(cleanTitle) + ".md"

	if newFilename != filepath.Base(e.currentFile.Path()) && cleanTitle != "" {
		newURI, err := storage.Child(e.currentDir, newFilename)
		if err != nil {
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Error",
				Content: "Failed to create new file URI: " + err.Error(),
			})
			return
		}

		if err := storage.Move(e.currentFile, newURI); err != nil {
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Error",
				Content: "Failed to rename file: " + err.Error(),
			})
			return
		}
		e.currentFile = newURI
	}

	writer, err := storage.Writer(e.currentFile)
	if err != nil {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Error",
			Content: "Failed to open file writer: " + err.Error(),
		})
		return
	}
	defer func() {
		if err := writer.Close(); err != nil {
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Error",
				Content: "Failed to close file writer: " + err.Error(),
			})
		}
	}()

	if _, err := writer.Write([]byte(content)); err != nil {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Error",
			Content: "Failed to write file: " + err.Error(),
		})
		return
	}

	e.refreshFileList()
}
