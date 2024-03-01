package server

import (
	"fmt"
	"github.com/Taboon/urlshortner/cmd/shortener/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

//var urlsTest = make(map[string]Url)
//var slruTest = make(map[Url]string)

func Test_sendUrl(t *testing.T) {
	url := storage.UrlData{
		Url: "http://ya.ru",
		Id:  "AAAAaaaa",
	}
	stor := storage.TempStorage{}
	stor.AddUrl(url)

	tests := []struct {
		name         string
		method       string
		path         string
		expectedCode int
		expectedUrl  string
	}{
		{name: "test1", method: http.MethodGet, path: "/AAAAaaaa", expectedCode: http.StatusTemporaryRedirect, expectedUrl: url.Url},
		{name: "test2", method: http.MethodGet, path: "/", expectedCode: http.StatusBadRequest, expectedUrl: ""},
		{name: "test3", method: http.MethodGet, path: "/aAaaaAAa", expectedCode: http.StatusBadRequest, expectedUrl: ""},
	}
	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			webhook(w, r)

			assert.Equal(t, tt.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			if w.Code == 307 {
				fmt.Println("Сравниваем:" + w.Header().Get("Location") + " и " + tt.expectedUrl)
				require.Equal(t, w.Header().Get("Location"), tt.expectedUrl)
			}
		})
	}
}

func Test_getUrl(t *testing.T) {
	url := storage.UrlData{
		Url: "http://ya.ru",
		Id:  "AAAAaaaa",
	}
	stor := storage.TempStorage{}
	stor.AddUrl(url)

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
	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {

			r := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			webhook(w, r)

			assert.Equal(t, tt.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			if w.Code == http.StatusCreated {
				require.Equal(t, w.Header().Get("Content-Type"), tt.contentType)
			}
		})
	}
}
