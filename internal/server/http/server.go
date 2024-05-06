package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/https-whoyan/dwellingPayload/config"
	"net/http"
)

type Server struct {
	srv *http.Server
}

// TODO create configs and create server from cfg

func New(cfg *config.Server) *Server {
	srv := &http.Server{
		Addr:         cfg.Addr,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
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

	// Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.DefaultLogger)

	return r
}
