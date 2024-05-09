package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Taboon/urlshortner/internal/config"
	"github.com/Taboon/urlshortner/internal/domain/usecase"
	"github.com/Taboon/urlshortner/internal/logger"
	"github.com/Taboon/urlshortner/internal/server/auth"
	gzipMW "github.com/Taboon/urlshortner/internal/server/gzip"
	"github.com/Taboon/urlshortner/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func initServer() (Server, error) {
	// инициализируем конфиг
	configBuilder := config.NewConfigBuilder()
	configBuilder.SetLocalAddress("127.0.0.1", 8080)
	configBuilder.SetBaseURL("127.0.0.1", 8080)
	configBuilder.SetFileBase("/tmp/short-url-db.json")
	configBuilder.SetLogger("Debug")
	conf := configBuilder.Build()
	// инициализируем логгер
	l, err := logger.Initialize(*conf)
	if err != nil {
		return Server{}, err
	}
	// инициализируем хранилище
	stor := storage.NewMemoryStorage(l)
	// инициализируем URL процессор
	urlProcessor := usecase.URLProcessor{
		Repo:            stor,
		Log:             l,
		Authentificator: auth.NewAuthentificator(l, stor, conf.BaseURL, "key123"),
	}
	// инициализируем сервер
	s := Server{
		LocalAddress: conf.LocalAddress.String(),
		BaseURL:      conf.BaseURL.String(),
		P:            urlProcessor,
		Log: &logger.Logger{
			Logger: l,
		},
	}

	return s, nil
}

func AddMock(urlMock storage.URLData, s *Server, id int) (*Server, error) {
	ctx := context.WithValue(context.Background(), storage.UserID, id)
	err := s.P.Repo.AddURL(ctx, urlMock)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func TestSendUrl(t *testing.T) {
	urlMock := storage.URLData{
		URL: "http://ya.ru",
		ID:  "AAAAaaaa",
	}

	s, err := initServer()
	require.NoError(t, err, "Error init server")
	cookie, id, err := s.P.Authentificator.SignCookies(context.Background(), nil)
	require.NoError(t, err, "Error set cookies")
	serv, err := AddMock(urlMock, &s, id)
	require.NoError(t, err, "Error add mock")
	s = *serv

	tests := []struct {
		name         string
		method       string
		path         string
		expectedCode int
		expectedURL  string
	}{
		{name: "test1", method: http.MethodGet, path: "/AAAAaaaa", expectedCode: http.StatusOK, expectedURL: urlMock.URL},
		{name: "test2", method: http.MethodGet, path: "/", expectedCode: http.StatusBadRequest, expectedURL: ""},
		{name: "test3", method: http.MethodGet, path: "/aAaaaAAa", expectedCode: http.StatusBadRequest, expectedURL: ""},
	}

	// Создаем тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(s.P.Authentificator.MiddlewareCookies(s.getURL)))
	defer server.Close()

	// Создаем HTTP клиент для выполнения запросов к тестовому серверу
	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return nil
	}}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			// Формируем URL с параметром id
			url := server.URL + tt.path
			fmt.Println(url)
			// Создаем GET запрос
			req, err := http.NewRequest(tt.method, url, nil)
			require.NoError(t, err, "Error new request")
			req.AddCookie(cookie)
			// Выполняем запрос
			resp, err := client.Do(req)
			require.NoError(t, err, "Error do request")
			body, _ := io.ReadAll(resp.Body)
			fmt.Println(string(body))
			fmt.Println(resp)

			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
		})
	}
}

func Test_getUrl(t *testing.T) {

	tests := []struct {
		name         string
		method       string
		body         string
		contentType  string
		expectedCode int
		expectedBody string
	}{
		{name: "test1", method: http.MethodPost, body: "http://ya.ru", contentType: "text/plain", expectedCode: http.StatusCreated},
		{name: "test2", method: http.MethodPost, body: "htt://ya2.ru", contentType: "", expectedCode: http.StatusBadRequest},
		{name: "test3", method: http.MethodPost, body: "http://ya.ru", contentType: "", expectedCode: http.StatusConflict},
	}

	s, err := initServer()
	require.NoError(t, err, "Error init server")
	cookie, _, err := s.P.Authentificator.SignCookies(context.Background(), nil)
	require.NoError(t, err, "Error set cookies")

	server := httptest.NewServer(http.HandlerFunc(s.P.Authentificator.MiddlewareCookies(s.shortURL)))
	defer server.Close()

	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return nil
	}}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			// Формируем URL с параметром id
			url := server.URL

			// Создаем GET запрос
			req, err := http.NewRequest("POST", url, strings.NewReader(tt.body))
			require.NoError(t, err)
			req.AddCookie(cookie)

			// Выполняем запрос
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
			body, _ := io.ReadAll(resp.Body)
			fmt.Println(string(body))
		})
	}
}

func Test_shortenJSON(t *testing.T) {
	tests := []struct {
		name         string
		request      string
		contentType  string
		expectedCode int
		expectedBody string
	}{
		{name: "test1", request: "{\"url\": \"http://ya.ru\"}", contentType: "application/json", expectedCode: http.StatusCreated},
		{name: "test2", request: "{\"url\": \"ya.ru\"}", contentType: "application/json", expectedCode: http.StatusBadRequest},
		{name: "test3", request: "{RL: \"ya.ru\"}", contentType: "application/json", expectedCode: http.StatusBadRequest},
		{name: "test4", request: "{\"url\": \"\"}", contentType: "application/json", expectedCode: http.StatusBadRequest},
	}

	s, err := initServer()
	require.NoError(t, err, "Error init server")

	server := httptest.NewServer(http.HandlerFunc(s.P.Authentificator.MiddlewareCookies(s.shortenJSON)))
	defer server.Close()

	client := &http.Client{CheckRedirect: func(_ *http.Request, via []*http.Request) error {
		return nil
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := server.URL + "/api/shorten"

			req, err := http.NewRequest("POST", url, strings.NewReader(tt.request))
			fmt.Println(tt.request)

			req.Header.Set("Content-Type", tt.contentType)

			require.NoError(t, err)

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			fmt.Println(string(respBody))

			assert.Equal(t, tt.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
		})
	}
}

func Test_shortenBatchJSON(t *testing.T) {
	regex := regexp.MustCompile(`^(https?|http)://[^\s/$.?#].[^\s]*$`)

	tests := []struct {
		name         string
		request      string
		response     map[string]bool
		contentType  string
		expectedCode int
		expectedBody string
	}{
		{name: "test1", request: "[{\"correlation_id\": \"a\", \"original_url\": \"http://yandex.ru\"},{\"correlation_id\": \"b\",\"original_url\": \"http://ya.ru\"}]", contentType: "application/json", expectedCode: http.StatusCreated, response: map[string]bool{"a": true, "b": true}},
		{name: "test2", request: "[{\"correlation_id\": \"a\", \"original_url\": \"http://otherurl.ru\"},{\"correlation_id\": \"b\", \"original_url\": \"http://ya.ru\"}]", contentType: "application/json", expectedCode: http.StatusCreated, response: map[string]bool{"a": true, "b": false}},
		{name: "test3", request: "[{\"correlation_id\": \"a\", \"original_url\": \"ya.ru\"},{\"correlation_id\": \"b\", \"original_url\": \"https://yandexru\"}]", contentType: "application/json", expectedCode: http.StatusCreated, response: map[string]bool{"a": false, "b": false}},
		{name: "test4", request: "[{ \"https://yandex.ru\"}]", contentType: "application/json", expectedCode: http.StatusBadRequest},
	}

	s, err := initServer()
	require.NoError(t, err, "Error init server")
	cookie, _, err := s.P.Authentificator.SignCookies(context.Background(), nil)
	require.NoError(t, err, "Error set cookies")

	server := httptest.NewServer(http.HandlerFunc(s.P.Authentificator.MiddlewareCookies(s.shortenBatchJSON)))
	defer server.Close()

	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return nil
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := server.URL + "/api/shorten/batch"

			req, err := http.NewRequest("POST", url, strings.NewReader(tt.request))
			fmt.Println(tt.request)

			req.AddCookie(cookie)
			req.Header.Set("Content-Type", tt.contentType)

			require.NoError(t, err)

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			fmt.Println(string(respBody))
			assert.Equal(t, tt.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")

			if tt.expectedCode != http.StatusBadRequest {
				var testedResp []storage.RespBatchURL

				assert.Equal(t, true, json.Valid(respBody))

				err = json.Unmarshal(respBody, &testedResp)
				assert.NoError(t, err)

				for _, v := range testedResp {
					assert.Equal(t, tt.response[v.ID], regex.MatchString(v.URL))
				}
			}
		})
	}
}

func TestGzipCompression(t *testing.T) {
	requestBody := `{"url": "https://ya.ru"}`

	s, err := initServer()
	require.NoError(t, err, "Error init server")

	handler := http.HandlerFunc(gzipMW.MiddlewareGzip(s.P.Authentificator.MiddlewareCookies(s.shortenJSON)))
	srv := httptest.NewServer(handler)
	defer srv.Close()

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Accept-Encoding", "")

		resp, err := http.DefaultClient.Do(r)
		fmt.Println(err)
		require.NoError(t, err)
		fmt.Println(err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		defer resp.Body.Close()

		_, err = io.ReadAll(resp.Body)
		require.NoError(t, err)
		//require.JSONEq(t, successBody, string(b))
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)
		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Accept-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		defer resp.Body.Close()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		_, err = io.ReadAll(zr)
		require.NoError(t, err)
		//TODO URL validate
		//require.JSONEq(t, successBody, string(b))
	})
}
