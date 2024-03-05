package server

import (
	"errors"
	"fmt"
	"github.com/Taboon/urlshortner/cmd/shortener/config"
	"github.com/Taboon/urlshortner/cmd/shortener/storage"
	"github.com/go-chi/chi/v5"
	"math/rand"
	"net/http"
	"strings"
)

var Stor storage.TempStorage

func Run() error {
	parseFlags()
	err := http.ListenAndServe(config.ConfigGlobal.URL(), URLRouter())
	if err != nil {
		return fmt.Errorf("ошибка запуска сервера: %v", err)
	}
	return nil
}

func URLRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/{id}", sendURL)
	r.Post("/", getURL)
	return r
}

func urlValidator(url string) (string, error) {
	url = strings.TrimSpace(url)
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		fmt.Println("Это не URL - не указан http:// или https://")
		return "", errors.New("URL должен начинаться с http:// или https://")
	}
	if !strings.Contains(url, ".") {
		return "", errors.New("is not url")
	}

	return url, nil
}

func urlSaver(url string) (string, error) {
	if _, ok := Stor.CheckURL(url); ok {
		return "", errors.New("url already exist")
	} else {
		id := generateID()
		urlObj := storage.URLData{URL: url, ID: id}
		err := Stor.AddURL(urlObj)
		if err != nil {
			return "", err
		}
		return id, nil
	}
}

func generateID() string {
	ok := true
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 8)
	for ok {
		for i := range b {
			if rand.Intn(2) == 0 {
				b[i] = letterBytes[rand.Intn(26)] // строчные символы
			} else {
				b[i] = letterBytes[rand.Intn(26)+26] // заглавные символы
			}
		}
		if _, ok := Stor.CheckID(string(b)); ok {
			continue
		}
		ok = false
	}
	return string(b)
}
