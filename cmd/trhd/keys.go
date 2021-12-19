package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"unicode"

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
	butOpenFile := widget.NewButtonWithIcon("Open", theme.FolderOpenIcon(), func() {
		fileOpen := dialog.NewFileOpen(pk.keystoreSelected, pk.TRHD.MainWindow)
		fileOpen.Show()
	})
	butCreateNew := widget.NewButtonWithIcon("New", theme.FolderOpenIcon(), pk.clickButtonNewKeystore)
	pk.QRKey = &canvas.Image{FillMode: canvas.ImageFillContain, ScaleMode: canvas.ImageScalePixels}
	pk.QRKey.SetMinSize(fyne.NewSize(300, 300))
	pk.QRAddress = &canvas.Image{FillMode: canvas.ImageFillContain, ScaleMode: canvas.ImageScalePixels}
	pk.QRAddress.SetMinSize(fyne.NewSize(300, 300))

	vertElems := []fyne.CanvasObject{
		container.NewHBox(butOpenFile, butCreateNew),
		container.NewHBox(pk.KeystoreLabel),
		container.NewDocTabs(
			container.NewTabItem("Main Key", pk.QRKey),
			container.NewTabItem("Main Address", pk.QRAddress),
		),
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
	dialPin := dialog.NewForm("Enter PIN", "OK", "Cancel", []*widget.FormItem{widget.NewFormItem("PIN", pinText)}, func(b bool) {
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

func (pk *PanelKeys) clickButtonNewKeystore() {
	phraseEntry := &widget.Entry{MultiLine: true, Wrapping: fyne.TextWrapBreak}
	keyEntry := &widget.Entry{}
	passEntry := &widget.Entry{}
	pinEntry := &widget.Entry{}
	nameEntry := &widget.Entry{}
	resCh := make(chan []string, 1)
	okPhraseButton := widget.NewButton("Create", func() {
		resCh <- []string{phraseEntry.Text, pinEntry.Text, nameEntry.Text}
	})
	okKeyPButton := widget.NewButton("Create", func() {
		resCh <- []string{keyEntry.Text, passEntry.Text, pinEntry.Text, nameEntry.Text}
	})
	phraseEntry.Validator = func(s string) error {
		if len(s) == 0 {
			return nil
		}
		for _, l := range s {
			if unicode.IsDigit(l) {
				okKeyPButton.Enable()
				return nil
			}
		}
		okKeyPButton.Disable()
		return fmt.Errorf("must contain a number")
	}
	pinEntry.Validator = func(s string) error {
		_, err := strconv.Atoi(s)
		if err != nil {
			okKeyPButton.Disable()
			okPhraseButton.Disable()
		} else {
			okKeyPButton.Enable()
			okPhraseButton.Enable()
		}
		return err
	}
	nameEntry.Validator = func(s string) error {
		for _, l := range s {
			if !unicode.IsLetter(l) && !unicode.IsDigit(l) {
				return fmt.Errorf("only alphanumeric")
			}
		}
		return nil
	}

	tabs := container.NewAppTabs(
		container.NewTabItem("Phrase", container.NewVBox(
			widget.NewForm(
				widget.NewFormItem("Phrase", phraseEntry),
				widget.NewFormItem("PIN", pinEntry),
				widget.NewFormItem("Name", nameEntry),
			),
			okPhraseButton)),
		container.NewTabItem("Key/Password", container.NewVBox(
			widget.NewForm(
				widget.NewFormItem("Key", keyEntry),
				widget.NewFormItem("Password", passEntry),
				widget.NewFormItem("PIN", pinEntry),
				widget.NewFormItem("Name", nameEntry),
			),
			okKeyPButton)),
	)

	dialNewKs := dialog.NewCustom("Enter Phrase or Key/Password", "Cancel", tabs, pk.TRHD.MainWindow)
	dialNewKs.Resize(fyne.NewSize(500, 300))
	dialNewKs.Show()
	go func() {
		res := <-resCh
		close(resCh)
		if len(res) == 0 {
			fmt.Printf("No params provided\n")
			return
		}
		if err := pk.createNewKeystore(res); err != nil {
			fmt.Printf("error creating new keystore: %v", err)
			dialog.NewError(err, pk.TRHD.MainWindow)
		}
		dialNewKs.Hide()
	}()
}

func (pk *PanelKeys) createNewKeystore(init []string) error {
	switch len(init) {
	case 3:
		fmt.Printf("creating keystore from phrase: %s", init[0])
		ksPathName := filepath.Join(getKeystorePath(), init[2]+".trh")
		ks, err := pk.TRHD.TRH.KeystoreGenFromPhrase(init[1], init[0], 3, ksPathName)
		if err != nil {
			return fmt.Errorf("cannot create keystore: %w", err)
		}
		fmt.Printf("create keystore: %s\n", ksPathName)
		pk.TRHD.Keystore = ks
	case 4:
		fmt.Printf("creating keystore from key and password: %s", init[1])
		ksPathName := filepath.Join(getKeystorePath(), init[3]+".trh")
		ks, err := pk.TRHD.TRH.KeystoreGenFromKey(init[2], init[0], init[1], ksPathName)
		if err != nil {
			return fmt.Errorf("cannot create keystore: %w", err)
		}
		fmt.Printf("create keystore: %s\n", ksPathName)
		pk.TRHD.Keystore = ks
	}
	return nil
}

func getKeystorePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error retrieving user dir: %v", err)
		home, err = os.Getwd()
		if err != nil {
			fmt.Printf("Error retrieving working dir: %v", err)
			home = "."
		}
	}
	return home
}
