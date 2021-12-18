package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/keys"
)

type PanelKeys struct {
	TRHD          *TRHD
	Panel         *fyne.Container
	KeystoreLabel *widget.Label
	QRKey         *canvas.Image
	QRAddress     *canvas.Image
}

func NewPanelKeys() *PanelKeys {
	pk := PanelKeys{}
	return &pk
}

func (pk *PanelKeys) Init(trhd *TRHD) {
	pk.TRHD = trhd
	pk.KeystoreLabel = widget.NewLabel("select a keystore file")
	butOpenFile := widget.NewButtonWithIcon("Open keystore", theme.FolderOpenIcon(), func() {
		fileOpen := dialog.NewFileOpen(pk.keystoreSelected, pk.TRHD.MainWindow)
		fileOpen.Show()
	})
	pk.QRKey = &canvas.Image{FillMode: canvas.ImageFillContain, ScaleMode: canvas.ImageScalePixels}
	pk.QRKey.SetMinSize(fyne.NewSize(300, 300))
	pk.QRAddress = &canvas.Image{FillMode: canvas.ImageFillContain, ScaleMode: canvas.ImageScalePixels}
	pk.QRAddress.SetMinSize(fyne.NewSize(300, 300))
	vertElems := []fyne.CanvasObject{
		container.NewHBox(pk.KeystoreLabel, butOpenFile),
		container.NewHBox(pk.QRKey, pk.QRAddress),
	}
	pk.Panel = container.NewVBox(vertElems...)

}

func (pk *PanelKeys) keystoreSelected(file fyne.URIReadCloser, err error) {
	if err != nil {
		fmt.Printf("error while opening keystore file")
	}
	pk.KeystoreLabel.Text = file.URI().Path()
	pinText := widget.NewEntry()

	pinCh := make(chan string, 1)
	dialPin := dialog.NewForm("Enter PIN", "confirm", "cancel", []*widget.FormItem{widget.NewFormItem("PIN", pinText)}, func(b bool) {
		if b {
			pinCh <- pinText.Text
		} else {
			pinCh <- ""
		}
	}, pk.TRHD.MainWindow)
	dialPin.Show()
	go func() {
		pin := <-pinCh
		if len(pin) == 0 {
			fmt.Printf("No PIN Provided\n")
			return
		}
		pk.setKeystore(file.URI().Path(), pin)

	}()

}

func (pk *PanelKeys) setKeystore(file string, pin string) {
	ks, err := keys.LoadKeystore(file, pin)
	if err != nil {
		fmt.Printf("Cannot load keystore")
	}
	pk.TRHD.TRH.SetKeystore(ks)
	qrImgK, err := ddb.QRCodeImage(ks.Source().Key())
	if err != nil {
		fmt.Printf("Cannot generate Key QRCode")
		dialog.NewError(err, pk.TRHD.MainWindow)
	}
	qrImgA, err := ddb.QRCodeImage(ks.Source().Address())
	if err != nil {
		fmt.Printf("Cannot generate Address QRCode")
		dialog.NewError(err, pk.TRHD.MainWindow)
	}
	pk.QRKey.Image = qrImgK
	pk.QRAddress.Image = qrImgA
	fmt.Printf("Image width: %d\n", qrImgA.Bounds().Dx())
	pk.Panel.Refresh()
}
