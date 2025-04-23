package editor

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

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

	return container.NewBorder(
		container.NewHBox(
			widget.NewButton("New File", e.newFile),
			widget.NewButton("Delete", e.deleteFile),
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
