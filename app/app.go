package app

import (
	"fmt"
	"net/http"
	"qsend/config"
	"qsend/server"

	"github.com/cheggaaa/pb/v3"
)

type Flags struct {
	config.Config
	ConfigPath string
}

type App struct {
	cfg config.Config

	srv *server.Server

	filepath string
}

func New(name string, cfg config.Config) (*App, error) {
	srv, err := server.New(name)
	if err != nil {
		return nil, err
	}

	return &App{
		cfg: cfg,
		srv: srv,
	}, nil
}

func WriteError(w http.ResponseWriter, code int, err error, wrap string) {
	fmt.Println("Error:", wrap+":", err)
	http.Error(w, err.Error(), http.StatusBadRequest)
}

var pbTemplate pb.ProgressBarTemplate = `{{bar . }} {{percent . }} {{speed . }} {{rtime . "ETA %s"}}`
