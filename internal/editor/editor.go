package editor

import (
	"errors"
	"fmt"
	"log"
	"markdown-editor/internal/app"
	"markdown-editor/internal/config"
	"markdown-editor/internal/fileservice"
	"markdown-editor/internal/ui/editorcomponent"
	"markdown-editor/internal/ui/filetreecomponent"
	"markdown-editor/internal/ui/previewcomponent"

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
	ErrEditorNoWorkspace        = errors.New("no workspace directory selected")
	ErrEditorCreateFileURI      = errors.New("failed to create file URI")
	ErrEditorNoFileToSave       = errors.New("no file selected to save")
	ErrEditorCheckFileExistence = errors.New("could not check for existing file during save")
	ErrEditorListDirectory      = errors.New("failed to list directory contents")
	ErrEditorFilenameInvalid    = errors.New("generated filename is invalid or empty")
)

const newFileBasePrefix = "note-"
const newFileExtension = ".md"
const editorFilenameDefault = "untitled"

type Editor struct {
	currentFile       fyne.URI
	editComponent     app.EditorComponent
	previewComponent  app.PreviewComponent
	filetreeComponent app.FiletreeComponent
	window            fyne.Window
	config            *config.Config
	currentDir        fyne.ListableURI
	editorMode        bool
	fs                *fileservice.Service
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
	content, err := e.fs.ReadFile(uri)
	if err != nil {
		userMsg := fmt.Sprintf("Could not read content from '%s'.", uri.Name())
		app.ShowErrorNotification("Error Loading File", userMsg, fmt.Errorf("loading file for editor: %w", err))
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

	baseName, err := e.fs.GenerateUniqueFilename(e.currentDir, newFileBasePrefix, newFileExtension)
	if err != nil {
		app.ShowErrorNotification("Error Creating File", "Could not generate a unique name for the new file.", fmt.Errorf("generating filename for new file: %w", err))
		return
	}

	newURI, err := storage.Child(e.currentDir, baseName)
	if err != nil {
		wrappedErr := fmt.Errorf("%w: for '%s' in '%s': %v", ErrEditorCreateFileURI, baseName, e.currentDir.Path(), err)
		app.ShowErrorNotification("Error Creating File", "Could not prepare the new file location.", wrappedErr)
		return
	}

	fileTitle := strings.TrimSuffix(baseName, newFileExtension)
	fileTitle = strings.TrimPrefix(fileTitle, newFileBasePrefix)
	header := "# " + fileTitle + "\n"

	if err := e.fs.WriteFile(newURI, []byte(header)); err != nil {
		app.ShowErrorNotification("Error Creating File", "Failed to write initial content to the new file.", fmt.Errorf("writing new file content: %w", err))
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

func (e *Editor) deleteFile(fileToDelete fyne.URI, index int) {
	dialog.ShowConfirm("Delete File", fmt.Sprintf("Are you sure you want to delete '%s'?", fileToDelete.Name()),
		func(ok bool) {
			if ok {
				if err := e.fs.DeleteFile(fileToDelete); err != nil {
					userMsg := fmt.Sprintf("Could not delete file '%s'.", fileToDelete.Name())
					app.ShowErrorNotification("Error Deleting File", userMsg, fmt.Errorf("deleting file for editor: %w", err))
					return
				}

				app.ShowInfoNotification("File Deleted", fmt.Sprintf("File '%s' deleted.", fileToDelete.Name()))

				if e.currentFile != nil && e.currentFile.String() == fileToDelete.String() {
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

	var potentialFilenameTitle string
	for line := range strings.SplitSeq(content, "\n") {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" {
			potentialFilenameTitle = trimmedLine
			break
		}
	}

	var newFilenameComponent string
	if potentialFilenameTitle != "" {
		sanitizedTitle := e.fs.SanitizeFilenameComponent(potentialFilenameTitle)
		if sanitizedTitle != "" {
			newFilenameComponent = sanitizedTitle
		}
	}
	if newFilenameComponent == "" {
		newFilenameComponent = editorFilenameDefault
	}
	desiredNewFilename := newFilenameComponent + ".md"

	if desiredNewFilename != originalFilename {
		newURI, err := storage.Child(e.currentDir, desiredNewFilename)
		if err != nil {
			userMsg := fmt.Sprintf("Could not determine path for new filename '%s'.", desiredNewFilename)
			wrappedErr := fmt.Errorf("%w: creating child URI for '%s': %v", ErrEditorCreateFileURI, desiredNewFilename, err)
			app.ShowErrorNotification("Error Saving File", userMsg, wrappedErr)
			return
		}

		exists, err := e.fs.FileExists(newURI)
		if err != nil {
			userMsg := fmt.Sprintf("Could not check if a file named '%s' already exists.", desiredNewFilename)
			app.ShowErrorNotification("Error Saving File", userMsg, fmt.Errorf("checking file existence during save: %w", err))
			return
		}
		if exists {
			app.ShowInfoNotification("File Exists", fmt.Sprintf("A file named '%s' already exists. Overwriting content after rename.", desiredNewFilename))
		}

		if err := e.fs.RenameFile(e.currentFile, newURI); err != nil {
			userMsg := fmt.Sprintf("Failed to rename file from '%s' to '%s'.", originalFilename, desiredNewFilename)
			app.ShowErrorNotification("Error Saving File", userMsg, fmt.Errorf("renaming file for save: %w", err))
			return
		}
		e.currentFile = newURI
		log.Printf("File renamed to: %s", newURI.Path())
		app.ShowInfoNotification("File Renamed", fmt.Sprintf("File renamed to '%s'.", desiredNewFilename))
	}

	if err := e.fs.WriteFile(e.currentFile, []byte(content)); err != nil {
		userMsg := fmt.Sprintf("Failed to write content to '%s'.", e.currentFile.Name())
		app.ShowErrorNotification("Error Saving File", userMsg, fmt.Errorf("writing file content for save: %w", err))
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
