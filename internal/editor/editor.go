package editor

import (
	"log"
	"markdown-editor/internal/config"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

type Editor struct {
	currentFile      fyne.URI
	editComponent    *EditComponent
	previewComponent *PreviewComponent
	fileList         *widget.List
	window           fyne.Window
	files            []fyne.URI
	config           *config.Config
	currentDir       fyne.ListableURI
	editorMode       bool
}

func NewEditor(w fyne.Window) *Editor {
	e := &Editor{
		editComponent:    NewEditComponent(),
		previewComponent: NewPreviewComponent(),
		window:           w,
		editorMode:       true,
	}

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
	e.window.Canvas().SetContent(container.NewHSplit(
		e.FileTree(),
		e.editComponent.View(),
	))
	e.refreshFileList()
}

func (e *Editor) toggleMode() {
	e.editorMode = !e.editorMode
	if e.editorMode {
		e.window.Canvas().SetContent(container.NewHSplit(
			e.FileTree(),
			e.editComponent.View(),
		))
	} else {
		e.window.Canvas().SetContent(container.NewHSplit(
			e.FileTree(),
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
