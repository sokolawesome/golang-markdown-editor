package app

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
)

func ShowErrorNotification(title string, userMessage string, err error) {
	fullErrorMessage := userMessage
	if err != nil {
		if userMessage == "" {
			userMessage = err.Error()
		}
		fullErrorMessage = fmt.Sprintf("%s (Details: %+v)", userMessage, err)
	}

	log.Printf("ERROR: %s: %s", title, fullErrorMessage)

	fyne.CurrentApp().SendNotification(&fyne.Notification{
		Title:   title,
		Content: userMessage,
	})
}

func ShowInfoNotification(title string, message string) {
	log.Printf("INFO: %s: %s", title, message)
	fyne.CurrentApp().SendNotification(&fyne.Notification{
		Title:   title,
		Content: message,
	})
}

func ShowSuccessNotification(title string, message string) {
	ShowInfoNotification(title, message)
}
