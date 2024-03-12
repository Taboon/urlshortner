package server

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (s *Server) sendURL(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path
	path = strings.Trim(path, "/")

	v, ok := s.Stor.CheckID(path)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", v.URL)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusTemporaryRedirect)

}

func (s *Server) getURL(w http.ResponseWriter, r *http.Request) {

	req, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Не удалось прочитать запрос", http.StatusBadRequest)
		return
	}

	url, err := s.urlValidator(string(req))
	if err != nil {
		http.Error(w, "Неверный URL: "+err.Error(), http.StatusBadRequest)
		return
	}

	id, err := s.urlSaver(url)
	if err != nil {
		http.Error(w, "Не удалось сохранить URL: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(httpPrefix + s.Conf.BaseURL.String() + "/" + id))

	if err != nil {
		fmt.Println("Ошибка отправки ответа")
		return
	}
}
