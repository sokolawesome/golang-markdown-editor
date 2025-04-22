package editor

import (
	"markdown-editor/config"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

type Editor struct {
	currentFile fyne.URI
	editArea    *widget.Entry
	previewArea *widget.RichText
	fileList    *widget.List
	window      fyne.Window
	files       []fyne.URI
	config      *config.Config
	currentDir  fyne.ListableURI
}

func NewEditor(w fyne.Window) *Editor {
	e := &Editor{
		editArea:    widget.NewMultiLineEntry(),
		previewArea: widget.NewRichText(),
		window:      w,
	}

	w.SetContent(container.NewVBox(
		widget.NewLabel("Initializing..."),
		widget.NewProgressBarInfinite(),
	))

	go e.initialize()

	e.editArea.OnChanged = e.updatePreview
	return e
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

	e.window.Canvas().SetContent(container.NewHSplit(
		e.FileTree(),
		container.NewVSplit(
			e.EditArea(),
			e.PreviewArea(),
		),
	))

	e.config = cfg
	e.currentDir = currentDir
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

func (e *Editor) EditArea() fyne.CanvasObject {
	return container.NewScroll(e.editArea)
}

func (e *Editor) PreviewArea() fyne.CanvasObject {
	return container.NewScroll(e.previewArea)
}

func (e *Editor) FileTree() fyne.CanvasObject {
	e.fileList = widget.NewList(
		func() int { return len(e.files) },
		func() fyne.CanvasObject { return widget.NewLabel("template") },
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(e.files[id].Name())
		},
	)

	e.fileList.OnSelected = func(id widget.ListItemID) {
		e.loadFile(e.files[id])
	}

	newBtn := widget.NewButton("New File", e.newFile)
	deleteBtn := widget.NewButton("Delete", e.deleteFile)
	saveBtn := widget.NewButton("Save", e.saveFile)

	return container.NewBorder(
		container.NewHBox(newBtn, deleteBtn, saveBtn),
		nil, nil, nil,
		e.fileList,
	)
}

func (e *Editor) refreshFileList() {
	if e.currentDir == nil {
		return
	}

	e.files = []fyne.URI{}
	items, err := e.currentDir.List()
	if err != nil {
		return
	}

	for _, f := range items {
		if strings.HasSuffix(strings.ToLower(f.Name()), ".md") {
			e.files = append(e.files, f)
		}
	}

	if e.fileList != nil {
		e.fileList.Refresh()
	}
}

func (e *Editor) loadFile(uri fyne.URI) {
	e.currentFile = uri
	read, err := storage.Reader(uri)
	if err != nil {
		return
	}
	defer read.Close()

	content, err := os.ReadFile(uri.Path())
	if err != nil {
		return
	}

	e.editArea.SetText(string(content))
	e.updatePreview(string(content))
}

func (e *Editor) newFile() {
	if e.currentDir == nil {
		return
	}

	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil || writer == nil {
			return
		}
		defer writer.Close()

		uri := writer.URI()
		path := uri.Path()
		if filepath.Ext(path) != ".md" {
			path += ".md"
			uri, _ = storage.Child(e.currentDir, filepath.Base(path))
			writer, _ = storage.Writer(uri)
			defer writer.Close()
		}

		_, err = writer.Write([]byte("# New Document\n"))
		if err != nil {
			return
		}

		e.refreshFileList()
	}, e.window)
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
				e.editArea.SetText("")
				e.previewArea.ParseMarkdown("")
			}
		}, e.window)
}

func (e *Editor) saveFile() {
	if e.currentFile == nil {
		return
	}

	writer, err := storage.Writer(e.currentFile)
	if err != nil {
		return
	}
	defer writer.Close()

	writer.Write([]byte(e.editArea.Text))
}

func (e *Editor) updatePreview(text string) {
	e.previewArea.ParseMarkdown(text)
}
