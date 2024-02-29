package main

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
)

type Url struct {
	url string
}

var urls = make(map[string]Url)
var slru = make(map[Url]string)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
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

func sendUrl(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path
	path = strings.Trim(path, "/")

	if v, ok := urls[path]; ok {
		w.Header().Set("Location", v.url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func getUrl(w http.ResponseWriter, r *http.Request) {

	req, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Не удалось прочитать запрос", http.StatusBadRequest)
		return
	}

	url, err := urlValidator(string(req))
	if err != nil {
		http.Error(w, "Неверный URL: "+err.Error(), http.StatusBadRequest)
		return
	}

	id, err := urlSaver(url)
	if err != nil {
		http.Error(w, "Не удалось сохранить URL: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte("http://localhost:8080/" + id))

	if err != nil {
		fmt.Println("Ошибка отправки ответа")
		return
	}
}

func urlValidator(url string) (string, error) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		fmt.Println("Это не URL - не указан http:// или https://")
		return "", errors.New("URL должен начинаться с http:// или https://")
	}
	if !strings.Contains(url, ".") {
		fmt.Println("Это не урл3")
		return "", errors.New("is not url")
	}

	return url, nil
}

func urlSaver(url string) (string, error) {
	urlObj := Url{url}
	if _, ok := slru[urlObj]; ok {
		return "", errors.New("url already exist")
	} else {
		id := generateId()
		urls[id] = urlObj
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
		if _, ok := urls[string(b)]; ok {
			continue
		}
		ok = false
	}
	return string(b)
}
