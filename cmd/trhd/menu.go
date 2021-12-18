package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type PanelMenu struct {
	TRHD  *TRHD
	Panel *fyne.Container
}

func NewPanelMenu() *PanelMenu {
	pm := PanelMenu{}
	return &pm
}

func (pm *PanelMenu) Init(trhd *TRHD) {
	pm.TRHD = trhd
	fmt.Printf("getMainMenu\n")
	butKeys := widget.NewButton("Keystore", nil)
	butStore := widget.NewButton("Store", nil)
	butRetrieve := widget.NewButton("Retrieve", nil)
	butTX := widget.NewButton("TXs", nil)
	pm.Panel = container.NewVBox(butKeys, butRetrieve, butStore, butTX)
}
