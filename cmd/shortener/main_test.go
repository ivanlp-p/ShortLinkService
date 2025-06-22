package main

import (
	"bytes"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/ivanlp-p/ShortLinkService/cmd/config"
	"github.com/ivanlp-p/ShortLinkService/internal/models"
	"github.com/ivanlp-p/ShortLinkService/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_handler(t *testing.T) {
	conf := config.Init()
	store := storage.NewMapStorage()
	fileStorage := storage.NewFileStorage("/tmp/short-url-db.json", store)

	type want struct {
		contentType string
		statusCode  int
		body        string
	}
	tests := []struct {
		name    string
		request string
		body    string
		want    want
	}{
		{
			name:    "create_short_link_correct",
			request: "/",
			body:    "https://rcimbvs.com/iuymedy",
			want: want{
				contentType: "text/plain",
				statusCode:  201,
				body:        "http://localhost:8080/-8eOIgoJ",
			},
		},
		{
			name:    "body_is_empty",
			request: "/",
			body:    "",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				body:        "Bad Request\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()
			h := handler(fileStorage, conf)
			h(w, request)

			result := w.Result()

			actualHeaderContentType := result.Header.Get("Content-Type")

			if actualHeaderContentType == "" || actualHeaderContentType != tt.want.contentType {
				t.Errorf("Actual Header Content-Type = %v, required Header Content-Type = %v",
					actualHeaderContentType, tt.want.contentType)
			}

			actualStatusCode := result.StatusCode

			if actualStatusCode != tt.want.statusCode {
				t.Errorf("Actual status code = %v, required Status code = %v",
					actualStatusCode, tt.want.statusCode)
			}

			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			actualResponseBody := string(body)

			if actualResponseBody == "" || actualResponseBody != tt.want.body {
				t.Errorf("Actual response body = %v, required response body = %v",
					actualResponseBody, tt.want.body)
			}
		})
	}
}

func Test_handlerGet(t *testing.T) {
	store := storage.NewMapStorage()
	shortLink := models.ShortLink{
		ShortURL:    "-8eOIgoJ",
		OriginalURL: "https://rcimbvs.com/iuymedy",
	}
	store.PutOriginalURL(context.Background(), shortLink)
	fileStorage := storage.NewFileStorage("/tmp/short-url-db.json", store)

	type want struct {
		location   string
		statusCode int
		body       string
	}
	tests := []struct {
		name     string
		request  string
		id       string
		urlStore map[string]string
		want     want
	}{
		{
			name:    "get_original_link_correct",
			request: "/{id}",
			id:      "-8eOIgoJ",
			urlStore: map[string]string{
				"-8eOIgoJ": "https://rcimbvs.com/iuymedy",
			},
			want: want{
				location:   "https://rcimbvs.com/iuymedy",
				statusCode: 307,
				body:       "",
			},
		},
		{
			name:    "wrong_short_id",
			request: "/",
			id:      "-8eOIg",
			urlStore: map[string]string{
				"-8eOIgoJ": "https://rcimbvs.com/iuymedy",
			},
			want: want{
				location:   "",
				statusCode: 404,
				body:       "Not Found\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.request, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)

			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			h := handlerGet(fileStorage)
			h(w, request)

			result := w.Result()
			err := result.Body.Close()

			if err != nil {
				return
			}
			assert.Equal(t, tt.want.statusCode, result.StatusCode, "In actual result status code not equals required")
			assert.Equal(t, tt.want.location, result.Header.Get("Location"), "Location not correct")
		})
	}
}

func Test_PostShortenRequest(t *testing.T) {
	conf := config.Init()
	store := storage.NewMapStorage()
	fileStorage := storage.NewFileStorage("/tmp/short-url-db.json", store)
	request := "/api/shorten/"

	successBody := `{
   "result": "http://localhost:8080/7CwAhsKq"
}`

	testCases := []struct {
		name         string // добавляем название тестов
		method       string
		body         string // добавляем тело запроса в табличные тесты
		expectedCode int
		expectedBody string
	}{
		{
			name:         "method_get",
			method:       http.MethodGet,
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: "",
		},
		{
			name:         "method_put",
			method:       http.MethodPut,
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: "",
		},
		{
			name:         "method_delete",
			method:       http.MethodDelete,
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: "",
		},
		{
			name:         "method_post_without_body",
			method:       http.MethodPost,
			expectedCode: http.StatusInternalServerError,
			expectedBody: "",
		},
		{
			name:         "method_post_success",
			method:       http.MethodPost,
			body:         `{"url": "https://practicum.yandex.ru"}`,
			expectedCode: http.StatusCreated,
			expectedBody: successBody,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, request, bytes.NewBufferString(tc.body))
			w := httptest.NewRecorder()
			h := PostShortenRequest(fileStorage, conf)
			h(w, request)

			result := w.Result()

			actualStatusCode := result.StatusCode

			if actualStatusCode != tc.expectedCode {
				t.Errorf("Actual status code = %v, required Status code = %v",
					actualStatusCode, tc.expectedCode)
			}

			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			if tc.expectedBody != "" {
				actualResponseBody := string(body)

				if actualResponseBody != tc.expectedBody {
					t.Errorf("Actual response body = %v, required response body = %v",
						actualResponseBody, tc.expectedBody)
				}
			}
		})
	}

}
