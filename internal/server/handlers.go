package server

import (
	"encoding/json"
	"fmt"
	"github.com/Taboon/urlshortner/internal/logger"
	"io"
	"net/http"
	"strings"
)

type Request struct {
	URL string `json:"url"`
}
type Response struct {
	Result string
}

func (s *Server) sendURL(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path
	path = strings.Trim(path, "/")

	v, ok, err := s.Repo.CheckID(path)
	if err != nil {
		http.Error(w, "Ошибка при проверке ID", http.StatusBadRequest)
	}
	if !ok {
		http.Error(w, "Отсутствует такой ID", http.StatusBadRequest)
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
	fmt.Println(string(req))
	err = json.Unmarshal(req, &body)
	if err != nil {
		http.Error(w, "Не удалось сериализовать JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	url, err := s.urlValidator(body.URL)
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
		http.Error(w, "Не удалось кодировать JSON: "+err.Error(), http.StatusBadRequest)
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
