package server

import (
	"encoding/json"
	"github.com/Taboon/urlshortner/internal/logger"
	"io"
	"net/http"
	"strings"
)

type Request struct {
	Url string
}
type Response struct {
	Result string
}

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
		logger.Log.Error("Ошибка отправки")
		return
	}
}

func (s *Server) shortenJSON(w http.ResponseWriter, r *http.Request) {
	req, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Не удалось прочитать запрос", http.StatusBadRequest)
		return
	}

	body := Request{}
	err = json.Unmarshal(req, &body)
	if err != nil {
		http.Error(w, "Не удалось сериализовать JSON", http.StatusBadRequest)
		return
	}
	if body.Url == "" {
		http.Error(w, "Пустой URL", http.StatusBadRequest)
		return
	}
	url, err := s.urlValidator(body.Url)
	if err != nil {
		http.Error(w, "Неверный URL: "+err.Error(), http.StatusBadRequest)
		return
	}
	id, err := s.urlSaver(url)
	if err != nil {
		http.Error(w, "Не удалось сохранить URL: "+err.Error(), http.StatusBadRequest)
		return
	}

	response := Response{Result: httpPrefix + s.Conf.BaseURL.String() + "/" + id}

	resp, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Не кодировать JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)

	if err != nil {
		logger.Log.Error("Ошибка отправки")
		return
	}
}
