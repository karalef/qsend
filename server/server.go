package server

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"net"
	"net/http"
	"net/url"
	"sync"
)

// New creates a new server.
func New(appname string) (*Server, error) {
	sessionBytes := make([]byte, 32)
	if _, err := rand.Read(sessionBytes); err != nil {
		panic(err)
	}
	handler := &handler{
		cookie: http.Cookie{
			Name:  appname,
			Value: base64.RawStdEncoding.EncodeToString(sessionBytes),
			Path:  "/",
		},
		mux: http.NewServeMux(),
	}
	return &Server{
		handler: handler,
		http: &http.Server{
			Handler: handler,
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
				CipherSuites: []uint16{
					tls.TLS_AES_128_GCM_SHA256,
					tls.TLS_AES_256_GCM_SHA384,
					tls.TLS_CHACHA20_POLY1305_SHA256,
				},
			},
		},
	}, nil
}

// Server is the http server which is allow only the first client to connect.
type Server struct {
	ctx     context.Context
	cancel  context.CancelFunc
	http    *http.Server
	handler *handler
	closed  chan struct{}

	listener *listener
	secure   bool
}

func (s *Server) Proto() string {
	if s.secure {
		return "https"
	}
	return "http"
}

func (s *Server) Addr() *net.TCPAddr {
	return s.listener.Addr().(*net.TCPAddr)
}

func (s *Server) URL(path string) string {
	return (&url.URL{
		Scheme: s.Proto(),
		Host:   s.Addr().AddrPort().String(),
		Path:   path,
	}).String()
}

func (s *Server) waitContext() {
	<-s.ctx.Done()
	if err := s.http.Shutdown(context.Background()); err != nil {
		panic(err)
	}
	close(s.closed)
}

func (s *Server) serve(cert, key string) {
	var err error
	if s.secure {
		err = s.http.ServeTLS(s.listener, cert, key)
	} else {
		err = s.http.Serve(s.listener)
	}
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

func (s *Server) Start(ctx context.Context, addr string, cert, key string) error {
	s.ctx, s.cancel = context.WithCancel(ctx)
	var err error
	s.listener, err = Listen(addr)
	if err != nil {
		return err
	}

	s.secure = cert != "" && key != ""
	s.closed = make(chan struct{})
	go s.waitContext()
	go s.serve(cert, key)

	return nil
}

func (s *Server) Handle(pattern string, handler http.HandlerFunc) {
	s.handler.mux.Handle(pattern, handler)
}

// Wait waits server shutdown.
func (s *Server) Wait() { <-s.closed }

// Shutdown sends the shutdown signal.
func (s *Server) Shutdown() { s.cancel() }

type handler struct {
	cookie     http.Cookie
	cookieOnce sync.Once
	mux        *http.ServeMux
}

func (h *handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	once := false
	h.cookieOnce.Do(func() {
		http.SetCookie(rw, &h.cookie)
		once = true
	})
	if !once {
		if c, err := r.Cookie(h.cookie.Name); err != nil || c.Value != h.cookie.Value {
			http.Error(rw, "Invalid session ID", http.StatusUnauthorized)
			return
		}
	}
	h.mux.ServeHTTP(rw, r)
}
