package main

import (
	"fmt"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

var trhURL *url.URL

func main() {
	appTRH := app.New()
	trhURL, _ = url.Parse("https://ejfhp.com/projects/trh")

	mainWin := appTRH.NewWindow("TRH - The Rabbit Hole")
	mainWin.Resize(fyne.NewSize(600, 500))

	pm := NewPanelMenu()
	pk := NewPanelKeys()

	trhd := NewTRHD(mainWin, pm, pk)
	pm.Init(trhd)
	pk.Init(trhd)

	err := trhd.showSplash(appTRH)
	if err != nil {
		fmt.Printf("failed to show splash: %v\n", err)
	}
	fmt.Printf("run\n")
	appTRH.Run()
	fmt.Printf("run end\n")
}
