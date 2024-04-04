package server

import (
	"fmt"
	"net/http"

	"github.com/Taboon/urlshortner/internal/config"
	"github.com/Taboon/urlshortner/internal/domain/usecase"
	"github.com/Taboon/urlshortner/internal/logger"
	"github.com/Taboon/urlshortner/internal/server/gzip"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	Conf *config.Config
	P    usecase.URLProcessor
}

func (s *Server) Run() error {
	err := http.ListenAndServe(s.Conf.URL(), s.URLRouter())
	if err != nil {
		return fmt.Errorf("ошибка запуска сервера: %v", err)
	}
	return nil
}

func (s *Server) URLRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/{id}", logger.RequestLogger(gzip.GzipMiddleware(s.sendURL)))
	r.Post("/", logger.RequestLogger(gzip.GzipMiddleware(s.getURL)))
	r.Post("/api/shorten", logger.RequestLogger(gzip.GzipMiddleware(s.shortenJSON)))
	return r
}
