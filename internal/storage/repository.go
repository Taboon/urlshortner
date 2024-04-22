package storage

import (
	"context"
)

type Repository interface {
	// AddURL Возвращает ошибку, если не удалось добавить URL
	AddURL(ctx context.Context, data URLData) error
	// AddBatchURL Возвращает ошибку, если не удалось добавить массив URL
	WriteBatchURL(ctx context.Context, b *ReqBatchURLs) (*ReqBatchURLs, error)
	// CheckID Возвращает \URLData и true, если идентификатор найден, иначе возвращает пустую структуру \URLData и false.
	CheckID(ctx context.Context, id string) (URLData, bool, error)
	// CheckURL Возвращает \URLData и true, если URL найден, иначе возвращает пустую структуру \URLData и false.
	CheckURL(ctx context.Context, url string) (URLData, bool, error)
	// CheckBatchURL Проверяет url на наличие в базе. Если присутствует в базе, то свойство Exist = false
	CheckBatchURL(ctx context.Context, urls *ReqBatchURLs) (*ReqBatchURLs, error)
	// RemoveURL Возвращает ошибку, если не удалось удалить URLData.
	RemoveURL(ctx context.Context, data URLData) error
	// Ping проверяет соединение с БД
	// Возвращает 200 или 500
	Ping() error
}

type URLData struct {
	URL string
	ID  string
}

type ReqBatchURLs []ReqBatchURL

type ReqBatchURL struct {
	ExternalID string `json:"correlation_id"`
	ID         string
	URL        string `json:"original_url"`
	Err        error
}

type RespBatchURLs []RespBatchURL

type RespBatchURL struct {
	ID  string `json:"correlation_id"`
	URL string `json:"short_url"`
}
