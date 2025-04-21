package main

import (
	"markdown-editor/editor"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

func main() {
	a := app.NewWithID("markdown-editor")
	w := a.NewWindow("Markdown Editor")
	w.SetMaster()
	w.Resize(fyne.NewSize(1200, 800))

	editor := editor.NewEditor(w)

	w.SetContent(container.NewHSplit(
		editor.FileTree(),
		container.NewVSplit(
			editor.EditArea(),
			editor.PreviewArea(),
		),
	))

	w.ShowAndRun()
}
