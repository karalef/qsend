package app

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"qsend/server"
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
	http.ServeFile(w, r, a.filepath)
	a.srv.Shutdown()
}
