package editor

import (
	"markdown-editor/internal/config"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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
}

func NewEditor(w fyne.Window) *Editor {
	e := &Editor{
		editComponent:    NewEditComponent(),
		previewComponent: NewPreviewComponent(),
		window:           w,
	}

	w.SetContent(container.NewVBox(
		widget.NewLabel("Initializing..."),
		widget.NewProgressBarInfinite(),
	))

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
			Content: err.Error(),
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
		container.NewVSplit(
			e.editComponent.View(),
			e.previewComponent.View(),
		),
	))
	e.refreshFileList()
}

func loadConfigWithRetry(w fyne.Window, maxAttempts int) (*config.Config, error) {
	var lastErr error
	for range maxAttempts {
		cfg, err := config.LoadConfig(w)
		if err == nil {
			return cfg, nil
		}
		lastErr = err
		time.Sleep(time.Second)
	}
	return nil, lastErr
}
