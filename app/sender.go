package app

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"qsend/server"

	"github.com/cheggaaa/pb/v3"
)

func (a *App) Send(ctx context.Context, filepath string) (string, *server.Server, error) {
	const path = "/receive/"
	a.filepath = filepath
	a.srv.Handle("GET "+path, a.handleReceive)
	err := a.srv.Start(ctx, a.cfg.Addr(), a.cfg.TlsCert, a.cfg.TlsKey)
	if err != nil {
		return "", nil, err
	}
	return a.srv.URL(path), a.srv, nil
}

func (a *App) handleReceive(w http.ResponseWriter, r *http.Request) {
	name := filepath.Base(a.filepath)
	w.Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s",
			name, url.QueryEscape(name)))
	f, err := os.Open(a.filepath)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err, "unable to open file")
		return
	}
	defer f.Close()
	fstat, err := f.Stat()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err, "unable to stat file")
		return
	}
	bar, reader := NewPBReader(fstat.Size(), f)
	http.ServeContent(w, r, name, fstat.ModTime(), &seeker{
		Reader: reader,
		bar:    bar,
		file:   f,
	})
	a.srv.Shutdown()
}

type seeker struct {
	io.Reader
	bar  *pb.ProgressBar
	file *os.File
}

func (s *seeker) Seek(offset int64, whence int) (int64, error) {
	if offset == 0 {
		s.bar.SetCurrent(0)
	}
	return s.file.Seek(offset, whence)
}
