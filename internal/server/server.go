package server

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Taboon/urlshortner/internal/config"
	"github.com/Taboon/urlshortner/internal/domain/usecase"
	"github.com/Taboon/urlshortner/internal/logger"
	"github.com/Taboon/urlshortner/internal/server/gzip"
	chi "github.com/go-chi/chi/v5"
)

type Server struct {
	BaseURL      string
	LocalAddress string
	P            usecase.URLProcessor
	Log          *logger.Logger
}

func (s *Server) Run(la config.Address) error {
	srv := &http.Server{
		Addr:         la.String(),
		Handler:      s.URLRouter(), // ваш обработчик запросов
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
		return err
	}
	return nil
}

func (s *Server) URLRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/{id}", s.Log.RequestLogger(gzip.MiddlewareGzip(s.sendURL)))
	r.Get("/ping", s.Log.RequestLogger(s.ping))
	r.Post("/", s.Log.RequestLogger(gzip.MiddlewareGzip(s.getURL)))
	r.Post("/api/shorten", s.Log.RequestLogger(gzip.MiddlewareGzip(s.shortenJSON)))
	r.Post("/api/shorten/batch", s.Log.RequestLogger(gzip.MiddlewareGzip(s.shortenBatchJSON)))

	return r
}
