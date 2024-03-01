package server

import (
	"errors"
	"fmt"
	"github.com/Taboon/urlshortner/cmd/shortener/storage"
	"math/rand"
	"net/http"
	"strings"
)

var stor storage.TempStorage

func Run() error {
	err := http.ListenAndServe(":8080", http.HandlerFunc(webhook))
	if err != nil {
		return fmt.Errorf("ошибка запуска сервера: %v", err)
	}
	return nil
}

func webhook(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		getUrl(w, r)
		return
	case http.MethodGet:
		sendUrl(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func urlValidator(url string) (string, error) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		fmt.Println("Это не URL - не указан http:// или https://")
		return "", errors.New("URL должен начинаться с http:// или https://")
	}
	if !strings.Contains(url, ".") {
		//fmt.Println("Это не урл3")
		return "", errors.New("is not url")
	}

	return url, nil
}

func urlSaver(url string) (string, error) {
	if _, ok := stor.CheckUrl(url); ok {
		return "", errors.New("url already exist")
	} else {
		id := generateId()
		urlObj := storage.UrlData{url, id}
		err := stor.AddUrl(urlObj)
		if err != nil {
			return "", err
		}
		return id, nil
	}
}

func generateId() string {
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
		if _, ok := stor.CheckId(string(b)); ok {
			continue
		}
		ok = false
	}
	return string(b)
}
