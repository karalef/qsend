package app

import (
	"image"
	"log"

	"github.com/skip2/go-qrcode"
)

// QRString returns the QR code as a string.
func QRString(s string, inverseColor bool) string {
	q, err := qrcode.New(s, qrcode.Medium)
	if err != nil {
		log.Fatal(err)
	}
	return q.ToSmallString(inverseColor)
}

// QRImage returns a QR code as an image.Image
func QRImage(s string) image.Image {
	q, err := qrcode.New(s, qrcode.Medium)
	if err != nil {
		log.Fatal(err)
	}
	return q.Image(256)
}
