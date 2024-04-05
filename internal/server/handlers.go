package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type RequestJSON struct {
	URL string `json:"url"`
}
type Response struct {
	Result string
}

const (
	httpPrefix = "http://"
)

func (s *Server) sendURL(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path
	path = strings.Trim(path, "/")

	v, err := s.P.Get(path)
	if err != nil {
		http.Error(w, "Не удалось получить URL", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", v.URL)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (s *Server) ping(w http.ResponseWriter, r *http.Request) {
	err := s.P.Ping()
	if err != nil {
		http.Error(w, "Ошибка при подключении к базе данных", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) getURL(w http.ResponseWriter, r *http.Request) {

	req, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Не удалось прочитать запрос", http.StatusBadRequest)
		return
	}

	url, err := s.P.URLValidator(string(req))
	if err != nil {
		http.Error(w, "Неверный URL: "+err.Error(), http.StatusBadRequest)
		return
	}

	id, err := s.P.URLSaver(url)
	if err != nil {
		http.Error(w, "Не удалось сохранить URL: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write([]byte(fmt.Sprintf("%s%s/%s", httpPrefix, s.Conf.BaseURL.String(), id)))

	if err != nil {
		http.Error(w, "Не удалось записать ответ: "+err.Error(), http.StatusBadRequest)
		return
	}
}

var requestBody = &RequestJSON{}
var response = &Response{}

func (s *Server) shortenJSON(w http.ResponseWriter, r *http.Request) {
	req, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Не удалось прочитать запрос", http.StatusBadRequest)
		return
	}

	if json.Valid(req) {
		err = json.Unmarshal(req, &requestBody)
		if err != nil {
			http.Error(w, "Не удалось сериализовать JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	url, err := s.P.URLValidator(requestBody.URL)
	if err != nil {
		http.Error(w, "Неверный URL: "+err.Error(), http.StatusBadRequest)
		return
	}

	id, err := s.P.URLSaver(url)
	if err != nil {
		http.Error(w, "Не удалось сохранить URL: "+err.Error(), http.StatusBadRequest)
		return
	}

	response = &Response{Result: fmt.Sprintf("%s%s/%s", httpPrefix, s.Conf.BaseURL.String(), id)}

	resp, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Не удалось кодировать JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)

	if err != nil {
		http.Error(w, "Не удалось записать ответ: "+err.Error(), http.StatusBadRequest)
		return
	}
}
