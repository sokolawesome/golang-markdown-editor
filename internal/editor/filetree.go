package editor

import (
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (e *Editor) FileTree() fyne.CanvasObject {
	e.fileList = widget.NewList(
		func() int { return len(e.files) },
		func() fyne.CanvasObject {
			deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
			deleteBtn.Importance = widget.LowImportance
			return container.NewBorder(
				nil,
				nil,
				widget.NewLabel("template"),
				deleteBtn,
				nil)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			container, ok := item.(*fyne.Container)
			if !ok {
				log.Printf("Failed to cast item to container for file %s", e.files[id].Name())
				return
			}
			label, ok := container.Objects[0].(*widget.Label)
			if !ok {
				log.Printf("Failed to cast object to label for file %s", e.files[id].Name())
				return
			}
			btn, ok := container.Objects[1].(*widget.Button)
			if !ok {
				log.Printf("Failed to cast object to button for file %s", e.files[id].Name())
				return
			}

			label.SetText(e.files[id].Name())

			btn.OnTapped = func() {
				e.deleteFile(e.files[id], id)
			}
		},
	)

	e.fileList.OnSelected = func(id widget.ListItemID) {
		if err := e.loadFile(e.files[id]); err != nil {
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Error",
				Content: "Failed to load file: " + err.Error(),
			})
		}
	}

	return container.NewBorder(
		container.NewHBox(
			widget.NewButton("New File", e.newFile),
			widget.NewButton("Save", e.saveFile),
		),
		nil, nil, nil,
		e.fileList,
	)
}

func (e *Editor) refreshFileList() {
	if e.currentDir == nil {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Error",
			Content: "No workspace directory selected",
		})
		return
	}

	e.files = []fyne.URI{}
	items, err := e.currentDir.List()
	if err != nil {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Error",
			Content: "Failed to list directory: " + err.Error(),
		})
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
