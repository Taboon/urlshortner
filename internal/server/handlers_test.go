package server

import (
	"bytes"
	"compress/gzip"
	"fmt"
	gzipMW "github.com/Taboon/urlshortner/internal/server/gzip"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Taboon/urlshortner/internal/config"
	"github.com/Taboon/urlshortner/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestSendUrl(t *testing.T) {
	urlMock := storage.URLData{
		URL: "http://ya.ru",
		ID:  "AAAAaaaa",
	}

	s := Server{
		Conf: &config.Config{
			LocalAddress: config.Address{
				IP:   "127.0.0.1",
				Port: 8080,
			},
			BaseURL: config.Address{
				IP:   "127.0.0.1",
				Port: 8080,
			},
		},
		Stor: storage.NewTempStorage(),
	}

	err := s.Stor.AddURL(urlMock)
	if err != nil {
		fmt.Println("Error add URL mock")
		return
	}

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
	server := httptest.NewServer(http.HandlerFunc(s.sendURL))
	defer server.Close()

	// Создаем HTTP клиент для выполнения запросов к тестовому серверу
	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return nil
	}}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			// Формируем URL с параметром id
			url := server.URL + tt.path
			// Создаем GET запрос
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				t.Fatal(err)
			}

			// Выполняем запрос
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
		})
	}
}

func Test_getUrl(t *testing.T) {

	urlMock := storage.URLData{
		URL: "http://ya.ru",
		ID:  "AAAAaaaa",
	}

	tests := []struct {
		name         string
		method       string
		body         string
		contentType  string
		expectedCode int
		expectedBody string
	}{
		{name: "test1", method: http.MethodPost, body: "http://ya2.ru", contentType: "text/plain", expectedCode: http.StatusCreated},
		{name: "test2", method: http.MethodPost, body: "htt://ya2.ru", contentType: "", expectedCode: http.StatusBadRequest},
		{name: "test3", method: http.MethodPost, body: "http://ya.ru", contentType: "", expectedCode: http.StatusBadRequest},
	}

	s := Server{
		Conf: &config.Config{
			LocalAddress: config.Address{
				IP:   "127.0.0.1",
				Port: 8080,
			},
			BaseURL: config.Address{
				IP:   "127.0.0.1",
				Port: 8080,
			},
		},
		Stor: storage.NewTempStorage(),
	}

	err := s.Stor.AddURL(urlMock)
	if err != nil {
		fmt.Println("Error add URL mock")
		return
	}

	server := httptest.NewServer(http.HandlerFunc(s.getURL))
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
			if err != nil {
				t.Fatal(err)
			}

			// Выполняем запрос
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			fmt.Println(req)
			fmt.Println(tt.body)
			fmt.Println(tt.name)
			assert.Equal(t, tt.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")

		})
	}
}

func v(t *testing.T) {
	tests := []struct {
		name         string
		request      string
		contentType  string
		expectedCode int
		expectedBody string
	}{
		{name: "test1", request: "{\"url\": \"http://ya.ru\"}", contentType: "application/json", expectedCode: http.StatusCreated},
		{name: "test2", request: "{\"url\": \"ya.ru\"}", contentType: "application/json", expectedCode: http.StatusBadRequest},
		{name: "test2", request: "{RL: \"ya.ru\"}", contentType: "application/json", expectedCode: http.StatusBadRequest},
		{name: "test3", request: "{\"url\": \"\"}", contentType: "application/json", expectedCode: http.StatusBadRequest},
	}

	s := Server{
		Conf: &config.Config{
			LocalAddress: config.Address{
				IP:   "127.0.0.1",
				Port: 8080,
			},
			BaseURL: config.Address{
				IP:   "127.0.0.1",
				Port: 8080,
			},
		},
		Stor: storage.NewTempStorage(),
	}

	server := httptest.NewServer(http.HandlerFunc(s.shortenJSON))
	defer server.Close()

	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return nil
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := server.URL + "/api/shorten"

			req, err := http.NewRequest("POST", url, strings.NewReader(tt.request))
			fmt.Println(tt.request)
			req.Header.Set("Content-Type", tt.contentType)

			if err != nil {
				t.Fatal(err)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println(string(respBody))

			assert.Equal(t, tt.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
		})
	}
}

func TestGzipCompression(t *testing.T) {
	requestBody := `{"url": "https://ya.ru"}`

	s := Server{
		Conf: &config.Config{
			LocalAddress: config.Address{
				IP:   "127.0.0.1",
				Port: 8080,
			},
			BaseURL: config.Address{
				IP:   "127.0.0.1",
				Port: 8080,
			},
		},
		Stor: storage.NewTempStorage(),
	}

	handler := http.HandlerFunc(gzipMW.GzipMiddleware(s.shortenJSON))

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
		require.NoError(t, err)
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
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		defer resp.Body.Close()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		_, err = io.ReadAll(zr)
		require.NoError(t, err)

		//require.JSONEq(t, successBody, string(b))
	})
}
