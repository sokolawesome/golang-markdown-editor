package editor

import (
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
			container := item.(*fyne.Container)
			label := container.Objects[0].(*widget.Label)
			btn := container.Objects[1].(*widget.Button)

			label.SetText(e.files[id].Name())

			btn.OnTapped = func() {
				e.deleteFile(e.files[id], id)
			}
		},
	)

	e.fileList.OnSelected = func(id widget.ListItemID) {
		e.loadFile(e.files[id])
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
