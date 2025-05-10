package editor

import (
	"errors"
	"fmt"
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

var (
	ErrEditorLoadConfig         = errors.New("failed to load configuration")
	ErrEditorInvalidConfigDir   = errors.New("invalid configured folder")
	ErrEditorFileLoad           = errors.New("failed to load file content")
	ErrEditorNoWorkspace        = errors.New("no workspace directory selected")
	ErrEditorGenerateFilename   = errors.New("failed to generate new filename")
	ErrEditorCreateFileURI      = errors.New("failed to create file URI")
	ErrEditorOpenFileWrite      = errors.New("failed to open file for writing")
	ErrEditorCloseFileWrite     = errors.New("failed to close file writer")
	ErrEditorWriteContent       = errors.New("failed to write content to file")
	ErrEditorFileDelete         = errors.New("failed to delete file")
	ErrEditorNoFileToSave       = errors.New("no file selected to save")
	ErrEditorRenameFile         = errors.New("failed to rename file")
	ErrEditorCheckFileExistence = errors.New("could not check for existing file during save")
	ErrEditorListDirectory      = errors.New("failed to list directory contents")
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
		wrappedErr := fmt.Errorf("%w: during editor initialization: %v", ErrEditorLoadConfig, err)
		app.ShowErrorNotification("Initialization Error", "Failed to load application configuration.", wrappedErr)
		e.window.SetContent(widget.NewLabel("Fatal Error: Could not load configuration. Please check logs."))
		return
	}

	dirURI := storage.NewFileURI(cfg.DefaultFolder)
	currentDir, err := storage.ListerForURI(dirURI)
	if err != nil {
		userMsg := fmt.Sprintf("The configured notes folder '%s' is invalid or inaccessible.", cfg.DefaultFolder)
		wrappedErr := fmt.Errorf("%w: '%s': %v", ErrEditorInvalidConfigDir, cfg.DefaultFolder, err)
		app.ShowErrorNotification("Initialization Error", userMsg, wrappedErr)
		e.window.SetContent(widget.NewLabel(fmt.Sprintf("Fatal Error: %s", userMsg)))
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
	app.ShowInfoNotification("Editor Ready", "Workspace initialized successfully.")
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
		log.Printf("Config load attempt %d failed: %+v", i+1, err)
		time.Sleep(time.Second)
	}
	return nil, fmt.Errorf("exceeded max attempts to load config: %w", lastErr)
}

func (e *Editor) loadFile(uri fyne.URI) {
	content, err := os.ReadFile(uri.Path())
	if err != nil {
		userMsg := fmt.Sprintf("Could not read content from '%s'.", uri.Name())
		wrappedErr := fmt.Errorf("%w: reading '%s': %v", ErrEditorFileLoad, uri.Path(), err)
		app.ShowErrorNotification("Error Loading File", userMsg, wrappedErr)
		return
	}

	e.currentFile = uri
	e.editComponent.SetContent(string(content))
	e.previewComponent.Update(string(content))
	log.Printf("File loaded: %s", uri.Path())
}

func (e *Editor) newFile() {
	if e.currentDir == nil {
		app.ShowErrorNotification("Error Creating File", ErrEditorNoWorkspace.Error(), ErrEditorNoWorkspace)
		return
	}

	baseName, err := generateNewFilename(e.currentDir)
	if err != nil {
		wrappedErr := fmt.Errorf("%w: %v", ErrEditorGenerateFilename, err)
		app.ShowErrorNotification("Error Creating File", "Could not generate a unique name for the new file.", wrappedErr)
		return
	}
	newURI, err := storage.Child(e.currentDir, baseName)
	if err != nil {
		wrappedErr := fmt.Errorf("%w: for '%s' in '%s': %v", ErrEditorCreateFileURI, baseName, e.currentDir.Path(), err)
		app.ShowErrorNotification("Error Creating File", "Could not prepare the new file location.", wrappedErr)
		return
	}

	header := "# " + strings.TrimSuffix(baseName, ".md") + "\n"
	writer, err := storage.Writer(newURI)
	if err != nil {
		wrappedErr := fmt.Errorf("%w: for '%s': %v", ErrEditorOpenFileWrite, newURI.Path(), err)
		app.ShowErrorNotification("Error Creating File", "Could not open the new file for writing.", wrappedErr)
		return
	}
	defer func() {
		if closeErr := writer.Close(); closeErr != nil {
			wrappedErr := fmt.Errorf("%w: for '%s': %v", ErrEditorCloseFileWrite, newURI.Path(), closeErr)
			app.ShowErrorNotification("Error Creating File", "Problem closing the new file after writing.", wrappedErr)
		}
	}()

	if _, err := writer.Write([]byte(header)); err != nil {
		wrappedErr := fmt.Errorf("%w: to '%s': %v", ErrEditorWriteContent, newURI.Path(), err)
		app.ShowErrorNotification("Error Creating File", "Failed to write initial content to the new file.", wrappedErr)
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
	app.ShowSuccessNotification("File Created", fmt.Sprintf("New file '%s' created.", newURI.Name()))
}

func (e *Editor) deleteFile(file fyne.URI, index int) {
	dialog.ShowConfirm("Delete File", fmt.Sprintf("Are you sure you want to delete '%s'?", file.Name()),
		func(ok bool) {
			if ok {
				if err := os.Remove(file.Path()); err != nil {
					userMsg := fmt.Sprintf("Could not delete file '%s'.", file.Name())
					wrappedErr := fmt.Errorf("%w: '%s': %v", ErrEditorFileDelete, file.Path(), err)
					app.ShowErrorNotification("Error Deleting File", userMsg, wrappedErr)
					return
				}

				app.ShowInfoNotification("File Deleted", fmt.Sprintf("File '%s' deleted.", file.Name()))

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
		app.ShowErrorNotification("Error Saving File", ErrEditorNoFileToSave.Error(), ErrEditorNoFileToSave)
		return
	}

	content := e.editComponent.Content()
	originalFilename := filepath.Base(e.currentFile.Path())
	var potentialFilename string
	for line := range strings.SplitSeq(content, "\n") {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" {
			potentialFilename = trimmedLine
			break
		}
	}

	var newFilename string
	if potentialFilename != "" {
		sanitizedTitle := sanitizeFilename(potentialFilename)
		if sanitizedTitle != "" && sanitizedTitle != filenameDefault {
			newFilename = sanitizedTitle + ".md"
		}
	}

	if newFilename != "" && newFilename != originalFilename {
		newURI, err := storage.Child(e.currentDir, newFilename)
		if err != nil {
			userMsg := fmt.Sprintf("Could not determine path for new filename '%s'.", newFilename)
			wrappedErr := fmt.Errorf("%w: creating child URI for '%s': %v", ErrEditorCreateFileURI, newFilename, err)
			app.ShowErrorNotification("Error Saving File", userMsg, wrappedErr)
			return
		}

		_, statErr := storage.Reader(newURI)
		if statErr == nil {
			app.ShowInfoNotification("File Exists", fmt.Sprintf("A file named '%s' already exists. Overwriting content in existing file after rename.", newFilename))
		} else if !errors.Is(statErr, storage.ErrNotExists) {
			userMsg := fmt.Sprintf("Could not check if a file named '%s' already exists.", newFilename)
			wrappedErr := fmt.Errorf("%w: checking existence of '%s': %v", ErrEditorCheckFileExistence, newURI.Path(), statErr)
			app.ShowErrorNotification("Error Saving File", userMsg, wrappedErr)
			return
		}

		if err := storage.Move(e.currentFile, newURI); err != nil {
			userMsg := fmt.Sprintf("Failed to rename file from '%s' to '%s'.", originalFilename, newFilename)
			wrappedErr := fmt.Errorf("%w: from '%s' to '%s': %v", ErrEditorRenameFile, e.currentFile.Path(), newURI.Path(), err)
			app.ShowErrorNotification("Error Saving File", userMsg, wrappedErr)
			return
		}
		e.currentFile = newURI
		log.Printf("File renamed to: %s", newURI.Path())
		app.ShowInfoNotification("File Renamed", fmt.Sprintf("File renamed to '%s'.", newFilename))
	}

	writer, err := storage.Writer(e.currentFile)
	if err != nil {
		userMsg := fmt.Sprintf("Could not open '%s' for saving.", e.currentFile.Name())
		wrappedErr := fmt.Errorf("%w: for '%s': %v", ErrEditorOpenFileWrite, e.currentFile.Path(), err)
		app.ShowErrorNotification("Error Saving File", userMsg, wrappedErr)
		return
	}
	defer func() {
		if closeErr := writer.Close(); closeErr != nil {
			userMsg := fmt.Sprintf("Problem closing '%s' after saving.", e.currentFile.Name())
			wrappedErr := fmt.Errorf("%w: for '%s': %v", ErrEditorCloseFileWrite, e.currentFile.Path(), closeErr)
			app.ShowErrorNotification("Error Saving File", userMsg, wrappedErr)
		}
	}()

	if _, err := writer.Write([]byte(content)); err != nil {
		userMsg := fmt.Sprintf("Failed to write content to '%s'.", e.currentFile.Name())
		wrappedErr := fmt.Errorf("%w: to '%s': %v", ErrEditorWriteContent, e.currentFile.Path(), err)
		app.ShowErrorNotification("Error Saving File", userMsg, wrappedErr)
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
	app.ShowSuccessNotification("File Saved", fmt.Sprintf("File '%s' saved successfully!", e.currentFile.Name()))
}
