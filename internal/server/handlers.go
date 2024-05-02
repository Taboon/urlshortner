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
	"sync"
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

	v, err := s.P.Get(r.Context(), path)
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

	doneCh := make(chan struct{})

	inputCh := s.generator(doneCh, requestBody)
	channels := s.fanOut(r.Context(), doneCh, inputCh)
	addResultCh := s.fanIn(doneCh, channels...)
	s.remover(r.Context(), doneCh, addResultCh)

	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) generator(doneCh chan struct{}, input []string) chan string {
	genChan := make(chan string)
	s.Log.Debug("Открыли канал genChan")
	go func() {
		defer func() {
			s.Log.Debug("Закрыли канал genChan")
			close(genChan)
		}()
		fmt.Println(input)

		for _, data := range input {
			select {
			case <-doneCh:
				s.Log.Debug("Получили DONE")
				return
			case genChan <- data:
				s.Log.Debug("Отправили в канал genChan", zap.String("data", data))
			}
		}
	}()

	return genChan
}

func (s *Server) remover(ctx context.Context, doneCh chan struct{}, idToRemove chan storage.URLData) {
	s.Log.Debug("remover начал работу")
	go func() {
		defer func() {
			close(doneCh)
			s.Log.Debug("Закрыли канал DONE")
		}()

		var batch = make([]storage.URLData, 0)

		for data := range idToRemove {
			batch = append(batch, data)
			s.Log.Debug("Добавили в batch", zap.String("id", data.ID))
		}

		s.Log.Debug("Подготовили batch", zap.Any("batch", batch))
		err := s.P.Repo.RemoveURL(ctx, batch)
		if err != nil {
			s.Log.Error("Ошибка удаления URL", zap.Error(err))
		}
	}()
}

func (s *Server) fanOut(ctx context.Context, doneCh chan struct{}, inputCh chan string) []chan storage.URLData {
	numWorkers := 5
	channels := make([]chan storage.URLData, numWorkers)
	for i := 0; i < numWorkers; i++ {
		addResultCh := s.checkID(ctx, doneCh, inputCh)
		channels[i] = addResultCh
	}
	return channels
}

func (s *Server) checkID(ctx context.Context, doneCh chan struct{}, in chan string) chan storage.URLData {
	checkIDout := make(chan storage.URLData)
	go func() {
		defer func() {
			s.Log.Debug("Закрыли канал checkIDout")
			close(checkIDout)
		}()

		for data := range in {
			url, ok, err := s.P.Repo.CheckID(ctx, data)
			if err != nil {
				s.Log.Error("ошибка при проверке ID", zap.Error(err))
			}
			s.Log.Debug("Получили инфу по id", zap.String("id", data), zap.Bool("ok", ok), zap.Any("url", url))
			if ok {
				select {
				case <-doneCh:
					return
				case checkIDout <- url:
					s.Log.Debug("Отправили инфу в канал checkIDout")
				}
			}
		}
	}()
	return checkIDout
}

func (s *Server) fanIn(doneCh chan struct{}, resultChs ...chan storage.URLData) chan storage.URLData {
	finalCh := make(chan storage.URLData)

	var wg sync.WaitGroup

	for _, ch := range resultChs {
		chClosure := ch

		wg.Add(1)
		go func() {
			defer wg.Done()

			for data := range chClosure {
				select {
				case <-doneCh:
					return
				case finalCh <- data:
					s.Log.Debug("Отправляем в канал fanIn", zap.String("id", data.ID))
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(finalCh)
		s.Log.Debug("Закрыли канал finalCh")
	}()

	return finalCh
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
