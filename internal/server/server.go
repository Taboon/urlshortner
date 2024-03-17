package server

import (
	"errors"
	"fmt"
	"github.com/Taboon/urlshortner/internal/config"
	"github.com/Taboon/urlshortner/internal/storage"
	"math/rand"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	Conf config.Config
	Stor storage.Repository
}

const (
	httpPrefix  = "http://"
	httpsPrefix = "https://"
)

func (s *Server) Run() error {
	fmt.Println(s.Conf.URL())
	err := http.ListenAndServe(s.Conf.URL(), s.URLRouter())
	if err != nil {
		return fmt.Errorf("ошибка запуска сервера: %v", err)
	}
	return nil
}

func (s *Server) URLRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/{id}", s.sendURL)
	r.Post("/", s.getURL)
	return r
}

func (s *Server) urlValidator(url string) (string, error) {
	url = strings.TrimSpace(url)
	if !strings.HasPrefix(url, httpPrefix) && !strings.HasPrefix(url, httpsPrefix) {
		fmt.Println("Это не URL - не указан http:// или https://")
		return "", errors.New("URL должен начинаться с http:// или https://")
	}
	if !strings.Contains(url, ".") {
		return "", errors.New("is not url")
	}

	return url, nil
}

func (s *Server) urlSaver(url string) (string, error) {
	if _, ok := s.Stor.CheckURL(url); ok {
		return "", errors.New("url already exist")
	}
	id := s.generateID()
	urlObj := storage.URLData{URL: url, ID: id}
	err := s.Stor.AddURL(urlObj)
	if err != nil {
		return "", err
	}
	return id, nil

}

func (s *Server) generateID() string {

	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 8)
	for {
		for i := range b {
			if rand.Intn(2) == 0 {
				b[i] = letterBytes[rand.Intn(26)] // строчные символы
			} else {
				b[i] = letterBytes[rand.Intn(26)+26] // заглавные символы
			}
		}
		if _, ok := s.Stor.CheckID(string(b)); !ok {
			return string(b)
		}
	}
}
