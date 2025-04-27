package filetreecomponent

import (
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type FileTreeComponent struct {
	fileList   *widget.List
	files      []fyne.URI
	currentDir fyne.ListableURI

	OnSelectFile func(fyne.URI)
	OnDeleteFile func(fyne.URI, int)
	OnNewFile    func()
	OnSaveFile   func()

	widget fyne.CanvasObject
}

func NewFileTreeComponent(onSelect func(fyne.URI), onDelete func(fyne.URI, int), onNew func(), onSave func()) *FileTreeComponent {
	ftc := &FileTreeComponent{
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
				log.Printf("Failed to cast item to container for file %s", ftc.files[id].Name())
				return
			}
			label, ok := container.Objects[0].(*widget.Label)
			if !ok {
				log.Printf("Failed to cast object to label for file %s", ftc.files[id].Name())
				return
			}
			btn, ok := container.Objects[1].(*widget.Button)
			if !ok {
				log.Printf("Failed to cast object to button for file %s", ftc.files[id].Name())
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
		if ftc.OnSelectFile != nil {
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

func (ftc *FileTreeComponent) SetDirectory(dir fyne.ListableURI) {
	ftc.currentDir = dir
}

func (ftc *FileTreeComponent) Refresh() {
	if ftc.currentDir == nil {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Error",
			Content: "No workspace directory selected",
		})
		return
	}

	ftc.files = []fyne.URI{}
	items, err := ftc.currentDir.List()
	if err != nil {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Error",
			Content: "Failed to list directory: " + err.Error(),
		})
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

func (ftc *FileTreeComponent) View() fyne.CanvasObject {
	return ftc.widget
}

func (ftc *FileTreeComponent) GetFiles() []fyne.URI {
	return ftc.files
}

func (ftc *FileTreeComponent) SelectFile(id widget.ListItemID) {
	ftc.fileList.Select(id)
}
