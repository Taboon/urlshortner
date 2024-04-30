package storage

import (
	"context"
	"go.uber.org/zap"
	"sync"

	"github.com/Taboon/urlshortner/internal/entity"
)

type InternalStorage struct {
	Users    map[int]UserURLs
	Log      *zap.Logger
	mu       sync.Mutex
	Backuper *FileStorage
}

var _ Repository = (*InternalStorage)(nil)

func NewMemoryStorage(logger *zap.Logger) *InternalStorage {
	return &InternalStorage{
		Users: make(map[int]UserURLs),
		Log:   logger,
		mu:    sync.Mutex{},
	}
}

func (is *InternalStorage) Ping(_ context.Context) error {
	return nil
}

func (is *InternalStorage) GetURLsByUser(_ context.Context, id int) (UserURLs, error) {
	return is.Users[id], nil
}

func (is *InternalStorage) GetNewUser(_ context.Context) (int, error) {
	return len(is.Users) + 1, nil
}

func (is *InternalStorage) WriteBatchURL(ctx context.Context, b *ReqBatchURLs) (*ReqBatchURLs, error) {
	urlData := URLData{}
	for i, v := range *b {
		urlData.ID = v.ID
		urlData.URL = v.URL
		err := is.AddURL(ctx, urlData)
		if err != nil {
			(*b)[i].Err = entity.ErrURLExist
		}
	}
	return b, nil
}

func (is *InternalStorage) CheckBatchURL(ctx context.Context, urls *ReqBatchURLs) (*ReqBatchURLs, error) {
	for i, v := range *urls {
		_, ok, err := is.CheckURL(ctx, v.URL)
		if err != nil {
			return nil, err
		}
		if ok {
			(*urls)[i].Err = entity.ErrURLExist
		}
	}
	return urls, nil
}

func (is *InternalStorage) AddURL(ctx context.Context, data URLData) error {
	is.Log.Debug("Сохраняем URL")
	is.mu.Lock()
	defer is.mu.Unlock()

	id := ctx.Value(UserID).(int)

	_, ok := is.Users[id]
	if ok {
		is.Users[id] = append(is.Users[id], data)
	}
	is.Users[id] = append(UserURLs{}, data)

	if is.Backuper != nil {
		is.Log.Debug("Пишем в файл бекапа")
		err := is.Backuper.Set(URLInFile{
			ID:     data.ID,
			URL:    data.URL,
			UserID: id,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (is *InternalStorage) CheckID(_ context.Context, id string) (URLData, bool, error) {
	is.Log.Debug("Проверяем ID")

	for _, u := range is.Users {
		for _, v := range u {
			if v.ID == id {
				return v, true, nil
			}
		}
	}
	return URLData{}, false, nil
}

func (is *InternalStorage) CheckURL(ctx context.Context, url string) (URLData, bool, error) {
	is.Log.Debug("Проверяем URL", zap.String("url", url))
	userid := ctx.Value(UserID).(int)
	user, ok := is.Users[userid]
	if ok {
		for _, v := range user {
			if v.URL == url {
				return v, true, nil
			}
		}
	}
	return URLData{}, false, nil
}

func (is *InternalStorage) RemoveURL(_ context.Context, _ URLData) error {
	return nil
}
