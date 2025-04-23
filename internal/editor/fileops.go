package editor

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
)

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

	e.editComponent.SetContent(string(content))
	e.previewComponent.Update(string(content))
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
		originalPath := uri.Path()
		if filepath.Ext(originalPath) != ".md" {
			modifiedPath := originalPath + ".md"
			modifiedUri, _ := storage.Child(e.currentDir, filepath.Base(modifiedPath))

			writer.Close()
			writer, _ = storage.Writer(modifiedUri)
			defer writer.Close()

			if _, err := os.Stat(originalPath); err == nil {
				os.Remove(originalPath)
			}
		}

		_, err = writer.Write([]byte("# New Document\n"))
		e.editComponent.SetContent("# New Document\n")
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
				e.editComponent.SetContent("")
				e.previewComponent.Update("")
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

	writer.Write([]byte(e.editComponent.Content()))
}
