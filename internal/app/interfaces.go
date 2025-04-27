package app

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type EditorComponent interface {
	View() fyne.CanvasObject
	Content() string
	SetContent(text string)
	SetOnChanged(fn func(string))
}

type PreviewComponent interface {
	View() fyne.CanvasObject
	Update(text string)
}

type FiletreeComponent interface {
	View() fyne.CanvasObject
	SetDirectory(dir fyne.ListableURI)
	Refresh()
	GetFiles() []fyne.URI
	SelectFile(id widget.ListItemID)
}
