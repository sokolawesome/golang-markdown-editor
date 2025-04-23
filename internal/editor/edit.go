package editor

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type EditComponent struct {
	entry     *widget.Entry
	container *container.Scroll
}

func NewEditComponent() *EditComponent {
	entry := widget.NewMultiLineEntry()
	entry.Wrapping = fyne.TextWrapWord
	return &EditComponent{
		entry:     entry,
		container: container.NewScroll(entry),
	}
}

func (ec *EditComponent) SetOnChanged(fn func(string)) {
	ec.entry.OnChanged = fn
}

func (ec *EditComponent) View() fyne.CanvasObject {
	return ec.container
}

func (ec *EditComponent) Content() string {
	return ec.entry.Text
}

func (ec *EditComponent) SetContent(text string) {
	ec.entry.SetText(text)
}
