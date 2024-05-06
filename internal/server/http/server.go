package http

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"
)

type Server struct {
	srv *http.Server
}

// TODO create configs and create server from cfg

func New() *Server {
	srv := &http.Server{
		Addr:         "localhost:8080",
		ReadTimeout:  time.Second * 4,
		WriteTimeout: time.Second * 4,
		IdleTimeout:  time.Second * 60,
		Handler:      newRouter(),
	}
	return &Server{
		srv: srv,
	}
}

// Creating chi router
func newRouter() http.Handler {
	r := chi.NewRouter()
	// There we need to write endpoints and middlewares

	return r
}
