package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"qsend/app/pages"
	"qsend/server"
	"strconv"
	"strings"
)

func (a *App) Receive(ctx context.Context) (string, *server.Server, error) {
	const path = "/send/"
	a.srv.Handle("GET "+path, a.renderSend)
	a.srv.Handle("POST "+path, a.handleSend)
	a.srv.Handle("GET /done/", a.renderDone)
	a.srv.Handle("GET /failed/", a.renderFailed)
	err := a.srv.Start(ctx, a.cfg.Addr(), a.cfg.TlsCert, a.cfg.TlsKey)
	if err != nil {
		return "", nil, err
	}
	return a.srv.URL(path), a.srv, nil
}

func (a *App) handleSend(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	dispos := r.Header.Get("Content-Disposition")
	if dispos == "" {
		WriteError(w, http.StatusBadRequest, errors.New("bad request"), "missing Content-Disposition header")
		return
	}
	_, params, err := mime.ParseMediaType(dispos)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err, "invalid Content-Disposition header")
		return
	}

	xFileName, err := url.QueryUnescape(params["filename"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, err, "invalid filename")
		return
	}

	entries, err := os.ReadDir(a.cfg.Output)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err, "invalid output directory")
		return
	}
	filename := getFileName(filepath.Base(xFileName), entries)
	out, err := os.Create(filepath.Join(a.cfg.Output, filename))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err, "unable to create output file")
		return
	}
	defer out.Close()

	fmt.Println("Transferring file:", out.Name())
	length, err := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err, "invalid content length")
		return
	}
	bar, reader := NewPBReader(length, r.Body)
	bar.Start()
	_, err = io.Copy(out, reader)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err, "unable to transfer file")
		return
	}
	bar.SetCurrent(bar.Total()).Finish()
	w.WriteHeader(http.StatusOK)
}

func (a *App) renderSend(w http.ResponseWriter, r *http.Request) {
	pages.RenderUpload(w, r.URL.Path)
}

func (a *App) renderDone(w http.ResponseWriter, r *http.Request) {
	defer a.srv.Shutdown()
	pages.RenderDone(w, a.cfg.Output)
}

func (a *App) renderFailed(w http.ResponseWriter, r *http.Request) {
	defer a.srv.Shutdown()
	pages.RenderError(w, r.URL.Query().Get("error"))
}

func getFileName(filename string, entries []fs.DirEntry) string {
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)

	number := 1
	for i := 0; i < len(entries); i++ {
		if filename != entries[i].Name() {
			continue
		}
		filename = fmt.Sprintf("%s(%v)%s", name, number, ext)
		number++
		i = 0
	}
	return filename
}
