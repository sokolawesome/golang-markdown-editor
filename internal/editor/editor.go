package editor

import (
	"log"
	"markdown-editor/internal/app"
	"markdown-editor/internal/config"
	"markdown-editor/internal/ui/editorcomponent"
	"markdown-editor/internal/ui/filetreecomponent"
	"markdown-editor/internal/ui/previewcomponent"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

type Editor struct {
	currentFile       fyne.URI
	editComponent     app.EditorComponent
	previewComponent  app.PreviewComponent
	filetreeComponent app.FiletreeComponent
	window            fyne.Window
	config            *config.Config
	currentDir        fyne.ListableURI
	editorMode        bool
}

func NewEditor(w fyne.Window) *Editor {
	e := &Editor{
		window:     w,
		editorMode: true,
	}

	e.editComponent = editorcomponent.NewEditComponent()
	e.previewComponent = previewcomponent.NewPreviewComponent()
	e.filetreeComponent = filetreecomponent.NewFiletreeComponent(
		e.loadFile,
		e.deleteFile,
		e.newFile,
		e.saveFile,
	)

	w.SetContent(container.NewVBox(
		widget.NewLabel("Initializing..."),
		widget.NewProgressBarInfinite(),
	))

	shortcut := &desktop.CustomShortcut{KeyName: fyne.KeyM, Modifier: fyne.KeyModifierControl}
	w.Canvas().AddShortcut(shortcut, func(_ fyne.Shortcut) {
		e.toggleMode()
	})

	go e.initialize()
	e.editComponent.SetOnChanged(e.updatePreview)
	return e
}

func (e *Editor) updatePreview(text string) {
	e.previewComponent.Update(text)
}

func (e *Editor) initialize() {
	cfg, err := loadConfigWithRetry(e.window, 3)
	if err != nil {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Error",
			Content: "Failed to load configuration: " + err.Error(),
		})
		return
	}

	dirURI := storage.NewFileURI(cfg.DefaultFolder)
	currentDir, err := storage.ListerForURI(dirURI)
	if err != nil {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Error",
			Content: "Invalid configured folder: " + err.Error(),
		})
		return
	}

	e.config = cfg
	e.currentDir = currentDir
	e.filetreeComponent.SetDirectory(e.currentDir)

	e.window.Canvas().SetContent(container.NewHSplit(
		e.filetreeComponent.View(),
		e.editComponent.View(),
	))
	e.filetreeComponent.Refresh()
}

func (e *Editor) toggleMode() {
	e.editorMode = !e.editorMode
	if e.editorMode {
		e.window.Canvas().SetContent(container.NewHSplit(
			e.filetreeComponent.View(),
			e.editComponent.View(),
		))
	} else {
		e.previewComponent.Update(e.editComponent.Content())
		e.window.Canvas().SetContent(container.NewHSplit(
			e.filetreeComponent.View(),
			e.previewComponent.View(),
		))
	}
}

func loadConfigWithRetry(w fyne.Window, maxAttempts int) (*config.Config, error) {
	var lastErr error
	for i := range maxAttempts {
		cfg, err := config.LoadConfig(w)
		if err == nil {
			return cfg, nil
		}
		lastErr = err
		log.Printf("Config load attempt %d failed: %v", i+1, err)
		time.Sleep(time.Second)
	}
	return nil, lastErr
}

func (e *Editor) loadFile(uri fyne.URI) {
	content, err := os.ReadFile(uri.Path())
	if err != nil {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Error",
			Content: "Failed to load file: " + err.Error(),
		})
		log.Printf("Failed to load file %s: %v", uri.Path(), err)
		return
	}

	e.currentFile = uri
	e.editComponent.SetContent(string(content))
	e.previewComponent.Update(string(content))
}

func (e *Editor) newFile() {
	if e.currentDir == nil {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Error",
			Content: "No workspace directory selected",
		})
		return
	}

	baseName, err := generateNewFilename(e.currentDir)
	if err != nil {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Error",
			Content: "Failed to generate new filename: " + err.Error(),
		})
		return
	}
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
	e.filetreeComponent.Refresh()

	files := e.filetreeComponent.GetFiles()
	for i, file := range files {
		if file.String() == newURI.String() {
			e.filetreeComponent.SelectFile(i)
			break
		}
	}
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
					log.Printf("Failed to delete file %s: %v", file.Path(), err)
					return
				}

				if e.currentFile != nil && e.currentFile.String() == file.String() {
					e.currentFile = nil
					e.editComponent.SetContent("")
					e.previewComponent.Update("")
				}

				e.filetreeComponent.Refresh()

				files := e.filetreeComponent.GetFiles()
				if len(files) > 0 {
					newIndex := index
					if newIndex >= len(files) {
						newIndex = len(files) - 1
					}
					if newIndex < 0 {
						newIndex = 0
					}
					e.filetreeComponent.SelectFile(newIndex)
					e.loadFile(files[newIndex])
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

	if cleanTitle != "" && newFilename != filepath.Base(e.currentFile.Path()) {
		newURI, err := storage.Child(e.currentDir, newFilename)
		if err != nil {
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Error",
				Content: "Failed to create new file URI: " + err.Error(),
			})
			return
		}

		if _, err := storage.Reader(newURI); err == nil {
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Warning",
				Content: "File with name " + newFilename + " already exists. Overwriting.",
			})
		} else if !os.IsNotExist(err) {
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Error",
				Content: "Could not check for existing file: " + err.Error(),
			})
			return
		}

		if err := storage.Move(e.currentFile, newURI); err != nil {
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Error",
				Content: "Failed to rename file: " + err.Error(),
			})
			log.Printf("Failed to rename file from %s to %s: %v", e.currentFile.Path(), newURI.Path(), err)
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
		log.Printf("Failed to open writer for file %s: %v", e.currentFile.Path(), err)
		return
	}
	defer func() {
		if err := writer.Close(); err != nil {
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Error",
				Content: "Failed to close file writer: " + err.Error(),
			})
			log.Printf("Failed to close writer for file %s: %v", e.currentFile.Path(), err)
		}
	}()

	if _, err := writer.Write([]byte(content)); err != nil {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Error",
			Content: "Failed to write file: " + err.Error(),
		})
		log.Printf("Failed to write content to file %s: %v", e.currentFile.Path(), err)
		return
	}

	e.filetreeComponent.Refresh()

	if e.currentFile != nil {
		files := e.filetreeComponent.GetFiles()
		for i, file := range files {
			if file.String() == e.currentFile.String() {
				e.filetreeComponent.SelectFile(i)
				break
			}
		}
	}
}
