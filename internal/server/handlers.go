package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Taboon/urlshortner/internal/entity"
	"github.com/Taboon/urlshortner/internal/storage"
	"go.uber.org/zap"
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
	s.Log.Debug("Получаем ID из пути", zap.String("path", path))
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
		if errors.Is(err, entity.ErrURLExist) {
			url := fmt.Sprintf("%s%s/%s", httpPrefix, s.BaseURL, id)
			url = strings.TrimSpace(url)
			http.Error(w, url, http.StatusConflict)
			return
		}
		http.Error(w, "Не удалось сохранить URL: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write([]byte(fmt.Sprintf("%s%s/%s", httpPrefix, s.BaseURL, id)))

	if err != nil {
		http.Error(w, "Не удалось записать ответ: "+err.Error(), http.StatusBadRequest)
		return
	}
}

var requestBody = RequestJSON{}
var response = Response{}

func (s *Server) shortenJSON(w http.ResponseWriter, r *http.Request) {
	req, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Не удалось прочитать запрос", http.StatusBadRequest)
		return
	}

	if !json.Valid(req) {
		http.Error(w, "Не валидный JSON", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(req, &requestBody)
	if err != nil {
		http.Error(w, "Не удалось сериализовать JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if requestBody.URL == "" {
		http.Error(w, "Пустой URL", http.StatusBadRequest)
		return
	}

	s.Log.Debug("Пытаемся проверить", zap.String("url", requestBody.URL))

	url, err := s.P.URLValidator(requestBody.URL)
	if err != nil {
		http.Error(w, "Неверный URL: "+err.Error(), http.StatusBadRequest)
		return
	}

	id, err := s.P.URLSaver(url)
	if err != nil {
		if errors.Is(err, entity.ErrURLExist) {
			url := fmt.Sprintf("%s%s/%s", httpPrefix, s.BaseURL, id)
			url = strings.TrimSpace(url)
			http.Error(w, url, http.StatusConflict)
			return
		}
		http.Error(w, "Не удалось сохранить URL: "+err.Error(), http.StatusBadRequest)
		return
	}

	response = Response{Result: fmt.Sprintf("%s%s/%s", httpPrefix, s.BaseURL, id)}

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

func (s *Server) shortenBatchJSON(w http.ResponseWriter, r *http.Request) {
	var reqBatchJSON []storage.ReqBatchJSON

	req, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Не удалось прочитать запрос", http.StatusBadRequest)
		return
	}

	if !json.Valid(req) {
		http.Error(w, "Не валидный JSON", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(req, &reqBatchJSON)
	if err != nil {
		http.Error(w, "Не удалось сериализовать JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(reqBatchJSON) == 0 {
		http.Error(w, "Пустой JSON", http.StatusBadRequest)
		return
	}

	s.Log.Debug("Пытаемся проверить массив ссылок")

	reqBatchJSON = *s.P.BatchURLValidator(&reqBatchJSON)

	urls, err := s.P.BatchURLSaver(&reqBatchJSON)
	if err != nil {
		http.Error(w, "Не удалось сохранить массив URL: "+err.Error(), http.StatusBadRequest)
		return
	}

	var respJSON = storage.RespBatchJSON{}
	var respBathJSON = []storage.RespBatchJSON{}

	for id, v := range urls {
		respJSON.ID = v.ID
		switch {
		case !v.Valid:
			respJSON.URL = "url invalid"
		case v.Exist:
			respJSON.URL = "url exist"
		default:
			respJSON.URL = fmt.Sprintf("%s%s/%s", httpPrefix, s.BaseURL, id)
		}
		respBathJSON = append(respBathJSON, respJSON)
	}

	resp, err := json.Marshal(respBathJSON)
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
