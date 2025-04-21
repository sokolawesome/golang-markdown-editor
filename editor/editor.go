package editor

import (
	"os"
	"path/filepath"
	"strings"

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
	currentDir  fyne.ListableURI
}

func NewEditor(w fyne.Window) *Editor {
	e := &Editor{
		editArea:    widget.NewMultiLineEntry(),
		previewArea: widget.NewRichText(),
		window:      w,
	}

	e.editArea.OnChanged = func(text string) {
		e.updatePreview(text)
	}

	return e
}

func (e *Editor) EditArea() fyne.CanvasObject {
	return container.NewScroll(e.editArea)
}

func (e *Editor) PreviewArea() fyne.CanvasObject {
	return container.NewScroll(e.previewArea)
}

func (e *Editor) FileTree() fyne.CanvasObject {
	e.fileList = widget.NewList(
		func() int {
			return len(e.files)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(e.files[id].Name())
		},
	)

	e.fileList.OnSelected = func(id widget.ListItemID) {
		e.loadFile(e.files[id])
	}

	openBtn := widget.NewButton("Open Folder", e.openFolder)
	newBtn := widget.NewButton("New File", e.newFile)
	deleteBtn := widget.NewButton("Delete", e.deleteFile)
	saveBtn := widget.NewButton("Save", e.saveFile)

	return container.NewBorder(
		container.NewHBox(openBtn, newBtn, deleteBtn, saveBtn),
		nil, nil, nil,
		e.fileList,
	)
}

func (e *Editor) openFolder() {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil || uri == nil {
			return
		}
		e.currentDir = uri
		e.refreshFileList()
	}, e.window)
}

func (e *Editor) refreshFileList() {
	if e.currentDir == nil {
		return
	}

	e.files = []fyne.URI{}
	items, err := e.currentDir.List()
	if err != nil {
		dialog.ShowError(err, e.window)
		return
	}

	for _, f := range items {
		if strings.HasSuffix(strings.ToLower(f.Name()), ".md") {
			e.files = append(e.files, f)
		}
	}

	e.fileList.Refresh()
}

func (e *Editor) loadFile(uri fyne.URI) {
	e.currentFile = uri
	read, err := storage.Reader(uri)
	if err != nil {
		dialog.ShowError(err, e.window)
		return
	}
	defer read.Close()

	content, err := os.ReadFile(uri.Path())
	if err != nil {
		dialog.ShowError(err, e.window)
		return
	}

	e.editArea.SetText(string(content))
	e.updatePreview(string(content))
}

func (e *Editor) newFile() {
	if e.currentDir == nil {
		dialog.ShowInformation("No folder open", "Please open a folder first", e.window)
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
			dialog.ShowError(err, e.window)
			return
		}

		e.refreshFileList()
	}, e.window)
}

func (e *Editor) deleteFile() {
	if e.currentFile == nil {
		dialog.ShowInformation("No file selected", "Please select a file to delete", e.window)
		return
	}

	dialog.ShowConfirm("Delete File", "Are you sure you want to delete "+e.currentFile.Name()+"?",
		func(ok bool) {
			if ok {
				err := os.Remove(e.currentFile.Path())
				if err != nil {
					dialog.ShowError(err, e.window)
					return
				}

				e.refreshFileList()
				e.currentFile = nil
				e.editArea.SetText("")
				e.previewArea.ParseMarkdown("")
			}
		}, e.window)
}

func (e *Editor) saveFile() {
	if e.currentFile == nil {
		dialog.ShowInformation("No file selected", "Please select a file to save to", e.window)
		return
	}

	writer, err := storage.Writer(e.currentFile)
	if err != nil {
		dialog.ShowError(err, e.window)
		return
	}
	defer writer.Close()

	_, err = writer.Write([]byte(e.editArea.Text))
	if err != nil {
		dialog.ShowError(err, e.window)
	}
}

func (e *Editor) updatePreview(text string) {
	e.previewArea.ParseMarkdown(text)
}
