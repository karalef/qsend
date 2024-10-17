package app

import (
	"fmt"
	"image"
	"io"
	"log"
	"net/http"
	"qsend/config"
	"qsend/server"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/skip2/go-qrcode"
)

type App struct {
	cfg config.Config

	srv *server.Server

	filepath string
}

func New(name string, cfg config.Config) *App {
	return &App{
		cfg: cfg,
		srv: server.New(name),
	}
}

func WriteError(w http.ResponseWriter, code int, err error, wrap string) {
	fmt.Println("Error:", wrap+":", err)
	http.Error(w, err.Error(), http.StatusBadRequest)
}

var pbTemplate pb.ProgressBarTemplate = `{{bar . }} {{percent . }} {{speed . }} {{rtime . "ETA %s"}}`

func NewPBReader(total int64, reader io.Reader) (*pb.ProgressBar, io.Reader) {
	bar := pb.New64(total).SetTemplate(pbTemplate).SetRefreshRate(time.Millisecond * 100)
	return bar, bar.NewProxyReader(reader)
}

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
