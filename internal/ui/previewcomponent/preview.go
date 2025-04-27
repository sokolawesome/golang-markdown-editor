package previewcomponent

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type PreviewComponent struct {
	preview   *widget.RichText
	container *container.Scroll
}

func NewPreviewComponent() *PreviewComponent {
	preview := widget.NewRichTextFromMarkdown("")
	preview.Wrapping = fyne.TextWrapWord
	return &PreviewComponent{
		preview:   preview,
		container: container.NewScroll(preview),
	}
}

func (pc *PreviewComponent) Update(text string) {
	pc.preview.ParseMarkdown(text)
	pc.preview.Refresh()
}

func (pc *PreviewComponent) View() fyne.CanvasObject {
	return pc.container
}
