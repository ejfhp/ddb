package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/ejfhp/ddb/trh"
)

type TRHD struct {
	Keys       *PanelKeys
	Menu       *PanelMenu
	MainWindow fyne.Window
	TRH        *trh.TRH
}

func NewTRHD(mainwin fyne.Window, menu *PanelMenu, Keys *PanelKeys) *TRHD {
	return &TRHD{MainWindow: mainwin, Menu: menu, Keys: Keys, TRH: trh.NewWithoutKeystore()}
}

func (trhd *TRHD) setMainPanel(panel *fyne.Container) {
	fmt.Printf("setMainPanel\n")
	trhd.MainWindow.SetContent(container.New(layout.NewHBoxLayout(), trhd.Menu.Panel, layout.NewSpacer(), panel))
	trhd.MainWindow.Show()
	fmt.Printf("setMainPanel done\n")
}

func (trhd *TRHD) showSplash(app fyne.App) error {
	drv := app.Driver()
	if drv, ok := drv.(desktop.Driver); ok {
		splashWin := drv.CreateSplashWindow()
		splashImage, err := resources.Open(imgLogo)
		if err != nil {
			return fmt.Errorf("error opening splash image: %w", err)
		}
		logo := canvas.NewImageFromReader(splashImage, "logo")
		logo.FillMode = canvas.ImageFillContain
		logo.SetMinSize(fyne.NewSize(300, 300))

		welcomePanel := container.NewCenter(container.NewVBox(
			widget.NewLabelWithStyle("The Rabbit Hole", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			logo,
			// container.NewCenter(
			// 	widget.NewHyperlink("Read the TRH docs.", trhURL),
			// ),
			container.NewCenter(
				widget.NewTextGridFromString("TRH copyright"),
			),
			container.NewPadded(
				widget.NewButton("contnue", func() {
					trhd.setMainPanel(trhd.Keys.Panel)
					splashWin.Close()
				})),
		))
		splashWin.SetContent(welcomePanel)
		splashWin.Show()
	}
	return nil
}
