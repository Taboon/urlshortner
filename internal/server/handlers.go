package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Taboon/urlshortner/internal/entity"
	"github.com/Taboon/urlshortner/internal/storage"
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

func (s *Server) getURL(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	s.Log.Debug("Получаем ID из пути", zap.String("path", path))
	path = strings.Trim(path, "/")

	v, err := s.P.Get(r.Context(), path)
	if err != nil {
		http.Error(w, "Не удалось получить URL", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", v.URL)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (s *Server) ping(w http.ResponseWriter, r *http.Request) {
	err := s.P.Repo.Ping(r.Context())
	if err != nil {
		http.Error(w, "Ошибка при подключении к базе данных", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) shortURL(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Second * 1)
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

	w.Header().Set("Content-Type", "text/plain")

	id, err := s.P.SaveURL(r.Context(), url)

	w = s.setHeader(w, err)

	_, err = w.Write([]byte(fmt.Sprintf("%s%s/%s", httpPrefix, s.BaseURL, id)))

	if err != nil {
		http.Error(w, "Не удалось записать ответ: "+err.Error(), http.StatusBadRequest)
		return
	}
}

func (s *Server) shortenJSON(w http.ResponseWriter, r *http.Request) {
	s.Log.Info("shortenJSON")
	// получаем JSON
	requestBody, err := s.getURLJSON(w, r)
	if err != nil {
		http.Error(w, "Ошибка получения JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// валидируем URL
	url, err := s.P.URLValidator(requestBody.URL)
	if err != nil {
		http.Error(w, "Неверный URL: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// сохраняем URL
	id, err := s.P.SaveURL(r.Context(), url)

	response := Response{Result: fmt.Sprintf("%s%s/%s", httpPrefix, s.BaseURL, id)}

	s.writeResponse(s.setHeader(w, err), response)
}

func (s *Server) setHeader(w http.ResponseWriter, err error) http.ResponseWriter {
	switch {
	case errors.Is(err, entity.ErrURLExist):
		w.WriteHeader(http.StatusConflict)
	case err != nil && !errors.Is(err, entity.ErrURLExist):
		http.Error(w, "Не удалось сохранить URL: "+err.Error(), http.StatusBadRequest)
		return w
	default:
		w.WriteHeader(http.StatusCreated)
	}
	return w
}

func (s *Server) getURLJSON(w http.ResponseWriter, r *http.Request) (RequestJSON, error) {
	var requestBody = RequestJSON{}

	req, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Не удалось прочитать запрос", http.StatusBadRequest)
		return RequestJSON{}, nil
	}

	if !json.Valid(req) {
		http.Error(w, "Не валидный JSON", http.StatusBadRequest)
		return RequestJSON{}, nil
	}

	err = json.Unmarshal(req, &requestBody)
	if err != nil {
		http.Error(w, "Не удалось сериализовать JSON: "+err.Error(), http.StatusBadRequest)
		return RequestJSON{}, nil
	}

	if requestBody.URL == "" {
		http.Error(w, "Пустой URL", http.StatusBadRequest)
		return RequestJSON{}, nil
	}
	return requestBody, err
}

func (s *Server) shortenBatchJSON(w http.ResponseWriter, r *http.Request) {
	s.Log.Info("shortenBatchJSON")
	// получаем все url в json
	urls, err := getReqBatchJSON(w, r)
	if err != nil {
		s.Log.Error("Ошибка получения JSON", zap.Error(err))
		http.Error(w, "Не валидный JSON", http.StatusBadRequest)
		return
	}
	if len(*urls) == 0 {
		http.Error(w, "Пустой JSON", http.StatusBadRequest)
		return
	}

	// пытаемся сохранить
	urls, err = s.P.BatchURLSave(r.Context(), urls)
	if err != nil {
		http.Error(w, "Не удалось сохранить массив URL: "+err.Error(), http.StatusBadRequest)
		return
	}

	// пишем результат в JSON
	respBathJSON := s.getRespBatchJSON(urls)

	// отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	s.writeResponse(w, respBathJSON)
}

func (s *Server) getRespBatchJSON(urls *storage.ReqBatchURLs) storage.RespBatchURLs {
	var respURL = storage.RespBatchURL{}
	var respBathJSON = storage.RespBatchURLs{}

	for _, v := range *urls {
		respURL.ID = v.ExternalID

		// в кейсах можно добавить обработку на каждый тип ошибки
		switch {
		case v.Err != nil:
			respURL.URL = v.Err.Error()
		default:
			respURL.URL = fmt.Sprintf("%s%s/%s", httpPrefix, s.BaseURL, v.ID)
		}
		respBathJSON = append(respBathJSON, respURL)
	}
	return respBathJSON
}

func (s *Server) writeResponse(w http.ResponseWriter, r interface{}) {
	resp, err := json.Marshal(r)
	if err != nil {
		http.Error(w, "Не удалось кодировать JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	_, err = w.Write(resp)

	if err != nil {
		http.Error(w, "Не удалось записать ответ: "+err.Error(), http.StatusBadRequest)
		return
	}
}

func getReqBatchJSON(w http.ResponseWriter, r *http.Request) (*storage.ReqBatchURLs, error) {
	var reqBatchJSON storage.ReqBatchURLs

	req, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Не удалось прочитать запрос", http.StatusBadRequest)
		return nil, err
	}

	if !json.Valid(req) {
		http.Error(w, "Не валидный JSON", http.StatusBadRequest)
		return nil, entity.ErrJSONInvalid
	}

	err = json.Unmarshal(req, &reqBatchJSON)
	if err != nil {
		http.Error(w, "Не удалось сериализовать JSON: "+err.Error(), http.StatusBadRequest)
		return nil, err
	}
	return &reqBatchJSON, nil
}

func (s *Server) getUserURLs(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("id").(int)
	urls, err := s.P.Repo.GetURLsByUser(r.Context(), id)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Нет доступных URL: "+err.Error(), http.StatusNoContent)
			return
		}
		http.Error(w, "Не удалось получить все URL пользователя: "+err.Error(), http.StatusBadRequest)
		return
	}
	if len(urls) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		s.writeResponse(w, s.setBaseURL(&urls))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) setBaseURL(ls *storage.UserURLs) *storage.UserURLs {
	for i, v := range *ls {
		(*ls)[i].ID = fmt.Sprintf("%s%s/%s", httpPrefix, s.BaseURL, v.ID)
	}
	return ls
}
