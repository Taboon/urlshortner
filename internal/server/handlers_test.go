package server

import (
	"encoding/json"
	"fmt"
	"log"
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

func TestServer_shortenJSON(t *testing.T) {
	tests := []struct {
		name         string
		request      Request
		contentType  string
		expectedCode int
		expectedBody string
	}{
		{name: "test1", request: Request{URL: "http://ya.ru"}, contentType: "application/json", expectedCode: http.StatusCreated},
		{name: "test2", request: Request{URL: "ya.ru"}, contentType: "application/json", expectedCode: http.StatusBadRequest},
		{name: "test3", request: Request{URL: ""}, contentType: "application/json", expectedCode: http.StatusBadRequest},
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
			url := server.URL

			body, err := json.Marshal(tt.request)
			if err != nil {
				log.Fatalf("Ошибка маршалинга")
				return
			}

			req, err := http.NewRequest("POST", url, strings.NewReader(string(body)))
			if err != nil {
				t.Fatal(err)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedCode, resp.StatusCode, "Код ответа не совпадает с ожидаемым")
		})
	}
}
