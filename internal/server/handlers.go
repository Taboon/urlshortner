package server

import (
	"context"
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

type RequestJSONRemoveURLs []string

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

	ctx := context.WithValue(r.Context(), storage.UserID, 0)
	v, err := s.P.Get(ctx, path)
	if err != nil {
		http.Error(w, "Не удалось получить URL", http.StatusBadRequest)
		return
	}

	if v.Deleted {
		http.Error(w, "URL удален", http.StatusGone)
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
	reqJSON := RequestJSON{}
	// получаем JSON
	requestBody, err := getURLJSON(w, r, reqJSON)
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

func (s *Server) removeURLs(w http.ResponseWriter, r *http.Request) {
	s.Log.Info("Получили запрос на удаление ссылок")
	var reqJSON []string
	// получаем JSON
	requestBody, err := getURLJSON(w, r, reqJSON)
	if err != nil {
		http.Error(w, "Ошибка получения JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	s.P.RemoveURLs(requestBody, r)

	w.WriteHeader(http.StatusAccepted)
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

func getURLJSON[T any](w http.ResponseWriter, r *http.Request, structure T) (T, error) {
	req, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Не удалось прочитать запрос", http.StatusBadRequest)
		return structure, err
	}

	if !json.Valid(req) {
		http.Error(w, "Не валидный JSON", http.StatusBadRequest)
		return structure, entity.ErrJSONInvalid
	}

	err = json.Unmarshal(req, &structure)
	if err != nil {
		http.Error(w, "Не удалось сериализовать JSON: "+err.Error(), http.StatusBadRequest)
		return structure, err
	}

	return structure, err
}

func (s *Server) shortenBatchJSON(w http.ResponseWriter, r *http.Request) {
	s.Log.Info("shortenBatchJSON")
	// получаем все url в json
	var reqBatchJSON storage.ReqBatchURLs
	urls, err := getURLJSON(w, r, reqBatchJSON)
	if err != nil {
		s.Log.Error("Ошибка получения JSON", zap.Error(err))
		http.Error(w, "Не валидный JSON", http.StatusBadRequest)
		return
	}
	if len(urls) == 0 {
		http.Error(w, "Пустой JSON", http.StatusBadRequest)
		return
	}

	// пытаемся сохранить
	u, err := s.P.BatchURLSave(r.Context(), &urls)
	if err != nil {
		http.Error(w, "Не удалось сохранить массив URL: "+err.Error(), http.StatusBadRequest)
		return
	}

	// пишем результат в JSON
	respBathJSON := s.getRespBatchJSON(u)

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

func (s *Server) getUserURLs(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value(storage.UserID).(int)
	urls, err := s.P.Repo.GetURLsByUser(r.Context(), id)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) { //nolint: typecheck
			http.Error(w, "Нет доступных URL: "+err.Error(), http.StatusUnauthorized)
			return
		}
		http.Error(w, "Не удалось получить все URL пользователя: "+err.Error(), http.StatusBadRequest)
		return
	}
	if len(urls) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Println("что ", urls)
		s.writeResponse(w, s.setBaseURL(&urls))
		return
	}
	w.WriteHeader(http.StatusUnauthorized)
}

func (s *Server) setBaseURL(ls *storage.UserURLs) *storage.UserURLs {
	for i, v := range *ls {
		(*ls)[i].ID = fmt.Sprintf("%s%s/%s", httpPrefix, s.BaseURL, v.ID)
	}
	return ls
}
