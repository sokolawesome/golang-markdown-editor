package filetreecomponent

import (
	"fmt"
	"log"
	"markdown-editor/internal/app"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type FiletreeComponent struct {
	fileList   *widget.List
	files      []fyne.URI
	currentDir fyne.ListableURI

	OnSelectFile func(fyne.URI)
	OnDeleteFile func(fyne.URI, int)
	OnNewFile    func()
	OnSaveFile   func()

	widget fyne.CanvasObject
}

func NewFiletreeComponent(onSelect func(fyne.URI), onDelete func(fyne.URI, int), onNew func(), onSave func()) *FiletreeComponent {
	ftc := &FiletreeComponent{
		OnSelectFile: onSelect,
		OnDeleteFile: onDelete,
		OnNewFile:    onNew,
		OnSaveFile:   onSave,
	}

	ftc.fileList = widget.NewList(
		func() int { return len(ftc.files) },
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
				log.Printf("Error: Failed to cast item to container for file ID %d", id)
				return
			}
			label, ok := container.Objects[0].(*widget.Label)
			if !ok {
				log.Printf("Error: Failed to cast object to label for file ID %d", id)
				return
			}
			btn, ok := container.Objects[1].(*widget.Button)
			if !ok {
				log.Printf("Error: Failed to cast object to button for file ID %d", id)
				return
			}

			label.SetText(ftc.files[id].Name())

			fileURI := ftc.files[id]
			fileIndex := id
			btn.OnTapped = func() {
				if ftc.OnDeleteFile != nil {
					ftc.OnDeleteFile(fileURI, fileIndex)
				}
			}
		},
	)

	ftc.fileList.OnSelected = func(id widget.ListItemID) {
		if ftc.OnSelectFile != nil && id < len(ftc.files) && id >= 0 {
			ftc.OnSelectFile(ftc.files[id])
		}
	}

	ftc.widget = container.NewBorder(
		container.NewHBox(
			widget.NewButton("New File", func() {
				if ftc.OnNewFile != nil {
					ftc.OnNewFile()
				}
			}),
			widget.NewButton("Save", func() {
				if ftc.OnSaveFile != nil {
					ftc.OnSaveFile()
				}
			}),
		),
		nil, nil, nil,
		ftc.fileList,
	)

	return ftc
}

func (ftc *FiletreeComponent) SetDirectory(dir fyne.ListableURI) {
	ftc.currentDir = dir
}

func (ftc *FiletreeComponent) Refresh() {
	if ftc.currentDir == nil {
		app.ShowErrorNotification("File Tree Error", "Workspace directory not set.", nil)
		return
	}

	ftc.files = []fyne.URI{}
	items, err := ftc.currentDir.List()
	if err != nil {
		userMsg := fmt.Sprintf("Could not list files in directory '%s'.", ftc.currentDir.Path())
		app.ShowErrorNotification("File Tree Error", userMsg, err)
		return
	}

	for _, f := range items {
		if strings.HasSuffix(strings.ToLower(f.Name()), ".md") {
			ftc.files = append(ftc.files, f)
		}
	}

	if ftc.fileList != nil {
		ftc.fileList.Refresh()
	}
}

func (ftc *FiletreeComponent) View() fyne.CanvasObject {
	return ftc.widget
}

func (ftc *FiletreeComponent) GetFiles() []fyne.URI {
	return ftc.files
}

func (ftc *FiletreeComponent) SelectFile(id widget.ListItemID) {
	if id < len(ftc.files) && id >= 0 {
		ftc.fileList.Select(id)
	} else {
		log.Printf("Warning: Attempted to select invalid file ID %d (total files: %d)", id, len(ftc.files))
	}
}
