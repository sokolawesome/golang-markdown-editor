package main

import (
	"log"
	"markdown-editor/internal/editor"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("markdown-editor: ")

	a := app.NewWithID("com.github.sokolawesome.markdown-editor")
	w := a.NewWindow("Markdown Editor")
	w.Resize(fyne.NewSize(1200, 800))

	editor.NewEditor(w)

	w.ShowAndRun()
}
