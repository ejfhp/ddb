package ddb

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"strings"

	"rsc.io/qr"
)

const white = "\033[107m  \033[0m"
const black = "\033[40m  \033[0m"

func PrintQRCode(w io.Writer, text string) error {
	code, err := qr.Encode(text, qr.M)
	if err != nil {
		return fmt.Errorf("error while generating QRCode: %w", err)
	}
	line := make([]string, code.Size+1)
	for x := 0; x <= code.Size; x++ {
		line[x] = white
	}
	writeLine(w, line)
	writeLine(w, line)
	writeLine(w, line)
	for y := 0; y <= code.Size; y++ {
		for x := 0; x <= code.Size; x++ {
			if code.Black(x, y) {
				line[x] = black
			} else {
				line[x] = white
			}
		}
		writeLine(w, line)
	}
	for x := 0; x <= code.Size; x++ {
		line[x] = white
	}
	writeLine(w, line)
	writeLine(w, line)
	writeLine(w, line)
	return nil
}

func writeLine(w io.Writer, line []string) {
	border := []string{white, white, white}
	w.Write([]byte(strings.Join(border, "")))
	w.Write([]byte(strings.Join(line, "")))
	w.Write([]byte(strings.Join(border, "")))
	w.Write([]byte("\n"))

}

//QRCodeImage returns the smallest image of the qrcode, it has then to be scaled
func QRCodeImage(text string) (image.Image, error) {
	code, err := qr.Encode(text, qr.M)
	if err != nil {
		return nil, fmt.Errorf("error while generating QRCode: %w", err)
	}
	size := code.Size
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y <= size; y++ {
		for x := 0; x <= size; x++ {
			if code.Black(x, y) {
				img.Set(x, y, color.Black)
			} else {
				img.Set(x, y, color.White)
			}
		}
	}
	return img, nil
}
