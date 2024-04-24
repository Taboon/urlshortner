package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/Taboon/urlshortner/internal/entity"
	"go.uber.org/zap"
	"sync"
)

type SafeMap struct {
	mu             sync.Mutex
	mapStor        map[string]string
	reverseMapStor map[string]string
	Log            *zap.Logger
}

func (sm *SafeMap) WriteBatchURL(ctx context.Context, b *ReqBatchURLs) (*ReqBatchURLs, error) {
	urlData := URLData{}
	for i, v := range *b {
		urlData.ID = v.ID
		urlData.URL = v.URL
		err := sm.AddURL(ctx, urlData)
		if err != nil {
			(*b)[i].Err = entity.ErrURLExist
		}
	}
	return b, nil
}

func (sm *SafeMap) CheckBatchURL(ctx context.Context, urls *ReqBatchURLs) (*ReqBatchURLs, error) {
	for i, v := range *urls {
		_, ok, err := sm.CheckURL(ctx, v.URL)
		if err != nil {
			return nil, err
		}
		if ok {
			(*urls)[i].Err = entity.ErrURLExist
		}
	}
	return urls, nil
}

var _ Repository = (*SafeMap)(nil)

func NewMemoryStorage(logger *zap.Logger) *SafeMap {
	return &SafeMap{
		mu:             sync.Mutex{},
		mapStor:        make(map[string]string),
		reverseMapStor: make(map[string]string),
		Log:            logger,
	}
}

func (sm *SafeMap) Ping() error {
	return nil
}

func (sm *SafeMap) AddURL(_ context.Context, data URLData) error {
	sm.Log.Debug("Сохраняем URL")
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Проверяем, что map был инициализирован
	sm.mapChecker()
	fmt.Println(data.URL)
	// Проверяем наличие данных в массиве
	fmt.Println(sm.mapStor)
	_, ok := sm.mapStor[data.ID]
	if ok {
		err := errors.New("id exist")
		return err
	}
	fmt.Println(sm.reverseMapStor)
	_, ok = sm.reverseMapStor[data.URL]
	if ok {
		err := errors.New("url exist")
		return err
	}
	// Пишем данные в map
	sm.mapStor[data.ID] = data.URL
	sm.reverseMapStor[data.URL] = data.ID
	return nil
}

func (sm *SafeMap) CheckID(_ context.Context, id string) (URLData, bool, error) {
	sm.Log.Debug("Проверяем ID")
	urlData := URLData{}
	val, ok := sm.mapStor[id]
	if ok {
		urlData.ID = id
		urlData.URL = val
		return urlData, true, nil
	}
	return urlData, false, nil
}

func (sm *SafeMap) CheckURL(_ context.Context, url string) (URLData, bool, error) {
	sm.Log.Debug("Проверяем URL", zap.String("url", url))
	urlData := URLData{}
	val, ok := sm.reverseMapStor[url]
	if ok {
		urlData.ID = val
		urlData.URL = url
		return urlData, true, nil
	}
	return urlData, false, nil
}

func (sm *SafeMap) RemoveURL(_ context.Context, data URLData) error {
	sm.Log.Debug("Удаляем URL")
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Проверяем, что map был инициализирован
	sm.mapChecker()

	// Удаляем данные из map
	_, ok := sm.mapStor[data.ID]
	if ok {
		_, ok := sm.reverseMapStor[data.URL]
		if ok {
			delete(sm.mapStor, data.ID)
			delete(sm.reverseMapStor, sm.mapStor[data.URL])
			return nil
		}
	}
	err := errors.New("removing entity")
	return err
}

func (sm *SafeMap) mapChecker() {
	if sm.mapStor == nil {
		sm.mapStor = make(map[string]string)
	}
	if sm.reverseMapStor == nil {
		sm.reverseMapStor = make(map[string]string)
	}
}
