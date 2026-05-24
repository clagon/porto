package main

import (
	"log"

	"github.com/clagon/port-mapper/backend/internal/app"
	"github.com/clagon/port-mapper/backend/internal/browseropener"
)

func main() {
	a, err := app.New(app.AppOptions{
		OpenBrowser:   true,
		BrowserOpener: browseropener.New(),
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}
