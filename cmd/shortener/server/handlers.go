package server

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func sendURL(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path
	path = strings.Trim(path, "/")

	if v, ok := Stor.CheckID(path); ok {
		w.Header().Set("Location", v.URL)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func getURL(w http.ResponseWriter, r *http.Request) {

	req, err := io.ReadAll(r.Body)
	fmt.Println(string(req))
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
